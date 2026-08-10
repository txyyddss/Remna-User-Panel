// Package admin implements audited, domain-safe administrative operations.
package admin

import (
	"context"
	"errors"
	"fmt"
	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/secret"
	"sort"
	"strings"
)

// SettingDefinition constrains dashboard-editable configuration.
type SettingDefinition struct {
	Secret   bool
	Required bool
	Validate func(string) error
}

var settingDefinitions = map[string]SettingDefinition{
	"telegram.group_chat_id":            {Required: true, Validate: validateInteger},
	"telegram.channel_chat_id":          {Required: true, Validate: validateInteger},
	"telegram.webhook_secret":           {Secret: true, Validate: validateWebhookSecret},
	"remnawave.base_url":                {Required: true, Validate: validateHTTPSURL},
	"remnawave.api_token":               {Secret: true, Required: true, Validate: nonempty},
	"billing.rate.txb_per_cny":          {Required: true, Validate: validatePositiveDecimal},
	"billing.rate.txb_per_usd":          {Required: true, Validate: validatePositiveDecimal},
	"billing.rate.txb_per_xtr":          {Required: true, Validate: validatePositiveDecimal},
	"billing.ezpay.enabled":             {Validate: validateBoolean},
	"billing.ezpay.base_url":            {Validate: validateHTTPSURL},
	"billing.ezpay.merchant_id":         {Validate: nonempty},
	"billing.ezpay.key":                 {Secret: true, Validate: nonempty},
	"billing.ezpay.methods":             {Validate: billing.ValidateEZPayMethods},
	"billing.bepusdt.enabled":           {Validate: validateBoolean},
	"billing.bepusdt.base_url":          {Validate: validateHTTPSURL},
	"billing.bepusdt.api_token":         {Secret: true, Validate: nonempty},
	"billing.bepusdt.methods":           {Validate: billing.ValidateBEPusdtMethods},
	"billing.bepusdt.ack":               {Validate: validateAck},
	"billing.stars.enabled":             {Validate: validateBoolean},
	"emby.base_url":                     {Secret: true, Validate: validateHTTPSURL},
	"emby.api_token":                    {Secret: true, Validate: nonempty},
	"emby.setup_price_txb":              {Validate: validateNonnegativeTXB},
	"activity.timezone":                 {Validate: validateTimezone},
	"activity.daily_reward_min_txb":     {Validate: validateNonnegativeTXB},
	"activity.daily_reward_max_txb":     {Validate: validateNonnegativeTXB},
	"activity.group_message_threshold":  {Validate: validateNonnegativeInteger},
	"activity.group_message_reward_txb": {Validate: validateNonnegativeTXB},
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
func (s *SettingsService) SafeList(ctx context.Context) ([]model.Setting, error) {
	stored, err := s.repository.ListSettings(ctx)
	if err != nil {
		return nil, err
	}
	byKey := make(map[string]database.Setting, len(stored))
	for _, setting := range stored {
		byKey[setting.Key] = setting
	}
	keys := make([]string, 0, len(settingDefinitions))
	for key := range settingDefinitions {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]model.Setting, 0, len(keys))
	for _, key := range keys {
		definition := settingDefinitions[key]
		setting, exists := byKey[key]
		value := setting.Value
		if definition.Secret {
			value = ""
		}
		result = append(result, model.Setting{Key: key, Value: value, Encrypted: definition.Secret, Configured: exists && setting.Value != "", Category: settingCategory(key), UpdatedAt: setting.UpdatedAt})
	}
	return result, nil
}

func settingCategory(key string) string {
	switch {
	case strings.HasPrefix(key, "telegram."):
		return "telegram"
	case strings.HasPrefix(key, "remnawave."):
		return "remnawave"
	case strings.HasPrefix(key, "billing.rate."):
		return "rates"
	case strings.HasPrefix(key, "billing."):
		return "payments"
	case strings.HasPrefix(key, "emby."):
		return "emby"
	case strings.HasPrefix(key, "activity."):
		return "activity"
	default:
		return "application"
	}
}

// Readiness returns stable setup issue identifiers for operations and the admin UI.
func (s *SettingsService) Readiness(ctx context.Context, activeComboCount int) []string {
	issues := make([]string, 0)
	for key, definition := range settingDefinitions {
		if !definition.Required {
			continue
		}
		value, err := s.Optional(ctx, key)
		if err != nil || value == "" || (definition.Validate != nil && definition.Validate(value) != nil) {
			issues = append(issues, "missing_or_invalid:"+key)
		}
	}
	for _, provider := range []string{"ezpay", "bepusdt"} {
		enabled, _ := s.Optional(ctx, "billing."+provider+".enabled")
		if enabled != "true" {
			continue
		}
		keys := []string{"base_url"}
		if provider == "ezpay" {
			keys = append(keys, "merchant_id", "key", "methods")
		} else {
			keys = append(keys, "api_token", "methods")
		}
		for _, suffix := range keys {
			if value, err := s.Optional(ctx, "billing."+provider+"."+suffix); err != nil || value == "" {
				issues = append(issues, "missing:billing."+provider+"."+suffix)
			}
		}
	}
	if activeComboCount == 0 {
		issues = append(issues, "missing:catalog.active_combo")
	}
	sort.Strings(issues)
	return issues
}
