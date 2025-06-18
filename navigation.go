package main

import (
	"errors"
	"fmt"
)

var (
	errMissingProperty = errors.New("missing property")
	errNotANode        = errors.New("node expected")
)

type UIBakery struct {
	data       map[string]any
	pages      map[string]any
	slots      map[string]any
	components map[string]any
	workflows  map[string]any
}

func castToNode(from any) (map[string]any, error) {
	node, ok := from.(map[string]any)
	if !ok {
		return nil, errNotANode
	}
	return node, nil
}

func getString(from map[string]any, field string) (string, error) {
	val, ok := from[field].(string)
	if !ok {
		return "", fmt.Errorf("property %q: %w", field, errMissingProperty)
	}
	return val, nil
}

func getStringOrBlank(from map[string]any, field string) string {
	val, ok := from[field].(string)
	if !ok {
		return ""
	}
	return val
}

func getArray(from map[string]any, node string) []any {
	array, ok := from[node].([]any)
	if !ok {
		return nil
	}
	if len(array) == 0 {
		return nil
	}
	return array
}

func mapInto(items []any, into map[string]any) error {
	for _, item := range items {
		node, err := castToNode(item)
		if err != nil {
			return err
		}

		id, err := getString(node, "id")
		if err != nil {
			return fmt.Errorf("property %q: %w", "id", err)
		}
		into[id] = item
		children := getArray(node, "children")
		if children == nil {
			continue
		}
		mapInto(children, into)
	}
	return nil
}

// page element is either top or owned by anothe "parent" page.
func (b *UIBakery) getpage_page(p map[string]any) (string, error) {
	id, err := getString(p, "id")
	ppid := getStringOrBlank(p, "parentPageId")
	for ppid != "" {
		p, e := castToNode(b.pages[ppid])
		if e != nil {
			break
		}
		id, err = getString(p, "id")
		ppid = getStringOrBlank(p, "parentPageId")
	}
	return id, err
}

// component element is always owned by a slot, or is unclaimed ("actions")
func (b *UIBakery) getpage_component(c map[string]any) (string, error) {
	psid := getStringOrBlank(c, "parentSlotId")
	if psid == "" {
		return "", nil
	}
	slot, err := castToNode(b.slots[psid])
	if err != nil {
		return "", nil
	}
	return b.getpage_slot(slot)

}

// slot element is owned by a component or directly by a page
func (b *UIBakery) getpage_slot(s map[string]any) (string, error) {
	ppid := getStringOrBlank(s, "parentPageId")
	if ppid != "" {
		node, err := castToNode(b.pages[ppid])
		if err != nil {
			return "", errNotANode
		}
		return b.getpage_page(node)
	}

	pcid := getStringOrBlank(s, "parentComponentId")
	if pcid == "" {
		return "", nil
	}
	comp, err := castToNode(b.components[pcid])
	if err != nil {
		return "", errNotANode
	}
	return b.getpage_component(comp)
}

func (b *UIBakery) getpage_workflow(w map[string]any) (string, error) {
	ppid := getStringOrBlank(w, "parentPageId")
	if ppid != "" {
		node, err := castToNode(b.pages[ppid])
		if err != nil {
			return "", errNotANode
		}
		return b.getpage_page(node)
	}

	return getStringOrBlank(w, "parentId"), nil
}

func mapById(b *UIBakery) error {
	b.pages = make(map[string]any)
	b.components = make(map[string]any)
	b.slots = make(map[string]any)
	b.workflows = make(map[string]any)

	if err := mapInto(getArray(b.data, "rootPageList"), b.pages); err != nil {
		return fmt.Errorf("parsing pages: %w", err)
	}

	if err := mapInto(getArray(b.data, "componentList"), b.components); err != nil {
		return fmt.Errorf("parsing components: %w", err)
	}

	if err := mapInto(getArray(b.data, "slotList"), b.slots); err != nil {
		return fmt.Errorf("parsing slots: %w", err)
	}

	if err := mapInto(getArray(b.data, "workflowList"), b.workflows); err != nil {
		return fmt.Errorf("parsing workflows: %w", err)
	}
	return nil
}
