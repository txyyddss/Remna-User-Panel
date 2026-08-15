package accounts

import (
	"context"
	"errors"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"testing"
	"time"
)

func TestReserveUsername(t *testing.T) {
	t.Parallel()

	joinedUser := model.User{ID: "user-1", GroupJoined: true, ChannelJoined: true}
	testError := errors.New("failure")
	tests := []struct {
		name       string
		username   string
		user       model.User
		repository *accountsRepository
		remote     *accountsRemnawave
		want       error
	}{
		{name: "success trims", username: "  alice ", user: joinedUser, repository: &accountsRepository{user: model.User{ID: "user-1"}}, remote: &accountsRemnawave{}},
		{name: "too short", username: "ab", user: joinedUser, repository: &accountsRepository{}, remote: &accountsRemnawave{}},
		{name: "uppercase", username: "Alice", user: joinedUser, repository: &accountsRepository{}, remote: &accountsRemnawave{}},
		{name: "membership", username: "alice", user: model.User{ID: "user-1"}, repository: &accountsRepository{}, remote: &accountsRemnawave{}, want: ErrMembershipRequired},
		{name: "preflight error", username: "alice", user: joinedUser, repository: &accountsRepository{}, remote: &accountsRemnawave{findResponses: []accountsFindResponse{{err: testError}}}},
		{name: "upstream owns", username: "alice", user: joinedUser, repository: &accountsRepository{}, remote: &accountsRemnawave{findResponses: []accountsFindResponse{{exists: true}}}, want: ErrUsernameUnavailable},
		{name: "local race", username: "alice", user: joinedUser, repository: &accountsRepository{reserveErr: database.ErrConflict}, remote: &accountsRemnawave{}, want: ErrUsernameUnavailable},
		{name: "reserve failure", username: "alice", user: joinedUser, repository: &accountsRepository{reserveErr: testError}, remote: &accountsRemnawave{}, want: testError},
		{name: "reload failure", username: "alice", user: joinedUser, repository: &accountsRepository{userByIDErr: testError}, remote: &accountsRemnawave{}, want: testError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := newAccountsServiceForTest(test.repository, &accountsValidator{}, &accountsTelegram{}, test.remote, &accountsSettings{}, 1).ReserveUsername(context.Background(), test.user, test.username)
			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("ReserveUsername() error = %v, want %v", err, test.want)
			}
			if test.want == nil && (test.name == "success trims" || test.name == "uppercase") {
				wantUsername := test.username
				if test.name == "success trims" {
					wantUsername = "alice"
				}
				if err != nil || test.repository.reservedUsername != wantUsername {
					t.Fatalf("ReserveUsername() = %v, reserved %q", err, test.repository.reservedUsername)
				}
			} else if test.want == nil && err == nil {
				t.Fatal("ReserveUsername() unexpectedly succeeded")
			}
		})
	}
}

func TestAcceptAgreementCreatesOrReconcilesUser(t *testing.T) {
	t.Parallel()

	username := "alice"
	telegramID := int64(42)
	remote := RemoteUser{ID: "remote-1", Username: username, TelegramID: &telegramID}
	tests := []struct {
		name       string
		remote     *accountsRemnawave
		wantCreate bool
	}{
		{name: "existing matching user", remote: &accountsRemnawave{findResponses: []accountsFindResponse{{remote: remote, exists: true}}}},
		{name: "new user", remote: &accountsRemnawave{createRemote: remote}, wantCreate: true},
		{name: "duplicate race reconciles", remote: &accountsRemnawave{
			findResponses: []accountsFindResponse{{}, {remote: remote, exists: true}}, createErr: errAccountsDuplicate,
		}, wantCreate: true},
		{name: "existing matching Telegram user", remote: &accountsRemnawave{
			telegramResponse: accountsFindResponse{remote: remote, exists: true},
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := &accountsRepository{user: model.User{ID: "user-1", OnboardingState: "complete"}, agreementRevision: 1, requiredAgreementIDs: []string{"terms"}}
			service := newAccountsServiceForTest(repository, &accountsValidator{}, &accountsTelegram{}, test.remote, &accountsSettings{}, 1)
			user := model.User{ID: "user-1", TelegramID: telegramID, Username: &username, OnboardingState: "agreement"}
			completed, err := service.AcceptAgreementRevision(context.Background(), user, 1, []string{"terms"})
			if err != nil || completed.OnboardingState != "complete" {
				t.Fatalf("AcceptAgreementRevision() = (%+v, %v)", completed, err)
			}
			if got := len(test.remote.created) > 0; got != test.wantCreate {
				t.Fatalf("CreateUser called = %t, want %t", got, test.wantCreate)
			}
			if test.wantCreate {
				input := test.remote.created[0]
				if input.Status != "ACTIVE" || input.TelegramID != telegramID || input.TrafficLimitBytes != 0 || input.TrafficLimitStrategy != "NO_RESET" || len(input.ActiveInternalSquads) != 0 || input.ExpireAt != time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC) {
					t.Fatalf("CreateUser input = %+v", input)
				}
			}
			if repository.completedRemoteID != remote.ID {
				t.Fatalf("completion remote ID = %q, want %q", repository.completedRemoteID, remote.ID)
			}
		})
	}
}

func TestAcceptAgreementUsesPersistedRemnawaveIdentity(t *testing.T) {
	t.Parallel()

	username := "alice"
	telegramID := int64(42)
	remoteID := "remote-1"
	remote := RemoteUser{ID: remoteID, Username: username}
	repository := &accountsRepository{user: model.User{ID: "user-1", OnboardingState: "complete"}, agreementRevision: 1, requiredAgreementIDs: []string{"terms"}}
	upstream := &accountsRemnawave{
		linkedResponse: accountsFindResponse{remote: remote, exists: true},
		findResponses:  []accountsFindResponse{{exists: true, remote: RemoteUser{Username: username}}},
	}
	linkedUser := model.User{ID: "user-1", TelegramID: telegramID, Username: &username, RemnaUserID: &remoteID, OnboardingState: "agreement"}

	completed, err := newAccountsServiceForTest(repository, &accountsValidator{}, &accountsTelegram{}, upstream, &accountsSettings{}, 1).
		AcceptAgreementRevision(context.Background(), linkedUser, 1, []string{"terms"})
	if err != nil {
		t.Fatalf("AcceptAgreementRevision() error = %v", err)
	}
	if repository.completedRemoteID != remoteID || completed.ID != repository.user.ID {
		t.Fatalf("completed identity = (%q, %+v), want (%q, %q)", repository.completedRemoteID, completed, remoteID, repository.user.ID)
	}
	if upstream.findIndex != 0 {
		t.Fatalf("username lookup count = %d, want 0", upstream.findIndex)
	}
}

func TestAcceptAgreementFailures(t *testing.T) {
	t.Parallel()

	username := "alice"
	telegramID := int64(42)
	otherID := int64(99)
	testError := errors.New("failure")
	validUser := model.User{ID: "user", TelegramID: telegramID, Username: &username, OnboardingState: "agreement"}
	tests := []struct {
		name       string
		accepted   bool
		user       model.User
		repository *accountsRepository
		remote     *accountsRemnawave
		want       error
	}{
		{name: "not accepted", user: validUser, accepted: false, repository: &accountsRepository{}, remote: &accountsRemnawave{}, want: database.ErrConflict},
		{name: "missing username", user: model.User{OnboardingState: "agreement"}, accepted: true, repository: &accountsRepository{}, remote: &accountsRemnawave{}, want: database.ErrConflict},
		{name: "wrong state", user: model.User{Username: &username, OnboardingState: "username"}, accepted: true, repository: &accountsRepository{}, remote: &accountsRemnawave{}, want: database.ErrConflict},
		{name: "find error", user: validUser, accepted: true, repository: &accountsRepository{}, remote: &accountsRemnawave{findResponses: []accountsFindResponse{{err: testError}}}, want: testError},
		{name: "existing without telegram", user: validUser, accepted: true, repository: &accountsRepository{}, remote: &accountsRemnawave{findResponses: []accountsFindResponse{{exists: true}}}, want: ErrUsernameUnavailable},
		{name: "existing other telegram", user: validUser, accepted: true, repository: &accountsRepository{}, remote: &accountsRemnawave{findResponses: []accountsFindResponse{{exists: true, remote: RemoteUser{TelegramID: &otherID}}}}, want: ErrUsernameUnavailable},
		{name: "telegram reconciliation error", user: validUser, accepted: true, repository: &accountsRepository{}, remote: &accountsRemnawave{telegramResponse: accountsFindResponse{err: testError}}, want: testError},
		{name: "telegram username mismatch", user: validUser, accepted: true, repository: &accountsRepository{}, remote: &accountsRemnawave{telegramResponse: accountsFindResponse{exists: true, remote: RemoteUser{Username: "other", TelegramID: &telegramID}}}, want: ErrUsernameUnavailable},
		{name: "create error", user: validUser, accepted: true, repository: &accountsRepository{}, remote: &accountsRemnawave{createErr: testError}, want: testError},
		{name: "duplicate missing", user: validUser, accepted: true, repository: &accountsRepository{}, remote: &accountsRemnawave{findResponses: []accountsFindResponse{{}, {}}, createErr: errAccountsDuplicate}, want: ErrUsernameUnavailable},
		{name: "duplicate refetch error", user: validUser, accepted: true, repository: &accountsRepository{}, remote: &accountsRemnawave{findResponses: []accountsFindResponse{{}, {err: testError}}, createErr: errAccountsDuplicate}, want: testError},
		{name: "complete error", user: validUser, accepted: true, repository: &accountsRepository{completeErr: testError}, remote: &accountsRemnawave{createRemote: RemoteUser{ID: "remote", TelegramID: &telegramID}}, want: testError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.repository.agreementRevision = 1
			test.repository.requiredAgreementIDs = []string{"terms"}
			revision := 1
			if !test.accepted {
				revision = 0
			}
			_, err := newAccountsServiceForTest(test.repository, &accountsValidator{}, &accountsTelegram{}, test.remote, &accountsSettings{}, 1).AcceptAgreementRevision(context.Background(), test.user, revision, []string{"terms"})
			if !errors.Is(err, test.want) {
				t.Fatalf("AcceptAgreementRevision() error = %v, want %v", err, test.want)
			}
		})
	}
}
