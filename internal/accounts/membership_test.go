package accounts

import (
	"context"
	"errors"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"strings"
	"testing"
	"time"
)

func TestCreateInvites(t *testing.T) {
	t.Parallel()

	settings := &accountsSettings{values: map[string]string{
		"telegram.group_chat_id":   "-1001",
		"telegram.channel_chat_id": "-1002",
		"telegram.webhook_secret":  "test-invite-secret",
	}}
	repository := &accountsRepository{}
	telegram := &accountsTelegram{}
	service := newAccountsServiceForTest(repository, &accountsValidator{}, telegram, &accountsRemnawave{}, settings, 1)
	user := model.User{ID: strings.Repeat("x", 40), TelegramID: 42, OnboardingState: "intro"}

	invites, expiresAt, err := service.CreateInvites(context.Background(), user)
	if err != nil {
		t.Fatalf("CreateInvites(): %v", err)
	}
	if invites["group"] == "" || invites["channel"] == "" || !repository.advanced {
		t.Fatalf("CreateInvites() = %+v, advanced %t", invites, repository.advanced)
	}
	if len(telegram.inviteNames) != 2 || len(telegram.inviteNames[0]) > 32 || !expiresAt.Equal(service.now().UTC().Add(30*time.Minute)) {
		t.Fatalf("invite name/expiry = %d/%s", len(telegram.inviteNames[0]), expiresAt)
	}
}

func TestCreateInviteFailures(t *testing.T) {
	t.Parallel()

	testError := errors.New("failure")
	tests := []struct {
		name       string
		user       model.User
		settings   *accountsSettings
		repository *accountsRepository
		telegram   *accountsTelegram
		wantRevoke bool
	}{
		{name: "advance", user: model.User{ID: "user", OnboardingState: "intro"}, settings: validAccountsSettings(), repository: &accountsRepository{advanceErr: testError}, telegram: &accountsTelegram{}},
		{name: "setting", user: model.User{ID: "user"}, settings: &accountsSettings{errs: map[string]error{"telegram.group_chat_id": testError}}, repository: &accountsRepository{}, telegram: &accountsTelegram{}},
		{name: "invalid chat", user: model.User{ID: "user"}, settings: &accountsSettings{values: map[string]string{"telegram.group_chat_id": "nope"}}, repository: &accountsRepository{}, telegram: &accountsTelegram{}},
		{name: "zero chat", user: model.User{ID: "user"}, settings: &accountsSettings{values: map[string]string{"telegram.group_chat_id": "0"}}, repository: &accountsRepository{}, telegram: &accountsTelegram{}},
		{name: "telegram", user: model.User{ID: "user"}, settings: validAccountsSettings(), repository: &accountsRepository{}, telegram: &accountsTelegram{createErr: testError}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, _, err := newAccountsServiceForTest(test.repository, &accountsValidator{}, test.telegram, &accountsRemnawave{}, test.settings, 1).CreateInvites(context.Background(), test.user)
			if err == nil {
				t.Fatal("CreateInvites() unexpectedly succeeded")
			}
			if got := len(test.telegram.revokedLinks) > 0; got != test.wantRevoke {
				t.Fatalf("provider invite revoked = %t, want %t", got, test.wantRevoke)
			}
		})
	}
}

func TestCheckMembership(t *testing.T) {
	t.Parallel()

	user := model.User{ID: "user-1", TelegramID: 42}
	repository := &accountsRepository{user: user}
	telegram := &accountsTelegram{memberships: map[string]bool{"-1001": true, "-1002": false}}
	service := newAccountsServiceForTest(repository, &accountsValidator{}, telegram, &accountsRemnawave{}, validAccountsSettings(), 1)

	if _, err := service.CheckMembership(context.Background(), user); err != nil {
		t.Fatalf("CheckMembership(): %v", err)
	}
	if !repository.groupJoined || repository.channelJoined {
		t.Fatalf("membership update = %t/%t", repository.groupJoined, repository.channelJoined)
	}

	settingsFailure := validAccountsSettings()
	settingsFailure.errs = map[string]error{"telegram.group_chat_id": errors.New("settings")}
	if _, err := newAccountsServiceForTest(&accountsRepository{}, &accountsValidator{}, &accountsTelegram{}, &accountsRemnawave{}, settingsFailure, 1).CheckMembership(context.Background(), user); err == nil {
		t.Fatal("CheckMembership() ignored settings failure")
	}
	if _, err := newAccountsServiceForTest(&accountsRepository{}, &accountsValidator{}, &accountsTelegram{membershipErr: errors.New("telegram")}, &accountsRemnawave{}, validAccountsSettings(), 1).CheckMembership(context.Background(), user); err == nil {
		t.Fatal("CheckMembership() ignored Telegram failure")
	}
}

func TestRefreshMembershipByTelegramID(t *testing.T) {
	t.Parallel()

	user := model.User{ID: "user-1", TelegramID: 42}
	repository := &accountsRepository{user: user}
	telegram := &accountsTelegram{memberships: map[string]bool{"-1001": true, "-1002": true}}
	service := newAccountsServiceForTest(repository, &accountsValidator{}, telegram, &accountsRemnawave{}, validAccountsSettings(), 1)
	refreshed, err := service.RefreshMembershipByTelegramID(context.Background(), user.TelegramID)
	if err != nil || refreshed.ID != user.ID || !repository.groupJoined || !repository.channelJoined {
		t.Fatalf("RefreshMembershipByTelegramID() = (%+v, %v), joined %t/%t", refreshed, err, repository.groupJoined, repository.channelJoined)
	}
	repository.userByTelegramErr = errors.New("lookup failure")
	if _, err := service.RefreshMembershipByTelegramID(context.Background(), user.TelegramID); err == nil {
		t.Fatal("RefreshMembershipByTelegramID() ignored lookup failure")
	}
}

func TestHandleSignedJoinRequest(t *testing.T) {
	t.Parallel()

	const telegramID, chatID = int64(42), int64(-1001)
	settings := validAccountsSettings()
	repository := &accountsRepository{}
	telegram := &accountsTelegram{}
	service := newAccountsServiceForTest(repository, &accountsValidator{}, telegram, &accountsRemnawave{}, settings, 1)
	expiresAt := service.now().UTC().Add(time.Hour)
	name, err := service.signedInviteName(context.Background(), telegramID, chatID, expiresAt)
	if err != nil {
		t.Fatalf("signedInviteName(): %v", err)
	}
	if err := service.HandleSignedJoinRequest(context.Background(), telegramID, chatID, "invite", name, expiresAt); err != nil {
		t.Fatalf("HandleSignedJoinRequest(): %v", err)
	}
	if telegram.approveCalls != 1 || len(telegram.revokedLinks) != 1 {
		t.Fatalf("provider calls = approve %d, revoke %d", telegram.approveCalls, len(telegram.revokedLinks))
	}
	if err := service.HandleSignedJoinRequest(context.Background(), telegramID+1, chatID, "invite", name, expiresAt); !errors.Is(err, ErrInvalidAuthentication) {
		t.Fatalf("identity mismatch error = %v", err)
	}
	if err := service.HandleSignedJoinRequest(context.Background(), telegramID, chatID, "invite", name, service.now().UTC()); !errors.Is(err, ErrInvalidAuthentication) {
		t.Fatalf("expired signature error = %v", err)
	}
}
