// Package logging configures structured process logging.
package logging

import (
	"log/slog"
	"os"
	"strings"
)

// Configure installs a slog text handler using the requested log level.
func Configure(level string) {
	var slogLevel slog.Level
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "DEBUG":
		slogLevel = slog.LevelDebug
	case "WARN", "WARNING":
		slogLevel = slog.LevelWarn
	case "ERROR":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel})
	slog.SetDefault(slog.New(handler))
}
