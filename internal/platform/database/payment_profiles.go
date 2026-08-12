package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

// PaymentProfileInput contains one provider-account profile and its enabled channels.
type PaymentProfileInput struct {
	ID, Provider, ProviderName, Endpoint, MerchantID string
	EnabledChannels                                  []string
	CredentialCiphertext, Acknowledgement            string
	Enabled                                          bool
}

// PaymentProfileRecord is the private storage representation of a provider.
type PaymentProfileRecord struct {
	model.PaymentProfile
	CredentialCiphertext string
}

func (s *Store) SavePaymentProfile(ctx context.Context, input PaymentProfileInput) (model.PaymentProfile, error) {
	if input.ID == "" || input.Provider == "" || input.Endpoint == "" {
		return model.PaymentProfile{}, ErrConflict
	}
	providerName := strings.TrimSpace(input.ProviderName)
	if providerName == "" {
		providerName = input.Provider
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO payment_profiles(id,provider,provider_name,enabled_channels,endpoint,merchant_id,credential_ciphertext,acknowledgement,enabled,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET provider=excluded.provider,provider_name=excluded.provider_name,enabled_channels=excluded.enabled_channels,endpoint=excluded.endpoint,merchant_id=excluded.merchant_id,credential_ciphertext=CASE WHEN excluded.credential_ciphertext='' THEN payment_profiles.credential_ciphertext ELSE excluded.credential_ciphertext END,acknowledgement=excluded.acknowledgement,enabled=excluded.enabled,updated_at=excluded.updated_at`,
		input.ID, input.Provider, providerName, strings.Join(input.EnabledChannels, ","), input.Endpoint, input.MerchantID, input.CredentialCiphertext, input.Acknowledgement, boolInt(input.Enabled), stamp(nowUTC()), stamp(nowUTC()))
	if err != nil {
		return model.PaymentProfile{}, fmt.Errorf("save payment profile: %w", err)
	}
	record, err := s.PaymentProfileRecordByID(ctx, input.ID, "")
	if err != nil {
		return model.PaymentProfile{}, err
	}
	return record.PaymentProfile, nil
}

func (s *Store) ListPaymentProfiles(ctx context.Context) ([]model.PaymentProfile, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,provider,provider_name,enabled_channels,endpoint,merchant_id,credential_ciphertext,acknowledgement,enabled FROM payment_profiles ORDER BY provider,provider_name,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]model.PaymentProfile, 0)
	for rows.Next() {
		profile, err := scanPaymentProfile(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, profile.PaymentProfile)
	}
	return result, rows.Err()
}

func (s *Store) PaymentProfile(ctx context.Context, provider, rail string) (model.PaymentProfile, error) {
	record, err := s.paymentProfileRecord(ctx, provider, rail)
	return record.PaymentProfile, err
}

func (s *Store) PaymentProfileRecord(ctx context.Context, provider, rail string) (PaymentProfileRecord, error) {
	return s.paymentProfileRecord(ctx, provider, rail)
}

// PaymentProfileRecordByID loads one profile account and optionally checks its rail.
func (s *Store) PaymentProfileRecordByID(ctx context.Context, id, rail string) (PaymentProfileRecord, error) {
	record, err := scanPaymentProfile(s.db.QueryRowContext(ctx, `SELECT id,provider,provider_name,enabled_channels,endpoint,merchant_id,credential_ciphertext,acknowledgement,enabled FROM payment_profiles WHERE id=?`, id))
	record.Rail = rail
	return record, err
}

func (s *Store) paymentProfileRecord(ctx context.Context, provider, rail string) (PaymentProfileRecord, error) {
	query := `SELECT id,provider,provider_name,enabled_channels,endpoint,merchant_id,credential_ciphertext,acknowledgement,enabled FROM payment_profiles WHERE provider=?`
	args := []any{provider}
	if rail != "" {
		query += ` AND (',' || enabled_channels || ',') LIKE ?`
		args = append(args, `%,`+rail+`,%`)
	}
	query += ` ORDER BY id LIMIT 1`
	record, err := scanPaymentProfile(s.db.QueryRowContext(ctx, query, args...))
	record.Rail = rail
	return record, err
}

func scanPaymentProfile(row interface{ Scan(...any) error }) (PaymentProfileRecord, error) {
	var record PaymentProfileRecord
	var channels string
	var enabled int
	if err := row.Scan(&record.ID, &record.Provider, &record.ProviderName, &channels, &record.Endpoint, &record.MerchantID, &record.CredentialCiphertext, &record.Acknowledgement, &enabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PaymentProfileRecord{}, ErrNotFound
		}
		return PaymentProfileRecord{}, fmt.Errorf("scan payment profile: %w", err)
	}
	record.Enabled = enabled == 1
	for _, channel := range strings.Split(channels, ",") {
		channel = strings.TrimSpace(channel)
		if channel != "" {
			record.EnabledChannels = append(record.EnabledChannels, channel)
		}
	}
	if record.ProviderName == "" {
		record.ProviderName = record.Provider
	}
	record.Configured = record.Endpoint != "" && record.CredentialCiphertext != ""
	if record.Configured {
		record.Credential = "********"
	}
	return record, nil
}

func nowUTC() time.Time { return time.Now().UTC() }
