package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	domain "github.com/txyyddss/Remna-User-Panel/internal/emby"
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
	payload, err := json.Marshal(map[string]string{"accountId": accountID})
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

// RetryEmbyProvisioning enqueues one retained transient setup unless work is
// already pending. Terminal failures cannot pass the secret/state predicate.
func (s *Store) RetryEmbyProvisioning(ctx context.Context, accountID string, now time.Time) (domain.Account, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Account{}, err
	}
	defer func() { _ = tx.Rollback() }()
	record, err := scanEmbyProvisioning(tx.QueryRowContext(ctx, embyAccountSelect+` WHERE id=?`, accountID))
	if err != nil {
		return domain.Account{}, err
	}
	if record.Status == domain.StatusActive {
		err = hydrateEmbyPreferences(ctx, tx, &record)
		return record.Account, err
	}
	if record.Status != domain.StatusQueued || record.PasswordCiphertext == "" {
		return domain.Account{}, ErrConflict
	}
	payload, marshalErr := json.Marshal(map[string]string{"accountId": accountID})
	if marshalErr != nil {
		return domain.Account{}, marshalErr
	}
	var existing int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM outbox_jobs WHERE kind=? AND payload=? AND status IN ('pending','processing')`,
		domain.ProvisionOutboxKind, string(payload)).Scan(&existing); err != nil {
		return domain.Account{}, fmt.Errorf("check Emby retry work: %w", err)
	}
	if existing == 0 {
		if err := insertOutboxTx(ctx, tx, domain.ProvisionOutboxKind, string(payload), now, now); err != nil {
			return domain.Account{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.Account{}, fmt.Errorf("commit Emby retry: %w", err)
	}
	return s.EmbyAccountForUser(ctx, record.UserID)
}

// BeginEmbyProvisioning leases the account state to a kind-specific outbox handler.
func (s *Store) BeginEmbyProvisioning(ctx context.Context, accountID string, now time.Time) (domain.ProvisioningRecord, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if _, err := s.db.ExecContext(ctx, `UPDATE emby_accounts SET status='provisioning',updated_at=? WHERE id=? AND status='queued'`, stamp(now), accountID); err != nil {
		return domain.ProvisioningRecord{}, fmt.Errorf("begin Emby provisioning: %w", err)
	}
	return s.EmbyProvisioningByID(ctx, accountID)
}

// SetEmbyCandidate persists the exact name before any create attempt.
func (s *Store) SetEmbyCandidate(ctx context.Context, accountID, candidate string, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE emby_accounts SET candidate_username=?,updated_at=?
		WHERE id=? AND status='provisioning' AND candidate_username IS NULL`, candidate, stamp(now), accountID)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAccountExists
		}
		return fmt.Errorf("set Emby candidate: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		var stored sql.NullString
		if loadErr := s.db.QueryRowContext(ctx, `SELECT candidate_username FROM emby_accounts WHERE id=?`, accountID).Scan(&stored); loadErr != nil {
			return mapEmbyNotFound(loadErr)
		}
		if !stored.Valid || stored.String != candidate {
			return domain.ErrAccountExists
		}
	}
	return nil
}

// MarkEmbyCreateAttempted records the ambiguity boundary before POST /Users/New.
func (s *Store) MarkEmbyCreateAttempted(ctx context.Context, accountID string, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE emby_accounts SET create_attempted=1,updated_at=?
		WHERE id=? AND status='provisioning' AND candidate_username IS NOT NULL`, stamp(now), accountID)
	if err != nil {
		return fmt.Errorf("mark Emby create attempt: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// SetEmbyRemoteIdentity stores the authoritative remote identity idempotently.
func (s *Store) SetEmbyRemoteIdentity(ctx context.Context, accountID, remoteID, remoteUsername string, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE emby_accounts SET remote_user_id=?,remote_username=?,candidate_username=?,updated_at=?
		WHERE id=? AND status='provisioning' AND (remote_user_id IS NULL OR remote_user_id=?)`,
		remoteID, remoteUsername, remoteUsername, stamp(now), accountID, remoteID)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAccountExists
		}
		return fmt.Errorf("set Emby remote identity: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return domain.ErrAccountExists
	}
	return nil
}

// RequeueEmbyProvisioning preserves the secret and records a retry-safe error.
func (s *Store) RequeueEmbyProvisioning(ctx context.Context, accountID string, provisionErr error, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE emby_accounts SET status='queued',last_error=?,updated_at=?
		WHERE id=? AND status='provisioning'`, sanitizeError(provisionErr), stamp(now), accountID)
	if err != nil {
		return fmt.Errorf("requeue Emby provisioning: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// MarkEmbyProvisioned persists disabled folders only after the service has
// re-fetched and verified the remote full policy.
func (s *Store) MarkEmbyProvisioned(ctx context.Context, accountID string, preferences domain.Preferences, now time.Time) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx, `UPDATE emby_accounts SET status='active',password_ciphertext='',password_context='',
		max_parental_rating=?,pending_preferences_json='{}',last_error='',provisioned_at=COALESCE(provisioned_at,?),updated_at=?
		WHERE id=? AND status='provisioning' AND remote_user_id IS NOT NULL`, nullableInt32(preferences.MaxParentalRating), stamp(now), stamp(now), accountID)
	if err != nil {
		return fmt.Errorf("mark Emby provisioned: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrConflict
	}
	if err := replaceEmbyFoldersTx(ctx, tx, accountID, preferences.DisabledLibraryIDs); err != nil {
		return err
	}
	return tx.Commit()
}

// FailAndRefundEmbySetup atomically records a terminal failure, erases the
// temporary password, and credits the exact debited setup price once.
func (s *Store) FailAndRefundEmbySetup(ctx context.Context, accountID, reason string, now time.Time) (domain.Account, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Account{}, err
	}
	defer func() { _ = tx.Rollback() }()
	record, err := scanEmbyProvisioning(tx.QueryRowContext(ctx, embyAccountSelect+` WHERE id=?`, accountID))
	if err != nil {
		return domain.Account{}, err
	}
	if record.Status == domain.StatusActive {
		return domain.Account{}, ErrConflict
	}
	if record.RefundedAt != nil {
		err = hydrateEmbyPreferences(ctx, tx, &record)
		return record.Account, err
	}
	balance, err := adjustBalanceTx(ctx, tx, record.UserID, record.SetupPriceTXBMinor, now)
	if err != nil {
		return domain.Account{}, fmt.Errorf("refund Emby setup balance: %w", err)
	}
	referenceID := fmt.Sprintf("%s:%d", record.ID, record.SetupAttempt)
	if _, err := insertLedgerTx(ctx, tx, record.UserID, record.SetupPriceTXBMinor, balance, "emby_setup_refund", referenceID, "Emby setup refund", now); err != nil {
		return domain.Account{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE emby_accounts SET status='failed',password_ciphertext='',password_context='',pending_preferences_json='{}',
		last_error=?,refunded_at=?,updated_at=? WHERE id=?`, truncateEmbyError(reason), stamp(now), stamp(now), record.ID); err != nil {
		return domain.Account{}, fmt.Errorf("record failed Emby setup: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.Account{}, fmt.Errorf("commit Emby setup refund: %w", err)
	}
	return s.EmbyAccountForUser(ctx, record.UserID)
}

// UpdateEmbyPreferences persists only preferences that were accepted upstream.
func (s *Store) UpdateEmbyPreferences(ctx context.Context, accountID string, preferences domain.Preferences, now time.Time) (domain.Account, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Account{}, err
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx, `UPDATE emby_accounts SET max_parental_rating=?,updated_at=? WHERE id=? AND status='active'`,
		nullableInt32(preferences.MaxParentalRating), stamp(now), accountID)
	if err != nil {
		return domain.Account{}, fmt.Errorf("update Emby preferences: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return domain.Account{}, domain.ErrNotFound
	}
	if err := replaceEmbyFoldersTx(ctx, tx, accountID, preferences.DisabledLibraryIDs); err != nil {
		return domain.Account{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Account{}, err
	}
	record, err := s.EmbyProvisioningByID(ctx, accountID)
	return record.Account, err
}

// TouchEmbyAccount records a successful password operation without storing password data.
func (s *Store) TouchEmbyAccount(ctx context.Context, accountID string, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE emby_accounts SET updated_at=? WHERE id=? AND status='active'`, stamp(now), accountID)
	if err != nil {
		return fmt.Errorf("touch Emby account: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanEmbyProvisioning(row rowScanner) (domain.ProvisioningRecord, error) {
	var record domain.ProvisioningRecord
	var remoteID, remoteUsername, candidate sql.NullString
	var parentalRating sql.NullInt64
	var provisioned, refunded sql.NullString
	var created, updated string
	var createAttempted int
	if err := row.Scan(&record.ID, &record.UserID, &record.BaseUsername, &remoteID, &remoteUsername, &candidate, &record.Status,
		&record.SetupPriceTXBMinor, &record.SetupAttempt, &record.PasswordCiphertext, &record.PasswordContext,
		&parentalRating, &createAttempted, &record.LastError, &created, &updated, &provisioned, &refunded, &record.PendingPreferencesJSON); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ProvisioningRecord{}, domain.ErrNotFound
		}
		return domain.ProvisioningRecord{}, fmt.Errorf("scan Emby account: %w", err)
	}
	record.RemoteUserID = remoteID.String
	record.RemoteUsername = remoteUsername.String
	record.CandidateUsername = candidate.String
	record.CreateAttempted = createAttempted == 1
	record.Retryable = record.Status == domain.StatusQueued && record.PasswordCiphertext != ""
	if parentalRating.Valid {
		if parentalRating.Int64 < -2147483648 || parentalRating.Int64 > 2147483647 {
			return domain.ProvisioningRecord{}, errors.New("invalid stored Emby parental rating")
		}
		converted := int32(parentalRating.Int64)
		record.Preferences.MaxParentalRating = &converted
	}
	var err error
	record.CreatedAt, err = parseStamp(created)
	if err != nil {
		return domain.ProvisioningRecord{}, err
	}
	record.UpdatedAt, err = parseStamp(updated)
	if err != nil {
		return domain.ProvisioningRecord{}, err
	}
	if provisioned.Valid {
		value, parseErr := parseStamp(provisioned.String)
		if parseErr != nil {
			return domain.ProvisioningRecord{}, parseErr
		}
		record.ProvisionedAt = &value
	}
	if refunded.Valid {
		value, parseErr := parseStamp(refunded.String)
		if parseErr != nil {
			return domain.ProvisioningRecord{}, parseErr
		}
		record.RefundedAt = &value
	}
	return record, nil
}

type embyQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func hydrateEmbyPreferences(ctx context.Context, queryer embyQueryer, record *domain.ProvisioningRecord) error {
	disabled, err := embyFolderIDs(ctx, queryer, record.ID)
	if err != nil {
		return err
	}
	record.Preferences.DisabledLibraryIDs = disabled
	if (record.Status == domain.StatusQueued || record.Status == domain.StatusProvisioning) && record.PendingPreferencesJSON != "" && record.PendingPreferencesJSON != "{}" {
		var pending domain.Preferences
		if err := json.Unmarshal([]byte(record.PendingPreferencesJSON), &pending); err != nil {
			return fmt.Errorf("decode pending Emby preferences: %w", err)
		}
		record.Preferences = pending
	}
	return nil
}

func embyFolderIDs(ctx context.Context, queryer embyQueryer, accountID string) ([]string, error) {
	rows, err := queryer.QueryContext(ctx, `SELECT folder_id FROM emby_account_disabled_folders WHERE account_id=? ORDER BY folder_id`, accountID)
	if err != nil {
		return nil, fmt.Errorf("list Emby folders: %w", err)
	}
	defer func() { _ = rows.Close() }()
	result := make([]string, 0)
	for rows.Next() {
		var folderID string
		if err := rows.Scan(&folderID); err != nil {
			return nil, err
		}
		result = append(result, folderID)
	}
	return result, rows.Err()
}

func replaceEmbyFoldersTx(ctx context.Context, tx *sql.Tx, accountID string, folderIDs []string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM emby_account_disabled_folders WHERE account_id=?`, accountID); err != nil {
		return fmt.Errorf("clear Emby folders: %w", err)
	}
	seen := make(map[string]struct{}, len(folderIDs))
	for _, folderID := range folderIDs {
		folderID = strings.TrimSpace(folderID)
		if folderID == "" {
			continue
		}
		if _, exists := seen[folderID]; exists {
			continue
		}
		seen[folderID] = struct{}{}
		if _, err := tx.ExecContext(ctx, `INSERT INTO emby_account_disabled_folders(account_id,folder_id) VALUES(?,?)`, accountID, folderID); err != nil {
			return fmt.Errorf("insert Emby folder: %w", err)
		}
	}
	return nil
}

func nullableInt32(value *int32) any {
	if value == nil {
		return nil
	}
	return *value
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "unique constraint")
}

func mapEmbyNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

func truncateEmbyError(value string) string {
	if len(value) > 500 {
		return value[:500]
	}
	return value
}

var _ domain.Repository = (*Store)(nil)
