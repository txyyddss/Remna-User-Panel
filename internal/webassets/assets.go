// Package webassets locates Mini App runtime assets.
package webassets

import (
	"fmt"
	"os"
	"path/filepath"
)

// Paths contains resolved runtime asset directories.
type Paths struct {
	TemplatesDir string
	ThemesDir    string
}

// Resolve returns asset paths relative to the process working directory.
func Resolve() (Paths, error) {
	root := os.Getenv("WEBASSETS_DIR")
	if root == "" {
		root = "internal/webassets"
	}
	paths := Paths{
		TemplatesDir: filepath.Join(root, "templates"),
		ThemesDir:    filepath.Join(root, "themes"),
	}
	if _, err := os.Stat(filepath.Join(paths.TemplatesDir, "subscription_webapp.html")); err != nil {
		return Paths{}, fmt.Errorf("subscription webapp template missing: %w", err)
	}
	return paths, nil
}
