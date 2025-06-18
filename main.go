package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func fileFlag(full string) (dir string, base string, file string) {
	dir, file = filepath.Split(full)
	base = strings.TrimSuffix(file, filepath.Ext(file))
	return
}

func main() {
	filename := flag.String("file", "", "input file name (expected .json)")
	flag.Parse()
	if *filename == "" {
		fmt.Println("Specify filename: -file=name.json")
		return
	}
	_, base, file := fileFlag(*filename)

	exportJson, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("Reading error: %v\n", err)
		return
	}
	var export map[string]any
	err = json.Unmarshal(exportJson, &export)
	if err != nil {
		fmt.Printf("Json error: %v\n", err)
		return
	}

	pgs, err := toPages(export)
	if err != nil {
		fmt.Printf("Proceesing error: %v\n", err)
		return
	}

	prod.WriteToFS(base, pgs, export)
}
