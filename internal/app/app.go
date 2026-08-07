// Package app wires the single-container TX Carpool application.
package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/accounts"
	"github.com/txyyddss/Remna-User-Panel/internal/admin"
	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/catalog"
	"github.com/txyyddss/Remna-User-Panel/internal/entitlements"
	"github.com/txyyddss/Remna-User-Panel/internal/httpapi"
	"github.com/txyyddss/Remna-User-Panel/internal/integrations/telegram"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/backup"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/config"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/secret"
	"github.com/txyyddss/Remna-User-Panel/internal/webui"
)

// Application owns all context-managed process resources.
type Application struct {
	config       config.Config
	logger       *slog.Logger
	httpServer   *http.Server
	store        *database.Store
	entitlements *entitlements.Worker
	backups      *backup.Service
	telegram     *telegram.Client
	settings     *admin.SettingsService
	billing      *billing.Service
}

// New opens persistence, constructs integrations, and builds the HTTP router.
func New(cfg config.Config, logger *slog.Logger) (*Application, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	db, err := database.Open(ctx, cfg.DatabasePath)
	if err != nil {
		return nil, err
	}
	cleanup := func(err error) (*Application, error) {
		_ = db.Close()
		return nil, err
	}
	store := database.NewStore(db)
	vault, err := secret.NewVault(cfg.MasterKey)
	if err != nil {
		return cleanup(err)
	}
	settings := admin.NewSettingsService(store, vault)
	if err := ensureBootstrapSettings(ctx, store, vault); err != nil {
		return cleanup(err)
	}
	verifier, err := telegram.NewInitDataVerifier(cfg.TelegramBotToken, cfg.InitDataMaxAge)
	if err != nil {
		return cleanup(err)
	}
	telegramClient, err := telegram.NewClient(cfg.TelegramBotToken)
	if err != nil {
		return cleanup(err)
	}
	remna := remnaAdapter{settings: settings}
	telegramBridge := telegramAdapter{client: telegramClient}
	paymentBridge := paymentAdapter{settings: settings, telegram: telegramClient, users: store}
	backupService := backup.NewService(db, store, filepath.Join(cfg.DataDir, "backups"), cfg.BackupRetention)
	accountsService := accounts.NewService(store, initDataAdapter{verifier: verifier}, telegramBridge, remna, settings, cfg.AdminTelegramID, cfg.SessionTTL)
	catalogService := catalog.NewService(store, remna, 2*time.Minute)
	billingService := billing.NewService(store, settings, paymentBridge, cfg.PublicBaseURL)
	adminService := admin.NewService(store, settings, remna, backupService, paymentBridge)
	entitlementWorker := entitlements.NewWorker(store, remna)
	static, err := fs.Sub(webui.Dist, "dist")
	if err != nil {
		return cleanup(fmt.Errorf("open embedded frontend: %w", err))
	}
	api, err := httpapi.New(httpapi.Dependencies{
		Accounts: accountsService, Catalog: catalogService, Billing: billingService, Admin: adminService, Settings: settings,
		Store: store, Telegram: telegramClient, Webhooks: paymentBridge, PublicURL: cfg.PublicBaseURL, Static: static,
		Logger: logger, SessionTTL: cfg.SessionTTL, SecureCookies: cfg.PublicBaseURL.Scheme == "https",
		AdminTelegramID: cfg.AdminTelegramID,
	})
	if err != nil {
		return cleanup(err)
	}
	httpServer := &http.Server{
		Addr: ":" + cfg.Port, Handler: api.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20,
	}
	return &Application{config: cfg, logger: logger, httpServer: httpServer, store: store, entitlements: entitlementWorker,
		backups: backupService, telegram: telegramClient, settings: settings, billing: billingService}, nil
}

// Run serves HTTP and the single scheduler until cancellation or a fatal listener error.
func (a *Application) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		a.logger.Info("TX Carpool listening", "address", a.httpServer.Addr)
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	go a.runScheduler(ctx)
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.config.ShutdownTimeout)
		defer cancel()
		return a.httpServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

// Close releases the SQLite pool after the HTTP server has stopped.
func (a *Application) Close() error { return a.store.DB().Close() }

func ensureBootstrapSettings(ctx context.Context, store *database.Store, vault *secret.Vault) error {
	if _, err := store.GetSetting(ctx, "telegram.webhook_secret"); errors.Is(err, database.ErrNotFound) {
		value, tokenErr := ids.Token(32)
		if tokenErr != nil {
			return tokenErr
		}
		encrypted, encryptErr := vault.Encrypt("telegram.webhook_secret", value)
		if encryptErr != nil {
			return encryptErr
		}
		if err := store.PutSetting(ctx, "telegram.webhook_secret", encrypted, true, nil); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	defaults := map[string]string{
		"billing.ezpay.enabled": "false", "billing.ezpay.payment_type": "alipay",
		"billing.bepusdt.enabled": "false", "billing.bepusdt.ack": "ok", "billing.stars.enabled": "true",
	}
	for key, value := range defaults {
		if _, err := store.GetSetting(ctx, key); errors.Is(err, database.ErrNotFound) {
			if err := store.PutSetting(ctx, key, value, false, nil); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (a *Application) runScheduler(ctx context.Context) {
	outboxTicker := time.NewTicker(5 * time.Second)
	transitionTicker := time.NewTicker(30 * time.Second)
	maintenanceTicker := time.NewTicker(10 * time.Minute)
	starsTicker := time.NewTicker(5 * time.Minute)
	backupTimer := time.NewTimer(time.Until(a.nextBackup(time.Now())))
	defer outboxTicker.Stop()
	defer transitionTicker.Stop()
	defer maintenanceTicker.Stop()
	defer starsTicker.Stop()
	defer backupTimer.Stop()
	a.configureTelegram(ctx)
	startupNow := time.Now().UTC()
	if err := a.store.RecoverOutbox(ctx, startupNow, startupNow); err != nil {
		a.logger.Error("startup outbox recovery failed", "error", err)
	}
	if err := a.store.EnqueueDueEntitlementTransitions(ctx, startupNow); err != nil {
		a.logger.Error("startup entitlement transition scan failed", "error", err)
	}
	if err := a.entitlements.Drain(ctx, 50); err != nil && !errors.Is(err, context.Canceled) {
		a.logger.Error("startup outbox drain failed", "error", err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-outboxTicker.C:
			if err := a.entitlements.Drain(ctx, 20); err != nil && !errors.Is(err, context.Canceled) {
				a.logger.Error("outbox drain failed", "error", err)
			}
		case now := <-transitionTicker.C:
			if err := a.store.RecoverOutbox(ctx, now.UTC().Add(-2*time.Minute), now.UTC()); err != nil {
				a.logger.Error("outbox lease recovery failed", "error", err)
			}
			if err := a.store.EnqueueDueEntitlementTransitions(ctx, now.UTC()); err != nil {
				a.logger.Error("entitlement transition scan failed", "error", err)
			}
			if err := a.store.ExpireStalePaymentOrders(ctx, now.UTC()); err != nil {
				a.logger.Error("payment expiry scan failed", "error", err)
			}
		case <-maintenanceTicker.C:
			a.configureTelegram(ctx)
			if err := a.store.DeleteExpiredSessions(ctx, time.Now().UTC()); err != nil {
				a.logger.Error("session cleanup failed", "error", err)
			}
			if err := a.backups.RemoveExpired(); err != nil {
				a.logger.Error("backup retention failed", "error", err)
			}
		case <-starsTicker.C:
			a.reconcileStars(ctx)
		case <-backupTimer.C:
			backupCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
			if _, err := a.backups.Run(backupCtx); err != nil && !errors.Is(err, context.Canceled) {
				a.logger.Error("scheduled backup failed", "error", err)
			}
			cancel()
			backupTimer.Reset(time.Until(a.nextBackup(time.Now())))
		}
	}
}

func (a *Application) configureTelegram(ctx context.Context) {
	secret, err := a.settings.Plaintext(ctx, "telegram.webhook_secret")
	if err != nil {
		a.logger.Error("load Telegram webhook secret", "error", err)
		return
	}
	setupCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	webhookURL := *a.config.PublicBaseURL
	webhookURL.Path = strings.TrimRight(webhookURL.Path, "/") + "/api/v1/webhooks/telegram"
	if err := a.telegram.SetWebhook(setupCtx, telegram.WebhookConfig{URL: webhookURL.String(), SecretToken: secret,
		AllowedUpdates: telegram.DefaultAllowedUpdates(), MaxConnections: 20}); err != nil {
		a.logger.Error("configure Telegram webhook", "error", err)
		return
	}
	if err := a.telegram.SetChatMenuButton(setupCtx, "Open TX Carpool", a.config.PublicBaseURL.String()); err != nil {
		a.logger.Error("configure Telegram menu", "error", err)
	}
}

func (a *Application) reconcileStars(ctx context.Context) {
	transactions, err := a.telegram.GetStarTransactions(ctx, 0, 100)
	if err != nil {
		a.logger.Error("Stars reconciliation failed", "error", err)
		return
	}
	for _, transaction := range transactions {
		event, refund, ok := normalizeStarTransaction(transaction)
		if !ok {
			continue
		}
		if !refund {
			if _, _, err := a.billing.Settle(ctx, event); err != nil && !errors.Is(err, database.ErrConflict) && !errors.Is(err, database.ErrNotFound) {
				a.logger.Error("reconcile Stars credit", "transaction_id", transaction.ID, "error", err)
			}
			continue
		}
		if _, err := a.billing.ValidateEvent(ctx, event); err == nil {
			if _, err := a.store.RefundPayment(ctx, nil, event.OrderID, "Telegram Stars reconciliation refund", time.Now().UTC()); err != nil && !errors.Is(err, database.ErrConflict) {
				a.logger.Error("reconcile Stars refund", "transaction_id", transaction.ID, "error", err)
			}
		} else if !errors.Is(err, database.ErrConflict) && !errors.Is(err, database.ErrNotFound) {
			a.logger.Error("reconcile Stars refund", "transaction_id", transaction.ID, "error", err)
		}
	}
}

func normalizeStarTransaction(transaction telegram.StarTransaction) (billing.ProviderEvent, bool, bool) {
	if transaction.NanostarAmount != 0 || transaction.ID == "" {
		return billing.ProviderEvent{}, false, false
	}
	amount := transaction.Amount
	if amount < 0 {
		amount = -amount
	}
	if amount == 0 {
		return billing.ProviderEvent{}, false, false
	}
	partner := transaction.Source
	refund := false
	if partner == nil {
		partner = transaction.Receiver
		refund = partner != nil
	}
	if partner == nil || partner.Type != "user" || partner.TransactionType != "invoice_payment" || partner.InvoicePayload == "" || partner.User.ID <= 0 {
		return billing.ProviderEvent{}, false, false
	}
	telegramID := partner.User.ID
	event := billing.ProviderEvent{Provider: "stars", OrderID: partner.InvoicePayload, TradeID: transaction.ID,
		PayableAmount: strconv.FormatInt(amount, 10), PayableCurrency: "XTR", TelegramID: &telegramID}
	if !refund {
		event.DedupeKey = transaction.ID
	}
	return event, refund, true
}

func (a *Application) nextBackup(now time.Time) time.Time {
	local := now.In(a.config.Timezone)
	next := time.Date(local.Year(), local.Month(), local.Day(), a.config.BackupHour, 0, 0, 0, a.config.Timezone)
	if !next.After(local) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}
