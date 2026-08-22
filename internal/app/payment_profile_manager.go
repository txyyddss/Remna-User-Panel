package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/admin"
	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/integrations/bepusdt"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/upstreamqueue"
	"github.com/txyyddss/Remna-User-Panel/internal/telegramformat"
)

// PaymentProfileManager probes BEPUSDT profiles and publishes process-local rails.
type PaymentProfileManager struct {
	settings *admin.SettingsService
	cache    *billing.PaymentChannelCache
	queue    *upstreamqueue.Queue
	telegram *queuedTelegram
	admins   []int64
	public   *url.URL
	logger   *slog.Logger
}

func newPaymentProfileManager(settings *admin.SettingsService, cache *billing.PaymentChannelCache,
	queue *upstreamqueue.Queue, telegram *queuedTelegram, admins []int64, public *url.URL, logger *slog.Logger) *PaymentProfileManager {
	return &PaymentProfileManager{settings: settings, cache: cache, queue: queue, telegram: telegram,
		admins: append([]int64(nil), admins...), public: public, logger: logger}
}

// RefreshAll probes every enabled BEPUSDT profile during startup.
func (m *PaymentProfileManager) RefreshAll(ctx context.Context) error {
	profiles, err := m.settings.PaymentProfiles(ctx)
	if err != nil {
		return err
	}
	var refreshErrors []error
	for _, profile := range profiles {
		if profile.Provider != "bepusdt" {
			continue
		}
		if _, refreshErr := m.RefreshPaymentProfile(ctx, profile.ID); refreshErr != nil {
			refreshErrors = append(refreshErrors, refreshErr)
		}
	}
	return errors.Join(refreshErrors...)
}

// RefreshPaymentProfile replaces one profile's discovered channels.
func (m *PaymentProfileManager) RefreshPaymentProfile(ctx context.Context, id string) (model.PaymentProfile, error) {
	runtime, err := m.settings.PaymentProfileByID(ctx, id, "")
	if err != nil {
		return model.PaymentProfile{}, err
	}
	if runtime.Provider != "bepusdt" || !runtime.Enabled {
		m.cache.Remove(id)
		return runtime.PaymentProfile, nil
	}
	channels, err := m.discover(ctx, runtime)
	if err != nil {
		return m.disableFailed(ctx, runtime.PaymentProfile, err)
	}
	m.cache.Replace(id, channels)
	runtime.EnabledChannels = m.cache.PaymentProfileRails(id)
	return runtime.PaymentProfile, nil
}

// DeletePaymentProfile removes an unused profile and its process-local rails.
func (m *PaymentProfileManager) DeletePaymentProfile(ctx context.Context, id string) error {
	if err := m.settings.DeletePaymentProfile(ctx, id); err != nil {
		return err
	}
	m.cache.Remove(id)
	return nil
}

func (m *PaymentProfileManager) discover(ctx context.Context, profile model.PaymentProfileRuntime) ([]model.PaymentChannel, error) {
	client, err := bepusdt.NewClient(profile.Endpoint, profile.CredentialPlaintext)
	if err != nil {
		return nil, err
	}
	probeID, err := ids.New()
	if err != nil {
		return nil, err
	}
	probeCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	methods, err := upstreamqueue.Do(probeCtx, m.queue, func(callCtx context.Context) ([]bepusdt.AvailableMethod, error) {
		return client.DiscoverMethods(callCtx, bepusdt.DiscoveryRequest{OrderID: "profile-probe-" + probeID,
			NotifyURL: m.absolute("/api/v1/webhooks/bepusdt/probe"), RedirectURL: m.absolute("/")})
	})
	if err != nil {
		return nil, err
	}
	return discoveredPaymentChannels(methods)
}

func (m *PaymentProfileManager) absolute(path string) string {
	result := *m.public
	result.Path = strings.TrimRight(result.Path, "/") + path
	return result.String()
}

func (m *PaymentProfileManager) disableFailed(ctx context.Context, profile model.PaymentProfile, cause error) (model.PaymentProfile, error) {
	m.cache.Remove(profile.ID)
	disabled, disableErr := m.settings.DisablePaymentProfile(ctx, profile.ID)
	notifyErr := m.notifyFailure(ctx, profile, cause)
	return disabled, errors.Join(fmt.Errorf("discover BEPUSDT profile %q: %w", profile.ID, cause), disableErr, notifyErr)
}

func (m *PaymentProfileManager) notifyFailure(ctx context.Context, profile model.PaymentProfile, cause error) error {
	body := "*BEPUSDT payment profile disabled*\nProfile: " + telegramformat.Escape(profile.ProviderName) +
		"\nReason: " + telegramformat.Escape(cause.Error())
	var sendErrors []error
	for _, adminID := range m.admins {
		if err := m.telegram.SendMarkdownV2Message(ctx, adminID, 0, telegramformat.Limit(body)); err != nil {
			sendErrors = append(sendErrors, err)
			m.logger.Error("notify administrator about BEPUSDT profile failure", "telegram_id", adminID, "profile_id", profile.ID, "error", err)
		}
	}
	return errors.Join(sendErrors...)
}

func discoveredPaymentChannels(methods []bepusdt.AvailableMethod) ([]model.PaymentChannel, error) {
	byRail := make(map[string]model.PaymentChannel, len(methods))
	for _, method := range methods {
		rail, err := method.TradeType()
		if err != nil {
			continue
		}
		byRail[rail] = model.PaymentChannel{Rail: rail, Currency: strings.ToUpper(strings.TrimSpace(method.Currency)),
			Network: strings.ToLower(strings.TrimSpace(method.Network)), NetworkName: strings.TrimSpace(method.NetworkName)}
	}
	channels := make([]model.PaymentChannel, 0, len(byRail))
	for _, rail := range billing.PaymentChannels("bepusdt") {
		if channel, ok := byRail[rail]; ok {
			channels = append(channels, channel)
		}
	}
	if len(channels) == 0 {
		return nil, errors.New("BEPUSDT returned no supported USDT or USDC channels")
	}
	return channels, nil
}
