package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

var errDuplicatedUrl = errors.New("duplicated url")

type FSDep struct {
	MkdirAll  func(string, fs.FileMode) error
	WriteFile func(string, []byte, fs.FileMode) error
}

var prod = FSDep{os.MkdirAll, os.WriteFile}

func write(base string, file string, data []any, writeFile func(string, []byte, fs.FileMode) error) error {
	if len(data) == 0 {
		return nil
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return writeBytes(base, file, b, writeFile)
}

func writeBytes(base string, file string, data []byte, writeFile func(string, []byte, fs.FileMode) error) error {
	err := writeFile(filepath.Join(base, file+".json"), data, os.ModePerm)
	return err
}

// getUrlForFragment returns a page url for the first page (from the collection
// "rootPageList" that has url), or empty string is such does not exists
// In other words:
//
//	data["rootPageList"][i].url ; for the smallest applicable i (if any)
//	""                          ; if not found
func getUrlForFragment(data map[string]any) string {
	pages := getArray(data, "rootPageList")

	url := ""
	for _, page := range pages {
		frst, err := castToNode(page)
		if err != nil {
			continue
		}
		url = getStringOrBlank(frst, "url")
		if url != "" {
			break
		}
	}
	return url
}

// checkDistinct checks that all export fragments will have unique urls (thus
// unique folders) so that overwrite doen't happen.
func checkDistinct(exportFragments map[string]any) error {
	urls := map[string]int{}
	for _, fragment := range exportFragments {
		node, err := castToNode(fragment)
		if err != nil {
			return err
		}
		name := getUrlForFragment(node)
		urls[name]++
	}
	for k, v := range urls {
		if v > 1 {
			return fmt.Errorf("url %s - %w", k, errDuplicatedUrl)
		}
	}
	return nil
}

func writeOtherToFS(root string, export map[string]any, writeFile func(string, []byte, fs.FileMode) error) error {
	for element, data := range export {
		if slices.Contains([]string{"rootPageList", "componentList", "slotList", "workflowList"}, element) {
			continue
		}
		array, ok := data.([]any)
		if !ok {
			continue
		}
		if err := write(root, element, array, writeFile); err != nil {
			return err
		}
	}
	return nil
}

func writeToFSInd(root string, exportFragments map[string]any, export map[string]any, mkdirAll func(string, fs.FileMode) error, writeFile func(string, []byte, fs.FileMode) error) error {
	if err := checkDistinct(exportFragments); err != nil {
		return err
	}

	for _, fragment := range exportFragments {
		node, err := castToNode(fragment)
		if err != nil {
			return err
		}
		name := getUrlForFragment(node)

		base := root
		if name != "" {
			base = filepath.Join(root, "pages", name)
		}
		mkdirAll(base, os.ModePerm)

		areas := []struct {
			filename string
			element  string
		}{
			{name, "rootPageList"},
			{"components", "componentList"},
			{"slots", "slotList"},
			{"workflows", "workflowList"},
		}
		for _, area := range areas {
			if err := write(base, area.filename, getArray(node, area.element), writeFile); err != nil {
				return err
			}
		}
	}

	writeOtherToFS(root, export, writeFile)
	return nil
}

func (fs *FSDep) WriteToFS(root string, exportFragments map[string]any, export map[string]any) error {
	return writeToFSInd(root, exportFragments, export, fs.MkdirAll, fs.WriteFile)
}
