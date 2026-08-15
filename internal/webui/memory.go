package webui

import (
	"io/fs"
	"testing/fstest"
)

// Preload copies an embedded asset tree into a private, read-only filesystem.
// The returned fs.FS never references the source tree and has no mutating API,
// so request handling cannot grow or alter the startup snapshot.
func Preload(source fs.FS) (fs.FS, error) {
	files := make(fstest.MapFS)
	err := fs.WalkDir(source, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(source, path)
		if err != nil {
			return err
		}
		files[path] = &fstest.MapFile{Data: append([]byte(nil), data...), Mode: entry.Type()}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return immutableFS{files: files}, nil
}

type immutableFS struct {
	files fstest.MapFS
}

func (f immutableFS) Open(name string) (fs.File, error) {
	return f.files.Open(name)
}
