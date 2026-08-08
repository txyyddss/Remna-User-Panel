// Package accounts implements trusted Telegram authentication and resumable onboarding.
package accounts

import (
	"context"
	"crypto/sha256"
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
	CompleteOnboarding(context.Context, string, string, string, time.Time) (model.User, error)
	SaveJoinInvite(context.Context, string, string, int64, string, time.Time) (database.JoinInvite, error)
	JoinInviteByLink(context.Context, string) (database.JoinInvite, error)
	MarkJoinInviteUsed(context.Context, string, time.Time) error
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
			if verifyErr != nil {
				return model.User{}, "", time.Time{}, fmt.Errorf("%w: %v", ErrUpstreamUnavailable, verifyErr)
			}
			if !exists {
				repository, supported := s.repository.(recoveryRepository)
				if !supported {
					return model.User{}, "", time.Time{}, errors.New("Remnawave recovery is unavailable")
				}
				user, err = repository.BeginRemnawaveRecovery(ctx, user.ID, "remnawave_user_missing", s.now().UTC())
				if err != nil {
					return model.User{}, "", time.Time{}, err
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
		inviteName := "TXC " + user.ID
		if len(inviteName) > 32 {
			inviteName = inviteName[:32]
		}
		link, err := s.telegram.CreateJoinRequestInvite(ctx, chatIDValue, inviteName, expiresAt)
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("create %s invite: %w", kind, err)
		}
		if _, err := s.repository.SaveJoinInvite(ctx, user.ID, kind, chatID, link, expiresAt); err != nil {
			_ = s.telegram.RevokeInviteLink(ctx, chatIDValue, link)
			return nil, time.Time{}, err
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

// HandleJoinRequest approves only the identity bound to the exact invite link, then revokes it.
func (s *Service) HandleJoinRequest(ctx context.Context, telegramID, chatID int64, inviteLink string) error {
	invite, err := s.repository.JoinInviteByLink(ctx, inviteLink)
	if err != nil {
		return err
	}
	user, err := s.repository.UserByID(ctx, invite.UserID)
	if err != nil {
		return err
	}
	if user.TelegramID != telegramID || invite.ChatID != chatID || invite.ApprovedAt != nil || invite.RevokedAt != nil || !s.now().UTC().Before(invite.ExpiresAt) {
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
	return s.repository.MarkJoinInviteUsed(ctx, invite.ID, s.now().UTC())
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

// AcceptAgreement creates or reconciles the permanent upstream identity.
func (s *Service) AcceptAgreement(ctx context.Context, user model.User, accepted bool) (model.User, error) {
	if !accepted || user.Username == nil || user.OnboardingState != "agreement" {
		return model.User{}, database.ErrConflict
	}
	remote, exists, err := s.remnawave.FindUserByUsername(ctx, *user.Username)
	if err != nil {
		return model.User{}, fmt.Errorf("reconcile Remnawave user: %w", err)
	}
	if exists && (remote.TelegramID == nil || *remote.TelegramID != user.TelegramID) {
		return model.User{}, ErrUsernameUnavailable
	}
	if !exists {
		byTelegram, telegramExists, telegramErr := s.remnawave.FindUserByTelegramID(ctx, user.TelegramID)
		if telegramErr != nil {
			return model.User{}, fmt.Errorf("reconcile Remnawave Telegram identity: %w", telegramErr)
		}
		if telegramExists {
			if byTelegram.Username != *user.Username {
				return model.User{}, ErrUsernameUnavailable
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
			return model.User{}, fmt.Errorf("create Remnawave user: %w", err)
		}
	}
	return s.repository.CompleteOnboarding(ctx, user.ID, remote.ID, remote.SubscriptionURL, s.now().UTC())
}
