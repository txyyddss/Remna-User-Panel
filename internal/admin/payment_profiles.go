package admin

import (
	"context"
	"errors"
	"strings"

	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

// PaymentProfileRepository persists masked provider profiles and encrypted credentials.
type PaymentProfileRepository interface {
	ListPaymentProfiles(context.Context) ([]model.PaymentProfile, error)
	SavePaymentProfile(context.Context, database.PaymentProfileInput) (model.PaymentProfile, error)
	PaymentProfileRecord(context.Context, string, string) (database.PaymentProfileRecord, error)
}

// SetPaymentProfileRepository attaches the durable profile store after bootstrap.
func (s *SettingsService) SetPaymentProfileRepository(repository PaymentProfileRepository) {
	s.profiles = repository
}

func (s *SettingsService) PaymentProfiles(ctx context.Context) ([]model.PaymentProfile, error) {
	if s.profiles == nil {
		return []model.PaymentProfile{}, nil
	}
	return s.profiles.ListPaymentProfiles(ctx)
}

// PaymentProfile returns one decrypted profile for trusted provider adapters.
func (s *SettingsService) PaymentProfile(ctx context.Context, provider, rail string) (model.PaymentProfileRuntime, error) {
	if s.profiles == nil {
		return model.PaymentProfileRuntime{}, database.ErrNotFound
	}
	record, err := s.profiles.PaymentProfileRecord(ctx, provider, rail)
	if err != nil {
		return model.PaymentProfileRuntime{}, err
	}
	if rail != "" && !containsPaymentChannel(record.EnabledChannels, rail) {
		return model.PaymentProfileRuntime{}, database.ErrNotFound
	}
	plaintext, decryptErr := s.vault.Decrypt("payment-profile:"+record.ID, record.CredentialCiphertext)
	if decryptErr != nil {
		legacyKey := "billing." + provider + "."
		if provider == "ezpay" {
			legacyKey += "key"
		} else {
			legacyKey += "api_token"
		}
		plaintext, decryptErr = s.Plaintext(ctx, legacyKey)
		if decryptErr != nil {
			return model.PaymentProfileRuntime{}, decryptErr
		}
	}
	return model.PaymentProfileRuntime{PaymentProfile: record.PaymentProfile, CredentialPlaintext: plaintext}, nil
}

// PaymentProfileRuntimes returns configured profiles for callback verification.
func (s *SettingsService) PaymentProfileRuntimes(ctx context.Context, provider string) ([]model.PaymentProfileRuntime, error) {
	profiles, err := s.PaymentProfiles(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]model.PaymentProfileRuntime, 0, len(profiles))
	for _, profile := range profiles {
		if profile.Provider != provider || !profile.Configured {
			continue
		}
		runtime, runtimeErr := s.PaymentProfile(ctx, profile.Provider, "")
		if runtimeErr != nil {
			continue
		}
		result = append(result, runtime)
	}
	return result, nil
}

// SavePaymentProfile validates and encrypts a complete provider profile. Blank
// credentials deliberately preserve the existing ciphertext.
func (s *SettingsService) SavePaymentProfile(ctx context.Context, actorID string, input model.PaymentProfile) (model.PaymentProfile, error) {
	if s.profiles == nil || (input.Provider != "ezpay" && input.Provider != "bepusdt") {
		return model.PaymentProfile{}, database.ErrConflict
	}
	if err := billing.ValidatePaymentChannels(input.Provider, input.EnabledChannels); err != nil {
		return model.PaymentProfile{}, database.ErrConflict
	}
	if err := validateHTTPSURL(strings.TrimSpace(input.Endpoint)); err != nil {
		return model.PaymentProfile{}, err
	}
	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = input.Provider
	}
	providerName := strings.TrimSpace(input.ProviderName)
	if providerName == "" {
		return model.PaymentProfile{}, errors.New("payment profile provider name is required")
	}
	channels := make([]string, 0, len(input.EnabledChannels))
	for _, channel := range input.EnabledChannels {
		channels = append(channels, strings.ToLower(strings.TrimSpace(channel)))
	}
	ciphertext := ""
	credential := strings.TrimSpace(input.Credential)
	if credential != "" && credential != "********" {
		var err error
		ciphertext, err = s.vault.Encrypt("payment-profile:"+id, credential)
		if err != nil {
			return model.PaymentProfile{}, err
		}
	}
	return s.profiles.SavePaymentProfile(ctx, database.PaymentProfileInput{ID: id, Provider: input.Provider, ProviderName: providerName,
		EnabledChannels: channels, Endpoint: strings.TrimSpace(input.Endpoint), MerchantID: strings.TrimSpace(input.MerchantID),
		CredentialCiphertext: ciphertext, Acknowledgement: strings.TrimSpace(input.Acknowledgement), Enabled: input.Enabled})
}

func containsPaymentChannel(channels []string, wanted string) bool {
	for _, channel := range channels {
		if channel == wanted {
			return true
		}
	}
	return false
}
