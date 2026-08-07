// Package webui embeds the production Vue build in the Go binary.
package webui

import "embed"

// Dist contains Vite output. The placeholder keeps Go tooling usable before the first frontend build.
//
//go:embed all:dist
var Dist embed.FS
