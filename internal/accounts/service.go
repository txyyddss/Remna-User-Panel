// Package accounts implements trusted Telegram authentication and resumable onboarding.
package accounts

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,36}$`)

// ErrInvalidAuthentication denotes untrusted or expired Telegram launch data.
var ErrInvalidAuthentication = errors.New("invalid Telegram authentication")

// ErrMembershipRequired means both configured chats are not yet joined.
var ErrMembershipRequired = errors.New("membership is required")

// ErrUsernameUnavailable means either the local database or Remnawave owns the name.
var ErrUsernameUnavailable = errors.New("username is unavailable")

// ErrUpstreamUnavailable distinguishes a temporary Remnawave failure from bad
// Telegram authentication data.
var ErrUpstreamUnavailable = errors.New("Remnawave is temporarily unavailable")

// InitDataValidator authenticates raw Mini App initData and returns only trusted identity fields.
type InitDataValidator interface {
	Validate(raw string) (model.TelegramProfile, error)
}

// TelegramClient is the narrow Bot API surface used by onboarding.
type TelegramClient interface {
	CreateJoinRequestInvite(ctx context.Context, chatID, name string, expiresAt time.Time) (string, error)
	GetMembership(ctx context.Context, chatID string, telegramID int64) (bool, error)
	ApproveJoinRequest(ctx context.Context, chatID string, telegramID int64) error
	RevokeInviteLink(ctx context.Context, chatID, inviteLink string) error
}

// RemoteUser is the minimum Remnawave identity used during onboarding.
type RemoteUser struct {
	ID         string
	Username   string
	TelegramID *int64
}

// RemoteCreateUser is the immutable v1 onboarding contract.
type RemoteCreateUser struct {
	Username             string
	TelegramID           int64
	Status               string
	ExpireAt             time.Time
	TrafficLimitBytes    int64
	TrafficLimitStrategy string
	ActiveInternalSquads []string
}

// RemnawaveClient is owned by this service and adapted by the integration package.
type RemnawaveClient interface {
	FindUserByUsername(ctx context.Context, username string) (RemoteUser, bool, error)
	FindUserByTelegramID(ctx context.Context, telegramID int64) (RemoteUser, bool, error)
	CreateUser(ctx context.Context, input RemoteCreateUser) (RemoteUser, error)
	IsDuplicateError(err error) bool
}

type linkedUserVerifier interface {
	FindUserByID(context.Context, string) (RemoteUser, bool, error)
}

type recoveryRepository interface {
	BeginRemnawaveRecovery(context.Context, string, string, time.Time) (model.User, error)
}

// Settings supplies validated runtime configuration.
type Settings interface {
	Plaintext(ctx context.Context, key string) (string, error)
}

// Repository is the persistence surface used by accounts.
type Repository interface {
	UpsertTelegramUser(context.Context, model.TelegramProfile, bool) (model.User, bool, error)
	CreateSession(context.Context, []byte, string, time.Time) error
	UserBySession(context.Context, []byte, time.Time) (model.User, error)
	UserByID(context.Context, string) (model.User, error)
	UserByTelegramID(context.Context, int64) (model.User, error)
	AdvanceToMembership(context.Context, string) error
	UpdateMembership(context.Context, string, bool, bool) (model.User, error)
	ReserveUsername(context.Context, string, string) error
	CurrentAgreementContract(context.Context) (int, []string, error)
	CompleteOnboardingRevision(context.Context, string, string, int, []string, time.Time) (model.User, error)
}

// Service coordinates authentication and onboarding state.
type Service struct {
	repository Repository
	validator  InitDataValidator
	telegram   TelegramClient
	remnawave  RemnawaveClient
	settings   Settings
	adminIDs   map[int64]struct{}
	sessionTTL time.Duration
	now        func() time.Time
}

// NewService constructs an accounts service.
func NewService(repository Repository, validator InitDataValidator, telegram TelegramClient, remnawave RemnawaveClient, settings Settings, adminIDs []int64, sessionTTL time.Duration) *Service {
	configuredAdmins := make(map[int64]struct{}, len(adminIDs))
	for _, adminID := range adminIDs {
		configuredAdmins[adminID] = struct{}{}
	}
	return &Service{repository: repository, validator: validator, telegram: telegram, remnawave: remnawave, settings: settings, adminIDs: configuredAdmins, sessionTTL: sessionTTL, now: time.Now}
}

func (s *Service) isAdminTelegramID(telegramID int64) bool {
	_, ok := s.adminIDs[telegramID]
	return ok
}

// Authenticate exchanges fresh initData for a server-owned opaque session.
func (s *Service) Authenticate(ctx context.Context, raw string) (model.User, string, time.Time, error) {
	profile, err := s.validator.Validate(raw)
	if err != nil {
		return model.User{}, "", time.Time{}, fmt.Errorf("%w: %v", ErrInvalidAuthentication, err)
	}
	user, _, err := s.repository.UpsertTelegramUser(ctx, profile, s.isAdminTelegramID(profile.ID))
	if err != nil {
		return model.User{}, "", time.Time{}, err
	}
	if user.OnboardingState == "complete" && user.RemnaUserID != nil {
		if verifier, ok := s.remnawave.(linkedUserVerifier); ok {
			_, exists, verifyErr := verifier.FindUserByID(ctx, *user.RemnaUserID)
			// Telegram initData is the authentication boundary. A temporary
			// Remnawave outage must not turn a valid Telegram launch into an
			// authentication failure; the local account can still use the app
			// and remote-backed operations will report their own availability.
			if verifyErr == nil && !exists {
				repository, supported := s.repository.(recoveryRepository)
				if supported {
					user, err = repository.BeginRemnawaveRecovery(ctx, user.ID, "remnawave_user_missing", s.now().UTC())
					if err != nil {
						return model.User{}, "", time.Time{}, err
					}
				}
			}
		}
	}
	token, err := ids.Token(32)
	if err != nil {
		return model.User{}, "", time.Time{}, err
	}
	expiresAt := s.now().UTC().Add(s.sessionTTL)
	hash := sha256.Sum256([]byte(token))
	if err := s.repository.CreateSession(ctx, hash[:], user.ID, expiresAt); err != nil {
		return model.User{}, "", time.Time{}, err
	}
	return user, token, expiresAt, nil
}

// UserBySession authenticates a cookie without storing the bearer token itself.
func (s *Service) UserBySession(ctx context.Context, token string) (model.User, error) {
	if token == "" {
		return model.User{}, ErrInvalidAuthentication
	}
	hash := sha256.Sum256([]byte(token))
	user, err := s.repository.UserBySession(ctx, hash[:], s.now().UTC())
	if err != nil {
		return model.User{}, fmt.Errorf("%w: %v", ErrInvalidAuthentication, err)
	}
	return user, nil
}
