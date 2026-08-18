// Package admin implements audited, domain-safe administrative operations.
package admin

import (
	"context"
	"errors"
	"fmt"
	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/secret"
	"strings"
)

type SettingDefinition struct {
	Secret   bool
	Required bool
	Validate func(string) error
}

var settingDefinitions = map[string]SettingDefinition{
	"telegram.group_chat_id":                 {Required: true, Validate: validateInteger},
	"telegram.channel_chat_id":               {Required: true, Validate: validateInteger},
	billing.PaymentAnnouncementChatIDSetting: {Validate: validateOptionalInteger},
	"telegram.webhook_secret":                {Secret: true, Validate: validateWebhookSecret},
	"remnawave.base_url":                     {Required: true, Validate: validateHTTPOrHTTPSURL},
	"remnawave.api_token":                    {Secret: true, Required: true, Validate: nonempty},
	"billing.rate.txb_per_cny":               {Required: true, Validate: validatePositiveDecimal},
	"billing.rate.txb_per_usd":               {Required: true, Validate: validatePositiveDecimal},
	"billing.rate.txb_per_xtr":               {Required: true, Validate: validatePositiveDecimal},
	"billing.ezpay.enabled":                  {Validate: validateBoolean},
	"billing.ezpay.base_url":                 {Validate: validateHTTPSURL},
	"billing.ezpay.merchant_id":              {Validate: nonempty},
	"billing.ezpay.key":                      {Secret: true, Validate: nonempty},
	"billing.ezpay.methods":                  {Validate: billing.ValidateEZPayMethods},
	"billing.bepusdt.enabled":                {Validate: validateBoolean},
	"billing.bepusdt.base_url":               {Validate: validateHTTPSURL},
	"billing.bepusdt.api_token":              {Secret: true, Validate: nonempty},
	"billing.bepusdt.methods":                {Validate: billing.ValidateBEPusdtMethods},
	"billing.bepusdt.ack":                    {Validate: validateAck},
	"billing.stars.enabled":                  {Validate: validateBoolean},
	"emby.base_url":                          {Secret: true, Validate: validateHTTPSURL},
	"emby.api_token":                         {Secret: true, Validate: nonempty},
	"emby.setup_price_txb":                   {Validate: validateNonnegativeTXB},
	"activity.timezone":                      {Validate: validateTimezone},
	"activity.daily_reward_min_txb":          {Validate: validateNonnegativeTXB},
	"activity.daily_reward_max_txb":          {Validate: validateNonnegativeTXB},
	"activity.group_message_threshold":       {Validate: validateNonnegativeInteger},
	"activity.group_message_reward_txb":      {Validate: validateNonnegativeTXB},
}

// SettingsRepository stores encrypted or plain values without interpreting them.

type SettingsRepository interface {
	PutSetting(context.Context, string, string, bool, *string) error
	GetSetting(context.Context, string) (database.Setting, error)
	ListSettings(context.Context) ([]database.Setting, error)
}

// SettingsService validates the fixed registry and protects secret values.

type SettingsService struct {
	repository SettingsRepository
	vault      *secret.Vault
	profiles   PaymentProfileRepository
}

// NewSettingsService creates the runtime settings facade.

func NewSettingsService(repository SettingsRepository, vault *secret.Vault) *SettingsService {
	return &SettingsService{repository: repository, vault: vault}
}

// Plaintext returns a setting for trusted server-side use.

func (s *SettingsService) Plaintext(ctx context.Context, key string) (string, error) {
	setting, err := s.repository.GetSetting(ctx, key)
	if err != nil {
		return "", err
	}
	if !setting.Encrypted {
		return setting.Value, nil
	}
	return s.vault.Decrypt(key, setting.Value)
}

// Optional returns an empty string for a missing non-required setting.

func (s *SettingsService) Optional(ctx context.Context, key string) (string, error) {
	value, err := s.Plaintext(ctx, key)
	if errors.Is(err, database.ErrNotFound) {
		return "", nil
	}
	return value, err
}

// Put stores one known setting. An empty secret keeps its existing value.

func (s *SettingsService) Put(ctx context.Context, actorID, key, value string) error {
	definition, ok := settingDefinitions[key]
	if !ok {
		return fmt.Errorf("unknown setting %q", key)
	}
	value = strings.TrimSpace(value)
	if definition.Secret && value == "" {
		if _, err := s.repository.GetSetting(ctx, key); err == nil {
			return nil
		}
	}
	if definition.Validate != nil {
		if err := definition.Validate(value); err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
	}
	if key == "activity.daily_reward_min_txb" || key == "activity.daily_reward_max_txb" {
		if err := s.validateActivityRewardRange(ctx, key, value); err != nil {
			return err
		}
	}
	stored := value
	// A non-blank secret is an explicit replacement. Encrypt it freshly so
	// the write-only value from the settings UI can never be persisted as-is.
	if definition.Secret && value != "" {
		var err error
		stored, err = s.vault.Encrypt(key, value)
		if err != nil {
			return err
		}
	}
	return s.repository.PutSetting(ctx, key, stored, definition.Secret, &actorID)
}

func (s *SettingsService) validateActivityRewardRange(ctx context.Context, key, value string) error {
	proposed, err := billing.ParseTXBMajor(value)
	if err != nil {
		return fmt.Errorf("%s: invalid TXB amount", key)
	}
	otherKey := "activity.daily_reward_max_txb"
	if key == otherKey {
		otherKey = "activity.daily_reward_min_txb"
	}
	otherValue, err := s.Optional(ctx, otherKey)
	if err != nil || strings.TrimSpace(otherValue) == "" {
		return err
	}
	other, err := billing.ParseTXBMajor(otherValue)
	if err != nil {
		return fmt.Errorf("%s: stored reward boundary is invalid", otherKey)
	}
	minimum, maximum := proposed, other
	if key == "activity.daily_reward_max_txb" {
		minimum, maximum = other, proposed
	}
	if minimum > maximum {
		return errors.New("daily reward minimum cannot exceed maximum")
	}
	return nil
}

// SafeList exposes only known keys and masks credentials as write-only.
