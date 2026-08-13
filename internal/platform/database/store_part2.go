package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"strings"
	"time"
)

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

func stamp(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }

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
