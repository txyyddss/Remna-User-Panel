package webui

import (
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestPreloadCopiesTheStartupAssetSet(t *testing.T) {
	source := fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("before")}}
	loaded, err := Preload(source)
	if err != nil {
		t.Fatalf("Preload() error = %v", err)
	}
	source["index.html"].Data = []byte("after")
	data, err := fs.ReadFile(loaded, "index.html")
	if err != nil || string(data) != "before" {
		t.Fatalf("preloaded index = (%q, %v), want before", data, err)
	}
	source["late.js"] = &fstest.MapFile{Data: []byte("late")}
	if _, err := fs.ReadFile(loaded, "late.js"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("late asset error = %v, want fs.ErrNotExist", err)
	}
}
