package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	domain "github.com/txyyddss/Remna-User-Panel/internal/emby"
	"strings"
	"time"
)

const embyAccountSelect = `SELECT id,user_id,base_username,remote_user_id,remote_username,candidate_username,status,
	setup_price_txb_minor,setup_attempt,password_ciphertext,password_context,max_parental_rating,create_attempted,
	last_error,created_at,updated_at,provisioned_at,refunded_at,pending_preferences_json FROM emby_accounts`

// EmbyBaseUsername returns the user's immutable local Remnawave username.
func (s *Store) EmbyBaseUsername(ctx context.Context, userID string) (string, error) {
	var username string
	err := s.db.QueryRowContext(ctx, `SELECT username FROM users WHERE id=? AND username IS NOT NULL`, userID).Scan(&username)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("load Emby base username: %w", err)
	}
	return username, nil
}

// QueueEmbySetup atomically debits the configured price, appends its ledger
// row, stores the sealed password, and enqueues provisioning. Replays against a
// queued or active account do not debit again. A refunded failure starts a new
// billed attempt on the same durable account identity.
func (s *Store) QueueEmbySetup(ctx context.Context, input domain.QueueSetupInput, now time.Time) (domain.Account, bool, error) {
	if strings.TrimSpace(input.ID) == "" || strings.TrimSpace(input.UserID) == "" || strings.TrimSpace(input.BaseUsername) == "" ||
		input.SetupPriceTXBMinor < 0 || input.PasswordCiphertext == "" || input.PasswordContext == "" {
		return domain.Account{}, false, domain.ErrInvalidSetup
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Account{}, false, err
	}
	defer func() { _ = tx.Rollback() }()

	var canonicalUsername string
	if err := tx.QueryRowContext(ctx, `SELECT username FROM users WHERE id=? AND username IS NOT NULL`, input.UserID).Scan(&canonicalUsername); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Account{}, false, ErrNotFound
		}
		return domain.Account{}, false, fmt.Errorf("validate Emby setup user: %w", err)
	}
	if canonicalUsername != input.BaseUsername {
		return domain.Account{}, false, ErrConflict
	}

	existing, loadErr := scanEmbyProvisioning(tx.QueryRowContext(ctx, embyAccountSelect+` WHERE user_id=?`, input.UserID))
	accountID, attempt, isRetry := input.ID, 1, false
	if loadErr == nil {
		if err = hydrateEmbyPreferences(ctx, tx, &existing); err != nil {
			return domain.Account{}, false, err
		}
		if existing.Status != domain.StatusFailed || existing.RefundedAt == nil {
			return existing.Account, false, nil
		}
		accountID, attempt, isRetry = existing.ID, existing.SetupAttempt+1, true
	} else if !errors.Is(loadErr, domain.ErrNotFound) {
		return domain.Account{}, false, loadErr
	}

	newBalance, err := debitBalanceTx(ctx, tx, input.UserID, input.SetupPriceTXBMinor, now)
	if err != nil {
		return domain.Account{}, false, err
	}
	pendingPreferences, err := json.Marshal(input.Preferences)
	if err != nil {
		return domain.Account{}, false, fmt.Errorf("encode pending Emby preferences: %w", err)
	}
	if isRetry {
		if _, err := tx.ExecContext(ctx, `UPDATE emby_accounts SET base_username=?,
			candidate_username=CASE WHEN remote_user_id IS NULL THEN NULL ELSE candidate_username END,
			remote_username=CASE WHEN remote_user_id IS NULL THEN NULL ELSE remote_username END,
			status='queued',setup_price_txb_minor=?,setup_attempt=?,password_ciphertext=?,password_context=?,
			pending_preferences_json=?,create_attempted=CASE WHEN remote_user_id IS NULL THEN 0 ELSE create_attempted END,
			last_error='',refunded_at=NULL,updated_at=? WHERE id=?`,
			input.BaseUsername, input.SetupPriceTXBMinor, attempt, input.PasswordCiphertext, input.PasswordContext,
			string(pendingPreferences), stamp(now), accountID); err != nil {
			return domain.Account{}, false, fmt.Errorf("restart Emby setup: %w", err)
		}
	} else {
		if _, err := tx.ExecContext(ctx, `INSERT INTO emby_accounts(id,user_id,base_username,status,setup_price_txb_minor,
			setup_attempt,password_ciphertext,password_context,pending_preferences_json,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,?,?,?,?)`, accountID, input.UserID, input.BaseUsername, domain.StatusQueued,
			input.SetupPriceTXBMinor, attempt, input.PasswordCiphertext, input.PasswordContext,
			string(pendingPreferences), stamp(now), stamp(now)); err != nil {
			if isUniqueViolation(err) {
				return domain.Account{}, false, ErrConflict
			}
			return domain.Account{}, false, fmt.Errorf("insert Emby setup: %w", err)
		}
	}
	referenceID := fmt.Sprintf("%s:%d", accountID, attempt)
	if _, err := insertLedgerTx(ctx, tx, input.UserID, -input.SetupPriceTXBMinor, newBalance, "emby_setup_debit", referenceID, "Emby account setup", now); err != nil {
		return domain.Account{}, false, err
	}
	payload, err := embyProvisionPayload(accountID, attempt)
	if err != nil {
		return domain.Account{}, false, err
	}
	if err := insertOutboxTx(ctx, tx, domain.ProvisionOutboxKind, string(payload), now, now); err != nil {
		return domain.Account{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Account{}, false, fmt.Errorf("commit Emby setup: %w", err)
	}
	account, err := s.EmbyAccountForUser(ctx, input.UserID)
	return account, true, err
}

func embyProvisionPayload(accountID string, attempt int) ([]byte, error) {
	return json.Marshal(map[string]string{"accountId": accountID, "attempt": fmt.Sprintf("%d", attempt)})
}

// EmbyAccountForUser returns the safe account view for a local user.
func (s *Store) EmbyAccountForUser(ctx context.Context, userID string) (domain.Account, error) {
	record, err := scanEmbyProvisioning(s.db.QueryRowContext(ctx, embyAccountSelect+` WHERE user_id=?`, userID))
	if err != nil {
		return domain.Account{}, err
	}
	err = hydrateEmbyPreferences(ctx, s.db, &record)
	return record.Account, err
}

// ListEmbyAccounts returns safe account states without provisioning secrets.
func (s *Store) ListEmbyAccounts(ctx context.Context, limit int) ([]domain.Account, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, embyAccountSelect+` ORDER BY updated_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list Emby accounts: %w", err)
	}
	defer func() { _ = rows.Close() }()
	accounts := make([]domain.Account, 0)
	for rows.Next() {
		record, scanErr := scanEmbyProvisioning(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		scanErr = hydrateEmbyPreferences(ctx, s.db, &record)
		if scanErr != nil {
			return nil, scanErr
		}
		accounts = append(accounts, record.Account)
	}
	return accounts, rows.Err()
}

// EmbyProvisioningByID returns one server-only durable provisioning record.
func (s *Store) EmbyProvisioningByID(ctx context.Context, accountID string) (domain.ProvisioningRecord, error) {
	record, err := scanEmbyProvisioning(s.db.QueryRowContext(ctx, embyAccountSelect+` WHERE id=?`, accountID))
	if err != nil {
		return domain.ProvisioningRecord{}, err
	}
	err = hydrateEmbyPreferences(ctx, s.db, &record)
	return record, err
}
