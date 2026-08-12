package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

// PaymentProfileInput contains already-encrypted rail credentials.
type PaymentProfileInput struct {
	ID, Provider, Rail, ChannelName, Endpoint, MerchantID string
	CredentialCiphertext, Acknowledgement                 string
	Enabled                                               bool
}

// PaymentProfileRecord is the private storage representation of a rail.
type PaymentProfileRecord struct {
	model.PaymentProfile
	CredentialCiphertext string
}

func (s *Store) SavePaymentProfile(ctx context.Context, input PaymentProfileInput) (model.PaymentProfile, error) {
	if input.ID == "" || input.Provider == "" || input.Rail == "" || input.ChannelName == "" || input.Endpoint == "" {
		return model.PaymentProfile{}, ErrConflict
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO payment_rail_profiles(id,provider,rail,channel_name,endpoint,merchant_id,credential_ciphertext,acknowledgement,enabled,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(provider,rail) DO UPDATE SET id=excluded.id,channel_name=excluded.channel_name,endpoint=excluded.endpoint,merchant_id=excluded.merchant_id,credential_ciphertext=CASE WHEN excluded.credential_ciphertext='' THEN payment_rail_profiles.credential_ciphertext ELSE excluded.credential_ciphertext END,acknowledgement=excluded.acknowledgement,enabled=excluded.enabled,updated_at=excluded.updated_at`,
		input.ID, input.Provider, input.Rail, input.ChannelName, input.Endpoint, input.MerchantID, input.CredentialCiphertext, input.Acknowledgement, boolInt(input.Enabled), stamp(nowUTC()), stamp(nowUTC()))
	if err != nil {
		return model.PaymentProfile{}, fmt.Errorf("save payment profile: %w", err)
	}
	return s.PaymentProfile(ctx, input.Provider, input.Rail)
}

func (s *Store) ListPaymentProfiles(ctx context.Context) ([]model.PaymentProfile, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,provider,rail,channel_name,endpoint,merchant_id,credential_ciphertext,acknowledgement,enabled FROM payment_rail_profiles ORDER BY provider,rail`)
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

func (s *Store) paymentProfileRecord(ctx context.Context, provider, rail string) (PaymentProfileRecord, error) {
	return scanPaymentProfile(s.db.QueryRowContext(ctx, `SELECT id,provider,rail,channel_name,endpoint,merchant_id,credential_ciphertext,acknowledgement,enabled FROM payment_rail_profiles WHERE provider=? AND rail=?`, provider, rail))
}

func scanPaymentProfile(row interface{ Scan(...any) error }) (PaymentProfileRecord, error) {
	var record PaymentProfileRecord
	var enabled int
	if err := row.Scan(&record.ID, &record.Provider, &record.Rail, &record.ChannelName, &record.Endpoint, &record.MerchantID, &record.CredentialCiphertext, &record.Acknowledgement, &enabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PaymentProfileRecord{}, ErrNotFound
		}
		return PaymentProfileRecord{}, fmt.Errorf("scan payment profile: %w", err)
	}
	record.Enabled = enabled == 1
	record.Configured = record.Endpoint != "" && record.CredentialCiphertext != ""
	if record.Configured {
		record.Credential = "********"
	}
	return record, nil
}

func nowUTC() time.Time { return time.Now().UTC() }
