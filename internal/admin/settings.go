// Package admin implements audited, domain-safe administrative operations.
package admin

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/secret"
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
	"activity.daily_reward_txb":         {Validate: validateNonnegativeTXB},
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
	stored := value
	if definition.Secret {
		var err error
		stored, err = s.vault.Encrypt(key, value)
		if err != nil {
			return err
		}
	}
	return s.repository.PutSetting(ctx, key, stored, definition.Secret, &actorID)
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

func validatePositiveDecimal(value string) error {
	decimal, err := billing.ParseDecimal(value)
	if err != nil || !decimal.Positive() {
		return errors.New("must be a positive fixed decimal")
	}
	return nil
}

func validateNonnegativeTXB(value string) error {
	minor, err := billing.ParseTXBMajor(value)
	if err != nil || minor < 0 {
		return errors.New("must be a non-negative TXB amount with at most two decimals")
	}
	return nil
}

func validateTimezone(value string) error {
	if _, err := time.LoadLocation(strings.TrimSpace(value)); err != nil {
		return errors.New("must be an IANA timezone")
	}
	return nil
}

func validateInteger(value string) error {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed == 0 {
		return errors.New("must be a non-zero integer")
	}
	return nil
}

func validateNonnegativeInteger(value string) error {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 0 {
		return errors.New("must be a non-negative integer")
	}
	return nil
}

func validateHTTPSURL(value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("must be an absolute HTTPS URL without credentials, query, or fragment")
	}
	return nil
}

func validateBoolean(value string) error {
	if value != "true" && value != "false" {
		return errors.New("must be true or false")
	}
	return nil
}

func validateAck(value string) error {
	if value != "ok" && !strings.EqualFold(value, "success") {
		return errors.New("must be ok or success")
	}
	return nil
}

// Kept for compatibility with callers validating one legacy selection. New
// configuration uses the ordered billing.ezpay.methods list.
func validateEZPayType(value string) error {
	return billing.ValidateEZPayMethods(value)
}

var webhookSecretPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,256}$`)

func validateWebhookSecret(value string) error {
	if !webhookSecretPattern.MatchString(value) {
		return errors.New("must contain 1-256 URL-safe characters")
	}
	return nil
}

func nonempty(value string) error {
	if value == "" {
		return errors.New("must not be empty")
	}
	return nil
}
