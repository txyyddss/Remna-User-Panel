// Package plugin defines the Go extension contract for Remnawave Minishop.
package plugin

import (
	"context"
	"net/http"
)

// Context is passed to Go plugins during setup.
type Context struct {
	Settings any
	Services map[string]any
}

// WorkerTask describes a long-running worker contributed by a plugin.
type WorkerTask struct {
	Name string
	Run  func(context.Context) error
}

// Plugin is the Go replacement for the old Python entry point plugin API.
type Plugin interface {
	Name() string
	Version() string
	Setup(context.Context, *Context) error
	RegisterHTTP(*http.ServeMux)
	WorkerTasks(*Context) []WorkerTask
}
