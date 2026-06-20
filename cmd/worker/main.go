package main

import (
	"context"
	"log/slog"
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
		slog.Error("failed to initialize worker runtime", "error", err)
		os.Exit(1)
	}

	if err := runtime.StartWorker(ctx); err != nil {
		slog.Error("worker stopped with error", "error", err)
		os.Exit(1)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := runtime.Close(shutdownCtx); err != nil {
		slog.Warn("worker shutdown finished with warnings", "error", err)
	}
}
