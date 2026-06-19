// Package app wires configuration, storage, HTTP servers, and workers.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"remna-user-panel/internal/config"
	"remna-user-panel/internal/db"
	"remna-user-panel/internal/httpapi"
	"remna-user-panel/internal/i18n"
	"remna-user-panel/internal/payments"
	"remna-user-panel/internal/redisclient"
	"remna-user-panel/internal/remnawave"
	appsettings "remna-user-panel/internal/settings"
	"remna-user-panel/internal/webassets"
	"remna-user-panel/internal/workers"
)

// Runtime owns process resources.
type Runtime struct {
	settings config.Settings
	db       *pgxpool.Pool
	redis    *redis.Client
	i18n     *i18n.Catalog
	assets   webassets.Paths
	payments *payments.Registry
	panel    *remnawave.Client
	servers  []*http.Server
}

// NewRuntime initializes shared runtime dependencies.
func NewRuntime(ctx context.Context, settings config.Settings) (*Runtime, error) {
	pool, err := db.Open(ctx, settings)
	if err != nil {
		return nil, err
	}
	redisClient, err := redisclient.Open(ctx, settings)
	if err != nil {
		pool.Close()
		return nil, err
	}
	catalog, err := i18n.Load("locales", settings.DefaultLanguage)
	if err != nil {
		if redisClient != nil {
			_ = redisClient.Close()
		}
		pool.Close()
		return nil, err
	}
	assets, err := webassets.Resolve()
	if err != nil {
		if redisClient != nil {
			_ = redisClient.Close()
		}
		pool.Close()
		return nil, err
	}
	return &Runtime{
		settings: settings,
		db:       pool,
		redis:    redisClient,
		i18n:     catalog,
		assets:   assets,
		payments: payments.NewRegistry(settings, pool),
		panel:    remnawave.NewClient(settings, appsettings.NewStore(pool)),
	}, nil
}

// StartBackend runs the combined HTTP server (webhooks + Mini App when enabled).
func (r *Runtime) StartBackend(ctx context.Context) error {
	var handler http.Handler
	if r.settings.WebAppEnabled {
		handler = httpapi.CombinedRouter(r.settings, r.db, r.redis, r.i18n, r.assets, r.payments, r.panel)
	} else {
		handler = httpapi.BackendRouter(r.settings, r.db, r.redis, r.payments, r.panel)
	}

	server := &http.Server{
		Addr:              r.settings.WebListenAddr(),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	r.servers = append(r.servers, server)

	errCh := make(chan error, 1)
	go func() {
		slog.Info("HTTP server listening", "addr", server.Addr, "webapp_enabled", r.settings.WebAppEnabled)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return r.Close(context.Background())
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

// StartWorker runs background workers until the context is cancelled.
func (r *Runtime) StartWorker(ctx context.Context) error {
	group := workers.NewGroup()
	group.Add("payment-provisioning", workers.Interval(r.settings.WorkerPaymentProvisionEvery, func(ctx context.Context) error {
		result, err := httpapi.ProvisionPendingPaidOrders(ctx, r.settings, r.db, r.panel, 50)
		if err != nil {
			slog.Warn("payment provisioning worker finished with pending errors", "error", err, "scanned", result.Scanned, "provisioned", result.Provisioned, "failed", result.Failed)
			return nil
		}
		if result.Scanned > 0 {
			slog.Info("payment provisioning worker finished", "scanned", result.Scanned, "provisioned", result.Provisioned, "failed", result.Failed)
		}
		return nil
	}))
	group.Add("panel-sync", workers.Interval(r.settings.WorkerPanelSyncEvery, func(ctx context.Context) error {
		result, err := httpapi.RunPanelSync(ctx, r.settings, r.db, r.panel, 500)
		if err != nil {
			slog.Warn("panel sync worker failed", "error", err, "status", result.Status, "users_processed", result.UsersProcessed, "subscriptions_synced", result.SubscriptionsSynced)
			return nil
		}
		slog.Info("panel sync worker finished", "status", result.Status, "users_processed", result.UsersProcessed, "subscriptions_synced", result.SubscriptionsSynced, "payments_provisioned", result.PaymentsProvisioned, "payments_failed", result.PaymentsFailed)
		return nil
	}))
	group.Add("webhook-queue", workers.Interval(5*time.Second, func(ctx context.Context) error {
		processed, err := httpapi.ProcessQueuedWebhookEvents(ctx, r.db, 100)
		if err != nil {
			slog.Warn("webhook queue worker failed", "error", err)
			return nil
		}
		if processed > 0 {
			slog.Debug("webhook queue worker processed events", "processed", processed)
		}
		return nil
	}))
	group.Add("subscription-notifications", workers.Interval(5*time.Minute, func(ctx context.Context) error {
		notified, err := httpapi.RunSubscriptionNotifications(ctx, r.settings, r.db, r.panel)
		if err != nil {
			slog.Warn("subscription notification worker failed", "error", err)
			return nil
		}
		if notified > 0 {
			slog.Info("subscription notification worker finished", "notified", notified)
		}
		return nil
	}))
	group.Add("telemetry-maintenance", workers.Interval(time.Hour, func(ctx context.Context) error {
		if err := httpapi.RunTelemetryMaintenance(ctx, r.settings, r.db); err != nil {
			slog.Warn("telemetry maintenance failed", "error", err)
		}
		return nil
	}))
	group.Add("data-cleanup", workers.Interval(30*time.Minute, func(ctx context.Context) error {
		if err := httpapi.RunDataCleanup(ctx, r.db); err != nil {
			slog.Warn("data cleanup failed", "error", err)
		}
		return nil
	}))
	return group.Run(ctx)
}

// Close shuts down network servers and storage clients.
func (r *Runtime) Close(ctx context.Context) error {
	var joined error
	for _, server := range r.servers {
		if err := server.Shutdown(ctx); err != nil {
			joined = errors.Join(joined, fmt.Errorf("shutdown %s: %w", server.Addr, err))
		}
	}
	if r.redis != nil {
		if err := r.redis.Close(); err != nil {
			joined = errors.Join(joined, fmt.Errorf("close redis: %w", err))
		}
	}
	if r.db != nil {
		r.db.Close()
	}
	return joined
}
