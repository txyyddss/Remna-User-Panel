package emby

import (
	"context"
	"errors"
	"fmt"
)

// HandleProvisionJob is the generic outbox handler for ProvisionOutboxKind.
func (s *Service) HandleProvisionJob(ctx context.Context, aggregateID string) error {
	return s.Provision(ctx, aggregateID)
}

// Provision advances one account through an idempotent remote provisioning saga.
func (s *Service) Provision(ctx context.Context, accountID string) error {
	now := s.now().UTC()
	record, err := s.repository.BeginEmbyProvisioning(ctx, accountID, now)
	if err != nil {
		return err
	}
	if record.Status == StatusActive {
		return nil
	}
	if record.Status == StatusFailed {
		return nil
	}
	if record.PasswordCiphertext == "" || record.PasswordContext != passwordContext(record.UserID) {
		return s.failTerminal(ctx, record.ID, errors.New("temporary provisioning secret is unavailable"))
	}
	password, err := s.secrets.Open(record.PasswordContext, record.PasswordCiphertext)
	if err != nil {
		return s.failTerminal(ctx, record.ID, fmt.Errorf("open temporary provisioning secret: %w", err))
	}
	defer zero(password)

	remoteUser, err := s.resolveOrCreate(ctx, &record)
	if err != nil {
		return s.handleProvisionError(ctx, record.ID, err)
	}
	if err := s.repository.SetEmbyRemoteIdentity(ctx, record.ID, remoteUser.ID, remoteUser.Name, s.now().UTC()); err != nil {
		return s.handleProvisionError(ctx, record.ID, fmt.Errorf("persist Emby remote identity: %w", err))
	}
	current, err := s.remote.GetUser(ctx, remoteUser.ID)
	if err != nil {
		return s.handleProvisionError(ctx, record.ID, fmt.Errorf("load created Emby user: %w", err))
	}
	if err := s.remote.SetPassword(ctx, current.ID, nil, password); err != nil {
		return s.handleProvisionError(ctx, record.ID, fmt.Errorf("set Emby password: %w", redactRemoteError(err, "remote password update failed")))
	}
	if err := s.remote.UpdatePolicy(ctx, current.ID, HardenPolicy(current.Policy, record.Preferences)); err != nil {
		return s.handleProvisionError(ctx, record.ID, fmt.Errorf("apply restricted Emby policy: %w", err))
	}
	verified, err := s.remote.GetUser(ctx, current.ID)
	if err != nil {
		return s.handleProvisionError(ctx, record.ID, fmt.Errorf("verify Emby policy: %w", err))
	}
	if !PolicyMatchesPreferences(verified.Policy, record.Preferences) {
		return s.handleProvisionError(ctx, record.ID, errors.New("Emby policy verification failed"))
	}
	if err := s.repository.MarkEmbyProvisioned(ctx, record.ID, record.Preferences, s.now().UTC()); err != nil {
		return fmt.Errorf("complete Emby provisioning: %w", err)
	}
	return nil
}
