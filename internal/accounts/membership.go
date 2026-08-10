package accounts

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"strconv"
	"strings"
	"time"
)

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
