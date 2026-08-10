// Package accounts implements trusted Telegram authentication and resumable onboarding.
package accounts

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

var usernamePattern = regexp.MustCompile(`^[a-z]{3,9}$`)

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

// RemoteUser is the minimum Remnawave identity persisted locally.
type RemoteUser struct {
	ID              string
	Username        string
	TelegramID      *int64
	SubscriptionURL string
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
	CompleteOnboardingRevision(context.Context, string, string, string, int, []string, time.Time) (model.User, error)
}

// Service coordinates authentication and onboarding state.
type Service struct {
	repository Repository
	validator  InitDataValidator
	telegram   TelegramClient
	remnawave  RemnawaveClient
	settings   Settings
	adminID    int64
	sessionTTL time.Duration
	now        func() time.Time
}

// NewService constructs an accounts service.
func NewService(repository Repository, validator InitDataValidator, telegram TelegramClient, remnawave RemnawaveClient, settings Settings, adminID int64, sessionTTL time.Duration) *Service {
	return &Service{repository: repository, validator: validator, telegram: telegram, remnawave: remnawave, settings: settings, adminID: adminID, sessionTTL: sessionTTL, now: time.Now}
}

// Authenticate exchanges fresh initData for a server-owned opaque session.
func (s *Service) Authenticate(ctx context.Context, raw string) (model.User, string, time.Time, error) {
	profile, err := s.validator.Validate(raw)
	if err != nil {
		return model.User{}, "", time.Time{}, fmt.Errorf("%w: %v", ErrInvalidAuthentication, err)
	}
	user, _, err := s.repository.UpsertTelegramUser(ctx, profile, profile.ID == s.adminID)
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

// CreateInvites creates short-lived, identity-bound links for both required chats.
func (s *Service) CreateInvites(ctx context.Context, user model.User) (map[string]string, time.Time, error) {
	if user.OnboardingState == "intro" {
		if err := s.repository.AdvanceToMembership(ctx, user.ID); err != nil {
			return nil, time.Time{}, err
		}
	}
	expiresAt := s.now().UTC().Add(30 * time.Minute)
	result := make(map[string]string, 2)
	type createdInvite struct{ chatID, link string }
	created := make([]createdInvite, 0, 2)
	complete := false
	defer func() {
		if complete {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, invite := range created {
			_ = s.telegram.RevokeInviteLink(cleanupCtx, invite.chatID, invite.link)
		}
	}()
	for _, kind := range []string{"group", "channel"} {
		chatIDValue, err := s.settings.Plaintext(ctx, "telegram."+kind+"_chat_id")
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("load %s chat: %w", kind, err)
		}
		chatID, err := strconv.ParseInt(chatIDValue, 10, 64)
		if err != nil || chatID == 0 {
			return nil, time.Time{}, fmt.Errorf("invalid %s chat id", kind)
		}
		inviteName, err := s.signedInviteName(ctx, user.TelegramID, chatID, expiresAt)
		if err != nil {
			return nil, time.Time{}, err
		}
		link, err := s.telegram.CreateJoinRequestInvite(ctx, chatIDValue, inviteName, expiresAt)
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("create %s invite: %w", kind, err)
		}
		created = append(created, createdInvite{chatID: chatIDValue, link: link})
		result[kind] = link
	}
	complete = true
	return result, expiresAt, nil
}

// CheckMembership asks Telegram for canonical state rather than trusting the browser.
func (s *Service) CheckMembership(ctx context.Context, user model.User) (model.User, error) {
	joined := make(map[string]bool, 2)
	for _, kind := range []string{"group", "channel"} {
		chatID, err := s.settings.Plaintext(ctx, "telegram."+kind+"_chat_id")
		if err != nil {
			return model.User{}, err
		}
		joined[kind], err = s.telegram.GetMembership(ctx, chatID, user.TelegramID)
		if err != nil {
			return model.User{}, fmt.Errorf("check %s membership: %w", kind, err)
		}
	}
	return s.repository.UpdateMembership(ctx, user.ID, joined["group"], joined["channel"])
}

// RefreshMembershipByTelegramID updates onboarding state after a Telegram membership event.
func (s *Service) RefreshMembershipByTelegramID(ctx context.Context, telegramID int64) (model.User, error) {
	user, err := s.repository.UserByTelegramID(ctx, telegramID)
	if err != nil {
		return model.User{}, err
	}
	return s.CheckMembership(ctx, user)
}

// HandleSignedJoinRequest verifies signature, identity, chat, and expiry before
// approval and immediate invite revocation. No invite link is stored locally.
func (s *Service) HandleSignedJoinRequest(ctx context.Context, telegramID, chatID int64, inviteLink, inviteName string, expiresAt time.Time) error {
	if strings.TrimSpace(inviteLink) == "" || len(inviteName) > 32 || !s.now().UTC().Before(expiresAt.UTC()) {
		return ErrInvalidAuthentication
	}
	parts := strings.Split(inviteName, ".")
	if len(parts) != 2 {
		return ErrInvalidAuthentication
	}
	signedTelegramID, err := strconv.ParseInt(parts[0], 36, 64)
	if err != nil || signedTelegramID != telegramID {
		return ErrInvalidAuthentication
	}
	expected, err := s.inviteSignature(ctx, signedTelegramID, chatID, expiresAt)
	if err != nil || !hmac.Equal([]byte(parts[1]), []byte(expected)) {
		return ErrInvalidAuthentication
	}
	chatIDValue := strconv.FormatInt(chatID, 10)
	if err := s.telegram.ApproveJoinRequest(ctx, chatIDValue, telegramID); err != nil {
		present, membershipErr := s.telegram.GetMembership(ctx, chatIDValue, telegramID)
		if membershipErr != nil || !present {
			return err
		}
	}
	if err := s.telegram.RevokeInviteLink(ctx, chatIDValue, inviteLink); err != nil {
		return err
	}
	return nil
}

func (s *Service) signedInviteName(ctx context.Context, telegramID, chatID int64, expiresAt time.Time) (string, error) {
	signature, err := s.inviteSignature(ctx, telegramID, chatID, expiresAt)
	if err != nil {
		return "", err
	}
	name := strconv.FormatInt(telegramID, 36) + "." + signature
	if len(name) > 32 {
		return "", errors.New("Telegram invite identity exceeds name limit")
	}
	return name, nil
}

func (s *Service) inviteSignature(ctx context.Context, telegramID, chatID int64, expiresAt time.Time) (string, error) {
	secret, err := s.settings.Plaintext(ctx, "telegram.webhook_secret")
	if err != nil || strings.TrimSpace(secret) == "" {
		return "", errors.New("Telegram invite signing secret is unavailable")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = fmt.Fprintf(mac, "%d|%d|%d", telegramID, chatID, expiresAt.UTC().Unix())
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)[:12]), nil
}

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
		return model.User{}, database.ErrConflict
	}
	currentRevision, requiredIDs, err := s.repository.CurrentAgreementContract(ctx)
	if err != nil {
		return model.User{}, err
	}
	if revision != currentRevision || !sameStringSet(requiredIDs, agreementIDs) {
		return model.User{}, database.ErrConflict
	}
	remote, err := s.reconcileAgreementUser(ctx, user)
	if err != nil {
		return model.User{}, err
	}
	return s.repository.CompleteOnboardingRevision(ctx, user.ID, remote.ID, remote.SubscriptionURL, revision, agreementIDs, s.now().UTC())
}

func (s *Service) reconcileAgreementUser(ctx context.Context, user model.User) (RemoteUser, error) {
	remote, exists, err := s.remnawave.FindUserByUsername(ctx, *user.Username)
	if err != nil {
		return RemoteUser{}, fmt.Errorf("reconcile Remnawave user: %w", err)
	}
	if exists && (remote.TelegramID == nil || *remote.TelegramID != user.TelegramID) {
		return RemoteUser{}, ErrUsernameUnavailable
	}
	if !exists {
		byTelegram, telegramExists, telegramErr := s.remnawave.FindUserByTelegramID(ctx, user.TelegramID)
		if telegramErr != nil {
			return RemoteUser{}, fmt.Errorf("reconcile Remnawave Telegram identity: %w", telegramErr)
		}
		if telegramExists {
			if byTelegram.Username != *user.Username {
				return RemoteUser{}, ErrUsernameUnavailable
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
				err = ErrUsernameUnavailable
			}
		}
		if err != nil {
			return RemoteUser{}, fmt.Errorf("create Remnawave user: %w", err)
		}
	}
	return remote, nil
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
