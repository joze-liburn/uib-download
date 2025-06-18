package main

import (
	"fmt"
)

func setNonempty(target map[string]any, key string, value []any) {
	if len(value) == 0 {
		return
	}
	target[key] = value
}

func toPages(export map[string]any) (map[string]any, error) {
	bakery := UIBakery{data: export}
	err := mapById(&bakery)
	if err != nil {
		return nil, err
	}

	pagesOnPage := map[string][]any{}
	var pagesQueue []any
	for pagesQueue = getArray(bakery.data, "rootPageList"); len(pagesQueue) > 0; {
		pg, err := castToNode(pagesQueue[0])
		pagesQueue = pagesQueue[1:]
		if err != nil {
			return nil, fmt.Errorf("pages, %w", err)
		}
		topPageId, err := bakery.getpage_page(pg)
		if err != nil {
			return nil, fmt.Errorf("pages, %w", err)
		}
		pagesOnPage[topPageId] = append(pagesOnPage[topPageId], pg)

		if children := getArray(pg, "children"); children != nil {
			pagesQueue = append(pagesQueue, children...)
		}
	}

	slotsOnPage := map[string][]any{}
	for _, slot := range getArray(bakery.data, "slotList") {
		s, err := castToNode(slot)
		if err != nil {
			return nil, fmt.Errorf("slots, %w", err)
		}
		topPageId, err := bakery.getpage_slot(s)
		if err != nil {
			return nil, fmt.Errorf("slots, %w", err)
		}
		slotsOnPage[topPageId] = append(slotsOnPage[topPageId], slot)
	}

	componentsOnPage := map[string][]any{}
	for _, cmp := range getArray(bakery.data, "componentList") {
		c, err := castToNode(cmp)
		if err != nil {
			return nil, fmt.Errorf("component, %w", err)
		}
		topPageId, err := bakery.getpage_component(c)
		if err != nil {
			return nil, fmt.Errorf("component, %w", err)
		}
		componentsOnPage[topPageId] = append(componentsOnPage[topPageId], cmp)
	}

	workflowsOnPage := map[string][]any{}
	for _, wf := range getArray(bakery.data, "workflowList") {
		w, err := castToNode(wf)
		if err != nil {
			return nil, fmt.Errorf("workflow, %w", err)
		}
		topPageId, err := bakery.getpage_workflow(w)
		if err != nil {
			return nil, fmt.Errorf("workflow, %w", err)
		}
		workflowsOnPage[topPageId] = append(workflowsOnPage[topPageId], wf)
	}

	top := map[string]any{}
	for p := range pagesOnPage {
		top[p] = nil
	}
	for p := range componentsOnPage {
		top[p] = nil
	}
	for p := range slotsOnPage {
		top[p] = nil
	}
	for p := range workflowsOnPage {
		top[p] = nil
	}

	for t := range top {
		out := map[string]any{}
		setNonempty(out, "rootPageList", pagesOnPage[t])
		setNonempty(out, "componentList", componentsOnPage[t])
		setNonempty(out, "slotList", slotsOnPage[t])
		setNonempty(out, "workflowList", workflowsOnPage[t])
		top[t] = out
	}

	return top, nil
}
