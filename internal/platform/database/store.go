package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

// ErrNotFound indicates that the requested domain record does not exist.
var ErrNotFound = errors.New("record not found")

// ErrConflict indicates that a uniqueness or state transition invariant was violated.
var ErrConflict = errors.New("record conflicts with current state")

// ErrInsufficientBalance indicates that a debit would make an account negative.
var ErrInsufficientBalance = errors.New("insufficient TXB balance")

// ErrPaymentCapacity indicates that all retained payment slots contain live,
// non-prunable orders. Callers may retry after an order settles or expires.
var ErrPaymentCapacity = errors.New("payment order capacity is full")

// Store implements persistent repositories on top of SQLite.
type Store struct {
	db      *sql.DB
	writeMu sync.Mutex
}

// NewStore creates an application store.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

// DB exposes the connection pool for health checks and online backup.
func (s *Store) DB() *sql.DB { return s.db }

// UpsertTelegramUser creates or refreshes a user from validated Telegram data.
func (s *Store) UpsertTelegramUser(ctx context.Context, profile model.TelegramProfile, isAdmin bool) (model.User, bool, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	now := time.Now().UTC()
	role := "user"
	if isAdmin {
		role = "admin"
	}
	userID, err := ids.New()
	if err != nil {
		return model.User{}, false, err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO users(id,telegram_id,telegram_first_name,telegram_last_name,telegram_username,role,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(telegram_id) DO UPDATE SET telegram_first_name=excluded.telegram_first_name,
		telegram_last_name=excluded.telegram_last_name,telegram_username=excluded.telegram_username,
		role=excluded.role,updated_at=excluded.updated_at`,
		userID, profile.ID, profile.FirstName, profile.LastName, profile.Username, role, stamp(now), stamp(now))
	if err != nil {
		return model.User{}, false, fmt.Errorf("upsert Telegram user: %w", err)
	}
	user, err := s.UserByTelegramID(ctx, profile.ID)
	if err != nil {
		return model.User{}, false, err
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO balances(user_id,txb_minor,updated_at) VALUES(?,?,?) ON CONFLICT(user_id) DO NOTHING`, user.ID, 0, stamp(now)); err != nil {
		return model.User{}, false, fmt.Errorf("initialize balance: %w", err)
	}
	return user, user.ID == userID, nil
}

// UserByID returns a user by internal ID.
func (s *Store) UserByID(ctx context.Context, userID string) (model.User, error) {
	return scanUser(s.db.QueryRowContext(ctx, userSelect+` WHERE users.id=?`, userID))
}

// UserByTelegramID returns a user by trusted Telegram ID.
func (s *Store) UserByTelegramID(ctx context.Context, telegramID int64) (model.User, error) {
	return scanUser(s.db.QueryRowContext(ctx, userSelect+` WHERE users.telegram_id=?`, telegramID))
}

const userSelect = `SELECT users.id,users.telegram_id,users.telegram_first_name,users.telegram_last_name,users.telegram_username,
	users.username,users.role,users.onboarding_state,users.group_joined,users.channel_joined,users.policy_accepted_at,
	users.accepted_agreement_revision,users.remna_user_id,users.recovery_reason,users.created_at,users.updated_at FROM users`

type rowScanner interface{ Scan(dest ...any) error }

func scanUser(row rowScanner) (model.User, error) {
	var user model.User
	var username, policy, recoveryReason sql.NullString
	var remnaID sql.NullString
	var groupJoined, channelJoined int
	var createdAt, updatedAt string
	if err := row.Scan(&user.ID, &user.TelegramID, &user.TelegramFirstName, &user.TelegramLastName, &user.TelegramUsername,
		&username, &user.Role, &user.OnboardingState, &groupJoined, &channelJoined, &policy, &user.AgreementRevision, &remnaID, &recoveryReason, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("scan user: %w", err)
	}
	user.Username = nullableString(username)
	user.GroupJoined = groupJoined == 1
	user.ChannelJoined = channelJoined == 1
	if policy.Valid {
		parsed, err := parseStamp(policy.String)
		if err != nil {
			return model.User{}, fmt.Errorf("parse policy timestamp: %w", err)
		}
		user.PolicyAcceptedAt = &parsed
	}
	user.RemnaUserID = nullableString(remnaID)
	user.RecoveryReason = recoveryReason.String
	var err error
	user.CreatedAt, err = parseStamp(createdAt)
	if err != nil {
		return model.User{}, fmt.Errorf("parse user creation timestamp: %w", err)
	}
	user.UpdatedAt, err = parseStamp(updatedAt)
	if err != nil {
		return model.User{}, fmt.Errorf("parse user update timestamp: %w", err)
	}
	return user, nil
}

// CreateSession stores a hash of an opaque session token.
func (s *Store) CreateSession(ctx context.Context, tokenHash []byte, userID string, expiresAt time.Time) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	now := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, `INSERT INTO sessions(token_hash,user_id,expires_at,created_at) VALUES(?,?,?,?)`, tokenHash, userID, stamp(expiresAt), stamp(now)); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// UserBySession returns the owner of a non-expired session.
func (s *Store) UserBySession(ctx context.Context, tokenHash []byte, now time.Time) (model.User, error) {
	return scanUser(s.db.QueryRowContext(ctx, userSelect+` JOIN sessions ON sessions.user_id=users.id WHERE sessions.token_hash=? AND sessions.expires_at>?`, tokenHash, stamp(now)))
}

// DeleteExpiredSessions removes stale session rows.
func (s *Store) DeleteExpiredSessions(ctx context.Context, now time.Time) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	if _, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at<=?`, stamp(now)); err != nil {
		return fmt.Errorf("delete expired sessions: %w", err)
	}
	return nil
}

// UpdateMembership persists canonical Telegram membership results.
func (s *Store) UpdateMembership(ctx context.Context, userID string, groupJoined, channelJoined bool) (model.User, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	_, err := s.db.ExecContext(ctx, `UPDATE users SET group_joined=?,channel_joined=?,onboarding_state=CASE
		WHEN onboarding_state IN ('intro','membership') AND ?=1 AND ?=1 THEN CASE WHEN username IS NULL THEN 'username' ELSE 'agreement' END
		ELSE onboarding_state END,updated_at=? WHERE id=?`, boolInt(groupJoined), boolInt(channelJoined), boolInt(groupJoined), boolInt(channelJoined), stamp(time.Now().UTC()), userID)
	if err != nil {
		return model.User{}, fmt.Errorf("update membership: %w", err)
	}
	return s.UserByID(ctx, userID)
}

// BeginRemnawaveRecovery restarts only external membership/agreement checks and
// preserves local identity, balance, purchases, and feature history.
func (s *Store) BeginRemnawaveRecovery(ctx context.Context, userID, reason string, now time.Time) (model.User, error) {
	reason = strings.TrimSpace(reason)
	if strings.TrimSpace(userID) == "" || reason == "" {
		return model.User{}, ErrConflict
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	result, err := s.db.ExecContext(ctx, `UPDATE users SET onboarding_state='membership',group_joined=0,channel_joined=0,
		policy_accepted_at=NULL,remna_user_id=NULL,remna_subscription_url=NULL,recovery_reason=?,updated_at=?
		WHERE id=? AND onboarding_state='complete'`, reason, stamp(now), userID)
	if err != nil {
		return model.User{}, fmt.Errorf("begin Remnawave recovery: %w", err)
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return model.User{}, fmt.Errorf("inspect Remnawave recovery: %w", rowsErr)
	} else if affected != 1 {
		// Concurrent authentication requests can both confirm the same linked 404.
		// Once one request establishes recovery, identical retries are successful.
		user, loadErr := s.UserByID(ctx, userID)
		if loadErr != nil {
			return model.User{}, loadErr
		}
		if user.OnboardingState == "membership" && user.RecoveryReason == reason && user.RemnaUserID == nil &&
			!user.GroupJoined && !user.ChannelJoined && user.PolicyAcceptedAt == nil {
			return user, nil
		}
		return model.User{}, ErrConflict
	}
	return s.UserByID(ctx, userID)
}

// AdvanceToMembership moves a new user past the intro animation.
func (s *Store) AdvanceToMembership(ctx context.Context, userID string) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	if _, err := s.db.ExecContext(ctx, `UPDATE users SET onboarding_state='membership',updated_at=? WHERE id=? AND onboarding_state='intro'`, stamp(time.Now().UTC()), userID); err != nil {
		return fmt.Errorf("advance onboarding: %w", err)
	}
	return nil
}

// ReserveUsername atomically assigns a locally unique username.
func (s *Store) ReserveUsername(ctx context.Context, userID, username string) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	result, err := s.db.ExecContext(ctx, `UPDATE users SET username=?,onboarding_state='agreement',updated_at=?
		WHERE id=? AND (onboarding_state='username' OR (onboarding_state='agreement' AND username=?))`,
		username, stamp(time.Now().UTC()), userID, username)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return ErrConflict
		}
		return fmt.Errorf("reserve username: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect username reservation: %w", err)
	}
	if affected == 0 {
		return ErrConflict
	}
	return nil
}

func stamp(value time.Time) string               { return value.UTC().Format(time.RFC3339Nano) }
func parseStamp(value string) (time.Time, error) { return time.Parse(time.RFC3339Nano, value) }
func nullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
