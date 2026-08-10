package emby

import (
	"context"
	"errors"
	"fmt"
)

// UpdatePreferences validates upstream choices, fetches the complete current
// policy, and overlays only the allowed fields plus mandatory restrictions.
func (s *Service) UpdatePreferences(ctx context.Context, userID string, preferences Preferences) (Account, error) {
	preferences = normalizePreferences(preferences)
	if err := s.validatePreferences(ctx, preferences); err != nil {
		return Account{}, err
	}
	account, err := s.repository.EmbyAccountForUser(ctx, userID)
	if err != nil {
		return Account{}, err
	}
	if account.Status != StatusActive || account.RemoteUserID == "" {
		return Account{}, ErrRemoteAccountMissing
	}
	remoteUser, err := s.remote.GetUser(ctx, account.RemoteUserID)
	if err != nil {
		if s.remote.IsNotFound(err) {
			return Account{}, ErrRemoteAccountMissing
		}
		return Account{}, fmt.Errorf("load linked Emby user: %w", err)
	}
	if err := s.remote.UpdatePolicy(ctx, remoteUser.ID, HardenPolicy(remoteUser.Policy, preferences)); err != nil {
		return Account{}, fmt.Errorf("update Emby policy: %w", err)
	}
	verified, err := s.remote.GetUser(ctx, remoteUser.ID)
	if err != nil {
		return Account{}, fmt.Errorf("verify Emby policy: %w", err)
	}
	if !PolicyMatchesPreferences(verified.Policy, preferences) {
		return Account{}, errors.New("Emby policy verification failed")
	}
	return s.repository.UpdateEmbyPreferences(ctx, account.ID, preferences, s.now().UTC())
}

// ChangePassword synchronously updates a linked Emby password without storing it.
func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if len(newPassword) == 0 || len(newPassword) > maxPasswordBytes || len(currentPassword) > maxPasswordBytes {
		return ErrInvalidSetup
	}
	account, err := s.repository.EmbyAccountForUser(ctx, userID)
	if err != nil {
		return err
	}
	if account.Status != StatusActive || account.RemoteUserID == "" {
		return ErrRemoteAccountMissing
	}
	currentBytes, nextBytes := []byte(currentPassword), []byte(newPassword)
	defer zero(currentBytes)
	defer zero(nextBytes)
	if err := s.remote.SetPassword(ctx, account.RemoteUserID, currentBytes, nextBytes); err != nil {
		if s.remote.IsNotFound(err) {
			return ErrRemoteAccountMissing
		}
		return fmt.Errorf("change Emby password: %w", redactRemoteError(err, "remote password update failed"))
	}
	return s.repository.TouchEmbyAccount(ctx, account.ID, s.now().UTC())
}
