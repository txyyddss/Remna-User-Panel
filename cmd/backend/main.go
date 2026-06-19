package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"remna-user-panel/internal/app"
	"remna-user-panel/internal/config"
	"remna-user-panel/internal/logging"
)

func main() {
	settings, err := config.Load()
	if err != nil {
		slog.Error("failed to load settings", "error", err)
		os.Exit(1)
	}
	logging.Configure(settings.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runtime, err := app.NewRuntime(ctx, settings)
	if err != nil {
		slog.Error("failed to initialize backend runtime", "error", err)
		os.Exit(1)
	}
	defer func() { _ = runtime.Close(context.Background()) }()

	if err := runtime.StartBackend(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("backend stopped with error", "error", err)
		os.Exit(1)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := runtime.Close(shutdownCtx); err != nil {
		slog.Warn("backend shutdown finished with warnings", "error", err)
	}
}
