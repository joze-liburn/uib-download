package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_getUrl(t *testing.T) {
	tests := []struct {
		name     string
		fragment string
		want     string
	}{
		{
			name:     "normal",
			fragment: `{ "rootPageList" : [ { "id":1, "url": "have one" } ] }`,
			want:     "have one",
		},
		{
			name:     "in-second",
			fragment: `{ "rootPageList" : [ { "id":1 }, { "id": 2, "url": "have one" } ] }`,
			want:     "have one",
		},
		{
			name:     "nowhere",
			fragment: `{ "rootPageList" : [ { "id":1 }, { "id": 2 } ] }`,
			want:     "",
		},
		{
			name:     "no-root-pages",
			fragment: `{ "something" : [ { "id":1 }, { "id": 2 } ] }`,
			want:     "",
		},
		{
			name:     "no-data",
			fragment: `{}`,
			want:     "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			frag := make(map[string]any)
			if err := json.Unmarshal([]byte(test.fragment), &frag); err != nil {
				t.Fatalf("%s: 'json' %q does not convert into JSON: %v", test.name, test.fragment, err)
			}
			got := getUrlForFragment(frag)
			if got != test.want {
				t.Errorf("%s: got %q, want %q", test.name, got, test.want)
			}
		})
	}
}

func Test_checkDistinct(t *testing.T) {
	tests := []struct {
		name      string
		fragments map[string]string
		want      error
	}{
		{
			name: "three-distinct",
			fragments: map[string]string{
				"fragment-1": `{ "rootPageList": [ { "id":1, "url": "en" }     ] }`,
				"fragment-2": `{ "rootPageList": [ { "id":2, "url": "dva" }    ] }`,
				"fragment-3": `{ "rootPageList": [ { "id":3, "url": "trije" }  ] }`,
			},
		},
		{
			name: "one-duplicated",
			fragments: map[string]string{
				"fragment-1": `{ "rootPageList": [ { "id":1, "url": "en" }     ] }`,
				"fragment-2": `{ "rootPageList": [ { "id":2, "url": "dva" }    ] }`,
				"fragment-3": `{ "rootPageList": [ { "id":3, "url": "en" }     ] }`,
				"fragment-4": `{ "rootPageList": [ { "id":4, "url": "štirje" } ] }`,
			},
			want: errDuplicatedUrl,
		},
		{
			name: "four-distinct-with-blank",
			fragments: map[string]string{
				"fragment-1": `{ "rootPageList": [ { "id":1, "url": "en" }     ] }`,
				"fragment-2": `{ "rootPageList": [ { "id":2, "url": "dva" }    ] }`,
				"fragment-3": `{ "rootPageList": [ { "id":3 }                  ] }`,
				"fragment-4": `{ "rootPageList": [ { "id":4, "url": "štirje" } ] }`,
			},
		},
		{
			name: "duplicate-blank",
			fragments: map[string]string{
				"fragment-1": `{ "rootPageList": [ { "id":1 }                  ] }`,
				"fragment-2": `{ "rootPageList": [ { "id":2, "url": "dva" }    ] }`,
				"fragment-3": `{ "rootPageList": [ { "id":3 }                  ] }`,
				"fragment-4": `{ "rootPageList": [ { "id":4, "url": "štirje" } ] }`,
			},
			want: errDuplicatedUrl,
		},
		{
			name: "no-data",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			frags := make(map[string]any)
			for k, f := range test.fragments {
				conv := make(map[string]any)
				if err := json.Unmarshal([]byte(f), &conv); err != nil {
					t.Fatalf("%s: 'json' %q does not convert into JSON: %v", test.name, test.fragments[k], err)
				}
				frags[k] = conv
			}
			got := checkDistinct(frags)
			if !errors.Is(got, test.want) {
				t.Errorf("%s: error got %s, want %s", test.name, got, test.want)
			}
		})
	}
}

func Test_writeOtherToFS(t *testing.T) {
	tests := []struct {
		name   string
		root   string
		export string
		want   map[string]int
		err    error
	}{
		{
			name:   "one-other",
			root:   "/some/folder",
			export: `{ "list" : [ { "id":1 } ] }`,
			want:   map[string]int{"/some/folder/list.json": 1},
		},
		{
			name:   "no-others",
			root:   "/some/folder",
			export: `{ "rootPageList" : [ { "id":1 }, { "id": 2 } ], "slotList": [ { "id":21 }, { "id": 22 } ] }`,
			want:   map[string]int{},
		},
		{
			name:   "both",
			root:   "/some/folder",
			export: `{ "rootPageList" : [ { "id":1 }, { "id": 2 } ], "someList": [ { "id":101 }, { "id": 102 } ] }`,
			want:   map[string]int{"/some/folder/someList.json": 1},
		},
		{
			name:   "no-root-pages",
			root:   "/some/folder",
			export: `{ "someList": [ { "id":101 }, { "id": 102 } ], "rootPageList" : [ { "id":1 }, { "id": 2 } ], "another": [ { "id":101 }, { "id": 102 } ] }`,
			want:   map[string]int{"/some/folder/someList.json": 1, "/some/folder/another.json": 1},
		},
		{
			name:   "no-data",
			root:   "/some/folder",
			export: `{}`,
			want:   map[string]int{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsys := make(map[string]int)
			export := make(map[string]any)
			if err := json.Unmarshal([]byte(test.export), &export); err != nil {
				t.Fatalf("%s: 'json' %q does not convert into JSON: %v", test.name, test.export, err)
			}
			err := writeOtherToFS(test.root, export, func(name string, data []byte, mode fs.FileMode) error { fsys[name] += 1; return nil })
			if !errors.Is(err, test.err) {
				t.Errorf("%s: error got %s, want %s", test.name, err, test.err)
			}
			if df := cmp.Diff(test.want, fsys); df != "" {
				t.Errorf("%s: -want +got %s", test.name, df)
			}
		})
	}
}
