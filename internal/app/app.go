// Package app wires the single-container TX Carpool application.
package app

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/accounts"
	"github.com/txyyddss/Remna-User-Panel/internal/activity"
	"github.com/txyyddss/Remna-User-Panel/internal/admin"
	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/catalog"
	"github.com/txyyddss/Remna-User-Panel/internal/coupons"
	"github.com/txyyddss/Remna-User-Panel/internal/emby"
	"github.com/txyyddss/Remna-User-Panel/internal/entitlements"
	"github.com/txyyddss/Remna-User-Panel/internal/httpapi"
	"github.com/txyyddss/Remna-User-Panel/internal/integrations/telegram"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	jobpayload "github.com/txyyddss/Remna-User-Panel/internal/outbox"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/backup"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/config"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/databaseadmin"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/outbox"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/secret"
	"github.com/txyyddss/Remna-User-Panel/internal/questionnaires"
	"github.com/txyyddss/Remna-User-Panel/internal/rollover"
	"github.com/txyyddss/Remna-User-Panel/internal/webui"
)

// Application owns all context-managed process resources.
type Application struct {
	config     config.Config
	logger     *slog.Logger
	httpServer *http.Server
	store      *database.Store
	outbox     *outbox.Worker
	backups    *backup.Service
	telegram   *telegram.Client
	settings   *admin.SettingsService
	billing    *billing.Service
	upstreams  *providerQueues
}

// New opens persistence, constructs integrations, and builds the HTTP router.
func New(cfg config.Config, logger *slog.Logger) (*Application, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	migrationVersions, err := database.MigrationVersions()
	if err != nil {
		return nil, err
	}
	if _, restoreErr := backup.ApplyPendingRestore(ctx, cfg.DatabasePath, migrationVersions); restoreErr != nil {
		if _, statErr := os.Stat(cfg.DatabasePath); statErr != nil {
			return nil, fmt.Errorf("apply pending database restore: %w", restoreErr)
		}
		logger.Error("pending database restore rolled back or was rejected", "error", restoreErr)
	}
	db, err := database.Open(ctx, cfg.DatabasePath)
	if err != nil {
		return nil, err
	}
	cleanup := func(err error) (*Application, error) {
		_ = db.Close()
		return nil, err
	}
	store := database.NewStore(db)
	if _, err := backup.RecordStartupRestore(ctx, db, cfg.DatabasePath); err != nil {
		return cleanup(fmt.Errorf("record startup database restore: %w", err))
	}
	vault, err := secret.NewVault(cfg.MasterKey)
	if err != nil {
		return cleanup(err)
	}
	settings := admin.NewSettingsService(store, vault)
	if err := ensureBootstrapSettings(ctx, store, vault); err != nil {
		return cleanup(err)
	}
	upstreams, err := newProviderQueues()
	if err != nil {
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
	remna := newRemnaAdapter(settings, upstreams.remnawave)
	telegramBridge := telegramAdapter{client: telegramClient}
	paymentBridge := paymentAdapter{settings: settings, telegram: telegramClient, users: store}
	backupService := backup.NewService(db, store, filepath.Join(cfg.DataDir, "backups"), cfg.BackupRetention)
	databaseEditor := databaseadmin.NewService(db, backupService, vault, logger)
	databaseAdminHTTP := httpapi.NewDatabaseAdministrationHTTP(databaseEditor, backupService, migrationVersions)
	accountsService := accounts.NewService(store, initDataAdapter{verifier: verifier}, telegramBridge, remna, settings, cfg.AdminTelegramIDs, cfg.SessionTTL)
	catalogService := catalog.NewService(store, remna, 2*time.Minute)
	billingService := billing.NewService(store, settings, paymentBridge, cfg.PublicBaseURL)
	activityService := activity.NewService(store, activity.CryptoRandom{}, nil)
	couponService := coupons.NewService(store, nil)
	questionnaireService := questionnaires.NewService(store, questionnaires.CryptoCodeGenerator{}, nil)
	embyPrice := embySetupPrice(settings)
	embyService := emby.NewService(store, newEmbyAdapter(settings, upstreams.emby), embyPrice, emby.NewSecretBox(vault))
	adminService := admin.NewService(store, settings, remna, backupService, paymentBridge)
	entitlementWorker := entitlements.NewWorker(store, remna)
	rolloverWorker := rollover.NewService(store, remna)
	outboxWorker := outbox.NewWorker(store)
	for _, kind := range []string{"remna_apply_entitlement", "remna_sync_user"} {
		if err := outboxWorker.Register(kind, entitlementWorker); err != nil {
			return cleanup(fmt.Errorf("register %s outbox handler: %w", kind, err))
		}
	}
	if err := outboxWorker.Register("rollover_finalize", outbox.HandlerFunc(rolloverWorker.HandleOutbox)); err != nil {
		return cleanup(fmt.Errorf("register rollover outbox handler: %w", err))
	}
	if err := outboxWorker.Register(emby.ProvisionOutboxKind, outbox.HandlerFunc(func(ctx context.Context, job model.OutboxJob) error {
		accountID, err := jobpayload.TargetID(job, "accountId")
		if err != nil {
			return err
		}
		return embyService.HandleProvisionJob(ctx, accountID)
	})); err != nil {
		return cleanup(fmt.Errorf("register Emby outbox handler: %w", err))
	}
	if err := outboxWorker.Register("questionnaire_settlement", outbox.HandlerFunc(func(ctx context.Context, job model.OutboxJob) error {
		importID, err := jobpayload.TargetID(job, "importId")
		if err != nil {
			return err
		}
		_, settleErr := store.SettleQuestionnaireImport(ctx, importID, time.Now().UTC())
		return settleErr
	})); err != nil {
		return cleanup(fmt.Errorf("register questionnaire outbox handler: %w", err))
	}
	static, err := fs.Sub(webui.Dist, "dist")
	if err != nil {
		return cleanup(fmt.Errorf("open embedded frontend: %w", err))
	}
	api, err := httpapi.New(httpapi.Dependencies{
		Accounts: accountsService, Catalog: catalogService, Billing: billingService, Activity: activityService,
		Coupons: couponService, Questionnaires: questionnaireService, Emby: embyService, EmbyPrice: embyPrice,
		Admin: adminService, Settings: settings, DatabaseAdmin: databaseAdminHTTP,
		Store: store, Telegram: telegramClient, Webhooks: paymentBridge, PublicURL: cfg.PublicBaseURL, Static: static,
		Logger: logger, SessionTTL: cfg.SessionTTL, SecureCookies: cfg.PublicBaseURL.Scheme == "https",
		AdminTelegramIDs: cfg.AdminTelegramIDs, RequestSigningKey: cfg.MasterKey,
	})
	if err != nil {
		return cleanup(err)
	}
	httpServer := &http.Server{
		Addr: ":" + cfg.Port, Handler: api.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20,
	}
	return &Application{
		config: cfg, logger: logger, httpServer: httpServer, store: store, outbox: outboxWorker,
		backups: backupService, telegram: telegramClient, settings: settings, billing: billingService, upstreams: upstreams,
	}, nil
}
