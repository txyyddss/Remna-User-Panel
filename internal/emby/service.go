package emby

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

const maxPasswordBytes = 1024

// Service coordinates priced setup, durable provisioning, and linked updates.
type Service struct {
	repository Repository
	remote     Remote
	prices     PriceSource
	secrets    SecretBox
	now        func() time.Time
}

// NewService constructs an Emby account service.
func NewService(repository Repository, remote Remote, prices PriceSource, secrets SecretBox) *Service {
	return &Service{repository: repository, remote: remote, prices: prices, secrets: secrets, now: time.Now}
}

// Account returns the caller's safe local Emby account state.
func (s *Service) Account(ctx context.Context, userID string) (Account, error) {
	return s.repository.EmbyAccountForUser(ctx, userID)
}

// ListAccounts returns safe account states for administrative monitoring.
func (s *Service) ListAccounts(ctx context.Context, limit int) ([]Account, error) {
	return s.repository.ListEmbyAccounts(ctx, limit)
}

// RetryProvisioning re-enqueues a retained transient setup by account id.
// Refunded terminal failures cannot be retried because their secret is erased.
func (s *Service) RetryProvisioning(ctx context.Context, accountID string) (Account, error) {
	return s.repository.RetryEmbyProvisioning(ctx, accountID, s.now().UTC())
}

// Options retrieves the current selectable folders and parental ratings.
func (s *Service) Options(ctx context.Context) (Options, error) {
	folders, err := s.remote.ListSelectableFolders(ctx)
	if err != nil {
		return Options{}, fmt.Errorf("list Emby folders: %w", err)
	}
	ratings, err := s.remote.ListParentalRatings(ctx)
	if err != nil {
		return Options{}, fmt.Errorf("list Emby parental ratings: %w", err)
	}
	return Options{Folders: folders, Ratings: ratings}, nil
}

// Setup validates current Emby choices, seals the temporary password, and
// atomically debits the setup price while queueing provisioning.
func (s *Service) Setup(ctx context.Context, userID, password string, preferences Preferences) (Account, bool, error) {
	if strings.TrimSpace(userID) == "" || len(password) == 0 || len(password) > maxPasswordBytes {
		return Account{}, false, ErrInvalidSetup
	}
	existing, err := s.repository.EmbyAccountForUser(ctx, userID)
	if err == nil && existing.Status != StatusFailed {
		return existing, false, nil
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return Account{}, false, err
	}
	preferences = normalizePreferences(preferences)
	if err := s.validatePreferences(ctx, preferences); err != nil {
		return Account{}, false, err
	}
	baseUsername, err := s.repository.EmbyBaseUsername(ctx, userID)
	if err != nil {
		return Account{}, false, fmt.Errorf("load Emby base username: %w", err)
	}
	baseUsername = strings.TrimSpace(baseUsername)
	if baseUsername == "" {
		return Account{}, false, fmt.Errorf("%w: local username is empty", ErrInvalidSetup)
	}
	price, err := s.prices.EmbySetupPriceTXBMinor(ctx)
	if err != nil {
		return Account{}, false, fmt.Errorf("load Emby setup price: %w", err)
	}
	if price < 0 {
		return Account{}, false, fmt.Errorf("%w: setup price is negative", ErrInvalidSetup)
	}
	accountID, err := ids.New()
	if err != nil {
		return Account{}, false, err
	}
	passwordBytes := []byte(password)
	defer zero(passwordBytes)
	secretContext := passwordContext(userID)
	ciphertext, err := s.secrets.Seal(secretContext, passwordBytes)
	if err != nil {
		return Account{}, false, fmt.Errorf("seal Emby provisioning password: %w", err)
	}
	account, created, err := s.repository.QueueEmbySetup(ctx, QueueSetupInput{
		ID: accountID, UserID: userID, BaseUsername: baseUsername,
		PasswordCiphertext: ciphertext, PasswordContext: secretContext,
		SetupPriceTXBMinor: price, Preferences: preferences,
	}, s.now().UTC())
	if err != nil {
		return Account{}, false, err
	}
	return account, created, nil
}

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
	if err := s.repository.MarkEmbyProvisioned(ctx, record.ID, s.now().UTC()); err != nil {
		return fmt.Errorf("complete Emby provisioning: %w", err)
	}
	return nil
}

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

func (s *Service) resolveOrCreate(ctx context.Context, record *ProvisioningRecord) (RemoteUser, error) {
	if record.RemoteUserID != "" {
		user, err := s.remote.GetUser(ctx, record.RemoteUserID)
		if err == nil {
			return user, nil
		}
		if !s.remote.IsNotFound(err) {
			return RemoteUser{}, err
		}
		// The stored id may have been lost after an Emby restore. Reconcile by
		// the persisted exact candidate before considering the setup terminal.
		if record.CandidateUsername != "" {
			user, exists, findErr := s.remote.FindUserByName(ctx, record.CandidateUsername)
			if findErr != nil {
				return RemoteUser{}, findErr
			}
			if exists {
				return user, nil
			}
		}
		return RemoteUser{}, ErrRemoteAccountMissing
	}

	if record.CandidateUsername == "" {
		candidate, err := s.chooseCandidate(ctx, record.BaseUsername, record.UserID)
		if err != nil {
			return RemoteUser{}, err
		}
		if err := s.repository.SetEmbyCandidate(ctx, record.ID, candidate, s.now().UTC()); err != nil {
			return RemoteUser{}, err
		}
		record.CandidateUsername = candidate
	}
	if record.CreateAttempted {
		user, exists, err := s.remote.FindUserByName(ctx, record.CandidateUsername)
		if err != nil {
			return RemoteUser{}, err
		}
		if exists {
			return user, nil
		}
	} else {
		_, exists, err := s.remote.FindUserByName(ctx, record.CandidateUsername)
		if err != nil {
			return RemoteUser{}, err
		}
		if exists {
			return RemoteUser{}, fmt.Errorf("%w: candidate appeared before create", ErrAccountExists)
		}
	}
	if err := s.repository.MarkEmbyCreateAttempted(ctx, record.ID, s.now().UTC()); err != nil {
		return RemoteUser{}, err
	}
	created, createErr := s.remote.CreateUser(ctx, record.CandidateUsername)
	if createErr == nil {
		return created, nil
	}
	if s.remote.IsTerminal(createErr) {
		return RemoteUser{}, fmt.Errorf("create Emby user: %w", createErr)
	}
	// Emby can commit creation and lose the response. The candidate was
	// persisted and preflighted before this call, so exact-name reconciliation
	// is the only identity adoption allowed.
	reconciled, exists, reconcileErr := s.remote.FindUserByName(ctx, record.CandidateUsername)
	if reconcileErr == nil && exists {
		return reconciled, nil
	}
	if reconcileErr != nil && !s.remote.IsTerminal(createErr) {
		return RemoteUser{}, fmt.Errorf("create Emby user: %w; reconcile: %v", createErr, reconcileErr)
	}
	return RemoteUser{}, fmt.Errorf("create Emby user: %w", createErr)
}

func (s *Service) chooseCandidate(ctx context.Context, baseUsername, userID string) (string, error) {
	baseUsername = strings.TrimSpace(baseUsername)
	if baseUsername == "" {
		return "", fmt.Errorf("%w: empty candidate username", ErrInvalidSetup)
	}
	_, exists, err := s.remote.FindUserByName(ctx, baseUsername)
	if err != nil {
		return "", fmt.Errorf("preflight Emby username: %w", err)
	}
	if !exists {
		return baseUsername, nil
	}
	suffixed := SuffixedUsername(baseUsername, userID)
	_, exists, err = s.remote.FindUserByName(ctx, suffixed)
	if err != nil {
		return "", fmt.Errorf("preflight suffixed Emby username: %w", err)
	}
	if exists {
		return "", fmt.Errorf("%w: generated Emby username already exists", ErrAccountExists)
	}
	return suffixed, nil
}

func (s *Service) validatePreferences(ctx context.Context, preferences Preferences) error {
	options, err := s.Options(ctx)
	if err != nil {
		return err
	}
	availableFolders := make(map[string]struct{}, len(options.Folders))
	for _, folder := range options.Folders {
		availableFolders[folder.ID] = struct{}{}
	}
	for _, selected := range preferences.LibraryIDs {
		if _, ok := availableFolders[selected]; !ok {
			return fmt.Errorf("%w: unknown Emby library", ErrInvalidSetup)
		}
	}
	if preferences.MaxParentalRating != nil {
		found := false
		for _, rating := range options.Ratings {
			if rating.Value == *preferences.MaxParentalRating {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%w: unknown Emby parental rating", ErrInvalidSetup)
		}
	}
	return nil
}

func (s *Service) handleProvisionError(ctx context.Context, accountID string, provisionErr error) error {
	if errors.Is(provisionErr, ErrRemoteAccountMissing) || errors.Is(provisionErr, ErrAccountExists) ||
		errors.Is(provisionErr, ErrInvalidSetup) || s.remote.IsTerminal(provisionErr) {
		return s.failTerminal(ctx, accountID, provisionErr)
	}
	if err := s.repository.RequeueEmbyProvisioning(ctx, accountID, provisionErr, s.now().UTC()); err != nil {
		return fmt.Errorf("record Emby provisioning retry: %w", err)
	}
	return provisionErr
}

func (s *Service) failTerminal(ctx context.Context, accountID string, cause error) error {
	if _, err := s.repository.FailAndRefundEmbySetup(ctx, accountID, safeFailure(cause), s.now().UTC()); err != nil {
		return fmt.Errorf("refund failed Emby setup: %w", err)
	}
	return nil
}

// SuffixedUsername appends a stable eight-character hash to a colliding base name.
func SuffixedUsername(baseUsername, userID string) string {
	hash := sha256.Sum256([]byte(userID))
	return strings.TrimSpace(baseUsername) + "-" + hex.EncodeToString(hash[:4])
}

func normalizePreferences(preferences Preferences) Preferences {
	seen := make(map[string]struct{}, len(preferences.LibraryIDs))
	libraries := make([]string, 0, len(preferences.LibraryIDs))
	for _, libraryID := range preferences.LibraryIDs {
		libraryID = strings.TrimSpace(libraryID)
		if libraryID == "" {
			continue
		}
		if _, exists := seen[libraryID]; exists {
			continue
		}
		seen[libraryID] = struct{}{}
		libraries = append(libraries, libraryID)
	}
	sort.Strings(libraries)
	preferences.LibraryIDs = libraries
	return preferences
}

func passwordContext(userID string) string { return "emby.provisioning.password:" + userID }

func zero(value []byte) {
	for index := range value {
		value[index] = 0
	}
}

func safeFailure(err error) string {
	if err == nil {
		return "provisioning failed"
	}
	value := err.Error()
	if len(value) > 500 {
		return value[:500]
	}
	return value
}

// redactedRemoteError keeps the original cause available for errors.Is/As and
// terminal classification while ensuring a provider or transport cannot echo
// password material into logs, outbox diagnostics, or durable account errors.
type redactedRemoteError struct {
	cause   error
	message string
}

func (e redactedRemoteError) Error() string { return e.message }
func (e redactedRemoteError) Unwrap() error { return e.cause }

func redactRemoteError(cause error, message string) error {
	return redactedRemoteError{cause: cause, message: message}
}
