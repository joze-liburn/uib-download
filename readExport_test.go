package main

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_toPages(t *testing.T) {
	tests := []struct {
		name   string
		export string
		want   map[string]string
		err    error
	}{
		{
			name:   "pages",
			export: `{ "rootPageList": [ { "id": "1" }, { "id": "2" }, { "id": "3", "parentPageId": "1"} ] }`,
			want: map[string]string{
				"1": `{ "rootPageList": [ { "id": "1" }, { "id": "3", "parentPageId": "1"} ] }`,
				"2": `{ "rootPageList": [ { "id": "2" } ] }`},
		},
		{
			name: "components",
			export: `{ 
				"rootPageList": [ { "id": "1" }, { "id": "2" } ], 
				"componentList": [ {"id": "11", "parentSlotId": "21" } ],
				"slotList": [ {"id": "21", "parentPageId": "1" } ] }`,
			want: map[string]string{
				"1": `{ "rootPageList": [ { "id": "1" } ],
						"componentList": [ {"id": "11", "parentSlotId": "21" } ],
						"slotList": [ {"id": "21", "parentPageId": "1" } ] }`,
				"2": `{ "rootPageList": [ { "id": "2" } ] }`},
		},
		{
			name: "components-chained",
			export: `{ 
				"rootPageList": [ { "id": "1" } ], 
				"componentList": [ {"id": "11", "parentSlotId": "21" }, {"id": "12", "parentSlotId": "22" }, {"id": "13", "parentSlotId": "23" } ],
				"slotList": [ {"id": "21", "parentPageId": "1" }, {"id": "22", "parentComponentId": "11" }, {"id": "23", "parentComponentId": "12" } ] }`,
			want: map[string]string{
				"1": `{ "rootPageList": [ { "id": "1" } ],
						"componentList": [ {"id": "11", "parentSlotId": "21" }, {"id": "12", "parentSlotId": "22" }, {"id": "13", "parentSlotId": "23" } ],
						"slotList": [ {"id": "21", "parentPageId": "1" }, {"id": "22", "parentComponentId": "11" }, {"id": "23", "parentComponentId": "12" } ] }`},
		},
		{
			name: "unexisting-page-id",
			export: `{ 
				"rootPageList":  [ { "id": "1" } ], 
				"componentList": [ { "id": "11", "parentSlotId": "21" } ],
				"slotList":      [ { "id": "21", "parentPageId": "2" } ] 
			}`,
			err: errNotANode,
		},
		{
			name: "broken-chain",
			export: `{ 
				"rootPageList":  [ { "id": "1" } ], 
				"componentList": [ { "id": "11", "parentSlotId": "21" } ],
				"slotList":      [ { "id": "22", "parentPageId": "1" } ] 
			}`,
			want: map[string]string{
				"": `{ "componentList": [ {"id": "11", "parentSlotId": "21" } ]}`,
				"1": `{ "rootPageList": [ { "id": "1" } ],
						"slotList": [ {"id": "22", "parentPageId": "1" } ] }`},
		},
		{
			name: "node-expected",
			export: `{ 
				"rootPageList": [ { "id": "1" }, { "id": "2" } ], 
				"componentList": [ {"id": "11", "parentSlotId": "21" } ],
				"slotList": [ 1, {"id": "21", "parentPageId": "1" } ] }`,
			err: errNotANode,
		},
		{
			name: "list-not-valid",
			export: `{ 
				"rootPageList": [ { "id": "1" }, { "id": "2" } ], 
				"componentList": [ {"id": "11", "parentSlotId": "21" } ],
				"slotList":  1 }`,
			want: map[string]string{
				"":  `{ "componentList": [ {"id": "11", "parentSlotId": "21" } ] }`,
				"1": `{ "rootPageList": [ { "id": "1" } ] }`,
				"2": `{ "rootPageList": [ { "id": "2" } ] }`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			export := make(map[string]any)
			if err := json.Unmarshal([]byte(test.export), &export); err != nil {
				t.Fatalf("%s: 'json' %q does not convert into JSON: %v", test.name, test.export, err)
			}
			want := make(map[string]any)
			for k, v := range test.want {
				var fragment any
				if err := json.Unmarshal([]byte(v), &fragment); err != nil {
					t.Fatalf("%s: 'json' %q does not convert into JSON: %v", test.name, test.export, err)
				}
				want[k] = fragment
			}
			got, err := toPages(export)
			if !errors.Is(err, test.err) {
				t.Errorf("%s: error got %s, want %s", test.name, err, test.err)
			}
			if err != nil {
				return
			}
			if df := cmp.Diff(want, got); df != "" {
				t.Errorf("%s: -want +got:\n %s", test.name, df)
			}
		})
	}
}
