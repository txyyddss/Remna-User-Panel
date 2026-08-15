package accounts

import (
	"context"
	"errors"
	"fmt"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"strings"
	"time"
)

// ReserveUsername applies local syntax and uniqueness plus an upstream preflight.
func (s *Service) ReserveUsername(ctx context.Context, user model.User, username string) (model.User, error) {
	username = strings.TrimSpace(username)
	if !usernamePattern.MatchString(username) {
		return model.User{}, fmt.Errorf("username must match %s", usernamePattern.String())
	}
	if !user.GroupJoined || !user.ChannelJoined {
		return model.User{}, ErrMembershipRequired
	}
	if _, exists, err := s.remnawave.FindUserByUsername(ctx, username); err != nil {
		return model.User{}, fmt.Errorf("preflight Remnawave username: %w", err)
	} else if exists {
		return model.User{}, ErrUsernameUnavailable
	}
	if err := s.repository.ReserveUsername(ctx, user.ID, username); err != nil {
		if errors.Is(err, database.ErrConflict) {
			return model.User{}, ErrUsernameUnavailable
		}
		return model.User{}, err
	}
	return s.repository.UserByID(ctx, user.ID)
}

// AcceptAgreementRevision rejects stale revisions and requires every currently
// published agreement ID before reconciling the permanent Remnawave identity.
func (s *Service) AcceptAgreementRevision(ctx context.Context, user model.User, revision int, agreementIDs []string) (model.User, error) {
	if user.Username == nil || user.OnboardingState != "agreement" || revision <= 0 {
		return model.User{}, fmt.Errorf("%w: %w", ErrAgreementStateConflict, database.ErrConflict)
	}
	currentRevision, requiredIDs, err := s.repository.CurrentAgreementContract(ctx)
	if err != nil {
		return model.User{}, err
	}
	if revision != currentRevision || !sameStringSet(requiredIDs, agreementIDs) {
		return model.User{}, fmt.Errorf("%w: %w", ErrAgreementRevisionConflict, database.ErrConflict)
	}
	remote, err := s.reconcileAgreementUser(ctx, user)
	if err != nil {
		return model.User{}, err
	}
	return s.repository.CompleteOnboardingRevision(ctx, user.ID, remote.ID, revision, agreementIDs, s.now().UTC())
}

func (s *Service) reconcileAgreementUser(ctx context.Context, user model.User) (RemoteUser, error) {
	// A published agreement revision temporarily moves completed local users
	// back to agreement. Their persisted Remnawave ID is the durable identity;
	// use it before username lookup so an old upstream record cannot look like a
	// new-account username collision.
	if user.RemnaUserID != nil && strings.TrimSpace(*user.RemnaUserID) != "" {
		if verifier, ok := s.remnawave.(linkedUserVerifier); ok {
			remote, exists, err := verifier.FindUserByID(ctx, *user.RemnaUserID)
			if err != nil {
				return RemoteUser{}, fmt.Errorf("reconcile linked Remnawave user: %w", err)
			}
			if exists {
				return remote, nil
			}
		}
	}
	remote, exists, err := s.remnawave.FindUserByUsername(ctx, *user.Username)
	if err != nil {
		return RemoteUser{}, fmt.Errorf("reconcile Remnawave user: %w", err)
	}
	if exists && (remote.TelegramID == nil || *remote.TelegramID != user.TelegramID) {
		return RemoteUser{}, remnawaveIdentityConflict()
	}
	if !exists {
		byTelegram, telegramExists, telegramErr := s.remnawave.FindUserByTelegramID(ctx, user.TelegramID)
		if telegramErr != nil {
			return RemoteUser{}, fmt.Errorf("reconcile Remnawave Telegram identity: %w", telegramErr)
		}
		if telegramExists {
			if byTelegram.Username != *user.Username {
				return RemoteUser{}, remnawaveIdentityConflict()
			}
			remote, exists = byTelegram, true
		}
	}
	if !exists {
		remote, err = s.remnawave.CreateUser(ctx, RemoteCreateUser{
			Username: *user.Username, TelegramID: user.TelegramID, Status: "ACTIVE",
			ExpireAt: time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC), TrafficLimitBytes: 0,
			TrafficLimitStrategy: "NO_RESET", ActiveInternalSquads: []string{},
		})
		if err != nil && s.remnawave.IsDuplicateError(err) {
			remote, exists, err = s.remnawave.FindUserByUsername(ctx, *user.Username)
			if err == nil && (!exists || remote.TelegramID == nil || *remote.TelegramID != user.TelegramID) {
				err = remnawaveIdentityConflict()
			}
		}
		if err != nil {
			return RemoteUser{}, fmt.Errorf("create Remnawave user: %w", err)
		}
	}
	return remote, nil
}

func remnawaveIdentityConflict() error {
	return fmt.Errorf("%w: %w", ErrRemnawaveIdentityConflict, ErrUsernameUnavailable)
}

func sameStringSet(expected, provided []string) bool {
	if len(expected) != len(provided) {
		return false
	}
	seen := make(map[string]struct{}, len(expected))
	for _, value := range expected {
		seen[value] = struct{}{}
	}
	for _, value := range provided {
		if _, exists := seen[value]; !exists {
			return false
		}
		delete(seen, value)
	}
	return len(seen) == 0
}
