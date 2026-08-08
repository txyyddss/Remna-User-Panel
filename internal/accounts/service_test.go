package accounts

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

func TestAuthenticateAndSessionLookup(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	profile := model.TelegramProfile{ID: 42, FirstName: "Ada"}
	repository := &accountsRepository{user: model.User{ID: "user-1", TelegramID: profile.ID}}
	validator := &accountsValidator{profile: profile}
	service := newAccountsServiceForTest(repository, validator, &accountsTelegram{}, &accountsRemnawave{}, &accountsSettings{}, profile.ID)

	user, token, expiresAt, err := service.Authenticate(context.Background(), "signed-init-data")
	if err != nil {
		t.Fatalf("Authenticate(): %v", err)
	}
	if user.ID != "user-1" || token == "" || !expiresAt.Equal(now.Add(7*24*time.Hour)) || !repository.upsertAdmin {
		t.Fatalf("Authenticate() = user %+v, token length %d, expiry %s, admin %t", user, len(token), expiresAt, repository.upsertAdmin)
	}
	wantHash := sha256.Sum256([]byte(token))
	if string(repository.sessionHash) != string(wantHash[:]) || repository.sessionUserID != user.ID || !repository.sessionExpires.Equal(expiresAt) {
		t.Fatal("session was not stored as the expected token hash")
	}

	repository.sessionUser = user
	got, err := service.UserBySession(context.Background(), token)
	if err != nil || got.ID != user.ID || string(repository.lookupSessionHash) != string(wantHash[:]) {
		t.Fatalf("UserBySession() = (%+v, %v)", got, err)
	}
}

func TestAuthenticationFailures(t *testing.T) {
	t.Parallel()

	testError := errors.New("failure")
	tests := []struct {
		name      string
		configure func(*accountsRepository, *accountsValidator)
		want      error
	}{
		{name: "invalid init data", configure: func(_ *accountsRepository, validator *accountsValidator) { validator.err = testError }, want: ErrInvalidAuthentication},
		{name: "upsert", configure: func(repository *accountsRepository, _ *accountsValidator) { repository.upsertErr = testError }, want: testError},
		{name: "session persistence", configure: func(repository *accountsRepository, _ *accountsValidator) { repository.createSessionErr = testError }, want: testError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := &accountsRepository{user: model.User{ID: "user"}}
			validator := &accountsValidator{profile: model.TelegramProfile{ID: 1}}
			test.configure(repository, validator)
			_, _, _, err := newAccountsServiceForTest(repository, validator, &accountsTelegram{}, &accountsRemnawave{}, &accountsSettings{}, 99).Authenticate(context.Background(), "raw")
			if !errors.Is(err, test.want) {
				t.Fatalf("Authenticate() error = %v, want %v", err, test.want)
			}
		})
	}

	service := newAccountsServiceForTest(&accountsRepository{sessionLookupErr: testError}, &accountsValidator{}, &accountsTelegram{}, &accountsRemnawave{}, &accountsSettings{}, 1)
	if _, err := service.UserBySession(context.Background(), ""); !errors.Is(err, ErrInvalidAuthentication) {
		t.Fatalf("empty UserBySession() error = %v", err)
	}
	if _, err := service.UserBySession(context.Background(), "token"); !errors.Is(err, ErrInvalidAuthentication) {
		t.Fatalf("failed UserBySession() error = %v", err)
	}
}

func TestAuthenticateRechecksLinkedRemnawaveUser(t *testing.T) {
	t.Parallel()

	remoteID := "remote-1"
	profile := model.TelegramProfile{ID: 42, FirstName: "Ada"}
	upstreamFailure := errors.New("upstream 503")
	tests := []struct {
		name         string
		linked       accountsFindResponse
		wantRecovery bool
		wantUpstream bool
		wantSession  bool
	}{
		{name: "linked user still exists", linked: accountsFindResponse{exists: true}, wantSession: true},
		{name: "confirmed missing starts recovery", linked: accountsFindResponse{}, wantRecovery: true, wantSession: true},
		{name: "temporary failure blocks authentication", linked: accountsFindResponse{err: upstreamFailure}, wantUpstream: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			user := model.User{ID: "user-1", TelegramID: profile.ID, OnboardingState: "complete", RemnaUserID: &remoteID}
			repository := &accountsRepository{user: user, recoveryUser: model.User{
				ID: user.ID, TelegramID: user.TelegramID, OnboardingState: "membership", RecoveryReason: "remnawave_user_missing",
			}}
			remote := &accountsRemnawave{linkedResponse: test.linked}
			_, token, _, err := newAccountsServiceForTest(repository, &accountsValidator{profile: profile}, &accountsTelegram{}, remote, &accountsSettings{}, 99).Authenticate(context.Background(), "signed")
			if test.wantUpstream {
				if !errors.Is(err, ErrUpstreamUnavailable) || token != "" || repository.sessionUserID != "" {
					t.Fatalf("Authenticate() = token %q, err %v, stored session %q", token, err, repository.sessionUserID)
				}
				return
			}
			if err != nil || (token != "") != test.wantSession {
				t.Fatalf("Authenticate() = token %q, err %v", token, err)
			}
			if repository.recoveryStarted != test.wantRecovery {
				t.Fatalf("recovery started = %t, want %t", repository.recoveryStarted, test.wantRecovery)
			}
			if test.wantRecovery && repository.recoveryReason != "remnawave_user_missing" {
				t.Fatalf("recovery reason = %q", repository.recoveryReason)
			}
		})
	}
}

func TestCreateInvites(t *testing.T) {
	t.Parallel()

	settings := &accountsSettings{values: map[string]string{
		"telegram.group_chat_id":   "-1001",
		"telegram.channel_chat_id": "-1002",
	}}
	repository := &accountsRepository{}
	telegram := &accountsTelegram{}
	service := newAccountsServiceForTest(repository, &accountsValidator{}, telegram, &accountsRemnawave{}, settings, 1)
	user := model.User{ID: strings.Repeat("x", 40), OnboardingState: "intro"}

	invites, expiresAt, err := service.CreateInvites(context.Background(), user)
	if err != nil {
		t.Fatalf("CreateInvites(): %v", err)
	}
	if invites["group"] == "" || invites["channel"] == "" || !repository.advanced || len(repository.savedInvites) != 2 {
		t.Fatalf("CreateInvites() = %+v, advanced %t, saved %+v", invites, repository.advanced, repository.savedInvites)
	}
	if len(telegram.inviteNames[0]) != 32 || !expiresAt.Equal(service.now().UTC().Add(30*time.Minute)) {
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
		{name: "persistence revokes provider invite", user: model.User{ID: "user"}, settings: validAccountsSettings(), repository: &accountsRepository{saveInviteErr: testError}, telegram: &accountsTelegram{}, wantRevoke: true},
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

func TestHandleJoinRequest(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	validInvite := database.JoinInvite{ID: "invite-1", UserID: "user-1", ChatID: -1001, InviteLink: "invite", ExpiresAt: now.Add(time.Hour)}
	tests := []struct {
		name       string
		configure  func(*accountsRepository, *accountsTelegram)
		telegramID int64
		chatID     int64
		want       error
	}{
		{name: "success", telegramID: 42, chatID: -1001},
		{name: "invite lookup", telegramID: 42, chatID: -1001, configure: func(repository *accountsRepository, _ *accountsTelegram) {
			repository.inviteLookupErr = errors.New("lookup")
		}},
		{name: "user lookup", telegramID: 42, chatID: -1001, configure: func(repository *accountsRepository, _ *accountsTelegram) {
			repository.userByIDErr = errors.New("lookup")
		}},
		{name: "wrong identity", telegramID: 99, chatID: -1001, want: ErrInvalidAuthentication},
		{name: "wrong chat", telegramID: 42, chatID: -1002, want: ErrInvalidAuthentication},
		{name: "approved", telegramID: 42, chatID: -1001, configure: func(repository *accountsRepository, _ *accountsTelegram) {
			at := now
			repository.invite.ApprovedAt = &at
		}, want: ErrInvalidAuthentication},
		{name: "revoked", telegramID: 42, chatID: -1001, configure: func(repository *accountsRepository, _ *accountsTelegram) {
			at := now
			repository.invite.RevokedAt = &at
		}, want: ErrInvalidAuthentication},
		{name: "expired", telegramID: 42, chatID: -1001, configure: func(repository *accountsRepository, _ *accountsTelegram) { repository.invite.ExpiresAt = now }, want: ErrInvalidAuthentication},
		{name: "approve error", telegramID: 42, chatID: -1001, configure: func(_ *accountsRepository, telegram *accountsTelegram) { telegram.approveErr = errors.New("approve") }},
		{name: "already joined after approve error", telegramID: 42, chatID: -1001, configure: func(_ *accountsRepository, telegram *accountsTelegram) {
			telegram.approveErr = errors.New("approve")
			telegram.memberships = map[string]bool{"-1001": true}
		}},
		{name: "revoke error", telegramID: 42, chatID: -1001, configure: func(_ *accountsRepository, telegram *accountsTelegram) { telegram.revokeErr = errors.New("revoke") }},
		{name: "mark error", telegramID: 42, chatID: -1001, configure: func(repository *accountsRepository, _ *accountsTelegram) {
			repository.markInviteErr = errors.New("mark")
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := &accountsRepository{user: model.User{ID: "user-1", TelegramID: 42}, invite: validInvite}
			telegram := &accountsTelegram{}
			if test.configure != nil {
				test.configure(repository, telegram)
			}
			err := newAccountsServiceForTest(repository, &accountsValidator{}, telegram, &accountsRemnawave{}, &accountsSettings{}, 1).HandleJoinRequest(context.Background(), test.telegramID, test.chatID, "invite")
			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("HandleJoinRequest() error = %v, want %v", err, test.want)
			}
			if test.want == nil && (test.name == "success" || test.name == "already joined after approve error") && (err != nil || !repository.inviteMarked || telegram.approveCalls != 1 || len(telegram.revokedLinks) != 1) {
				t.Fatalf("successful HandleJoinRequest() = err %v, marked %t, approve %d, revoke %d", err, repository.inviteMarked, telegram.approveCalls, len(telegram.revokedLinks))
			}
			if test.want == nil && test.name != "success" && test.name != "already joined after approve error" && err == nil {
				t.Fatal("HandleJoinRequest() unexpectedly succeeded")
			}
		})
	}
}

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
			if test.want == nil && test.name == "success trims" {
				if err != nil || test.repository.reservedUsername != "alice" {
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
	remote := RemoteUser{ID: "remote-1", Username: username, TelegramID: &telegramID, SubscriptionURL: "https://subscription.test"}
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
			repository := &accountsRepository{user: model.User{ID: "user-1", OnboardingState: "complete"}}
			service := newAccountsServiceForTest(repository, &accountsValidator{}, &accountsTelegram{}, test.remote, &accountsSettings{}, 1)
			user := model.User{ID: "user-1", TelegramID: telegramID, Username: &username, OnboardingState: "agreement"}
			completed, err := service.AcceptAgreement(context.Background(), user, true)
			if err != nil || completed.OnboardingState != "complete" {
				t.Fatalf("AcceptAgreement() = (%+v, %v)", completed, err)
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
			if repository.completedRemoteID != remote.ID || repository.completedSubscription != remote.SubscriptionURL {
				t.Fatalf("completion remote values = %q/%q", repository.completedRemoteID, repository.completedSubscription)
			}
		})
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
			_, err := newAccountsServiceForTest(test.repository, &accountsValidator{}, &accountsTelegram{}, test.remote, &accountsSettings{}, 1).AcceptAgreement(context.Background(), test.user, test.accepted)
			if !errors.Is(err, test.want) {
				t.Fatalf("AcceptAgreement() error = %v, want %v", err, test.want)
			}
		})
	}
}

type accountsRepository struct {
	user                  model.User
	sessionUser           model.User
	invite                database.JoinInvite
	upsertErr             error
	createSessionErr      error
	sessionLookupErr      error
	advanceErr            error
	saveInviteErr         error
	inviteLookupErr       error
	userByIDErr           error
	markInviteErr         error
	reserveErr            error
	completeErr           error
	userByTelegramErr     error
	upsertAdmin           bool
	sessionHash           []byte
	sessionUserID         string
	sessionExpires        time.Time
	lookupSessionHash     []byte
	advanced              bool
	savedInvites          []database.JoinInvite
	groupJoined           bool
	channelJoined         bool
	inviteMarked          bool
	reservedUsername      string
	completedRemoteID     string
	completedSubscription string
	recoveryUser          model.User
	recoveryStarted       bool
	recoveryReason        string
}

func (r *accountsRepository) UpsertTelegramUser(_ context.Context, _ model.TelegramProfile, admin bool) (model.User, bool, error) {
	r.upsertAdmin = admin
	return r.user, true, r.upsertErr
}
func (r *accountsRepository) CreateSession(_ context.Context, hash []byte, userID string, expires time.Time) error {
	r.sessionHash, r.sessionUserID, r.sessionExpires = append([]byte(nil), hash...), userID, expires
	return r.createSessionErr
}
func (r *accountsRepository) UserBySession(_ context.Context, hash []byte, _ time.Time) (model.User, error) {
	r.lookupSessionHash = append([]byte(nil), hash...)
	return r.sessionUser, r.sessionLookupErr
}
func (r *accountsRepository) UserByID(context.Context, string) (model.User, error) {
	return r.user, r.userByIDErr
}
func (r *accountsRepository) UserByTelegramID(context.Context, int64) (model.User, error) {
	return r.user, r.userByTelegramErr
}
func (r *accountsRepository) AdvanceToMembership(context.Context, string) error {
	r.advanced = true
	return r.advanceErr
}
func (r *accountsRepository) UpdateMembership(_ context.Context, _ string, group, channel bool) (model.User, error) {
	r.groupJoined, r.channelJoined = group, channel
	return r.user, nil
}
func (r *accountsRepository) ReserveUsername(_ context.Context, _ string, username string) error {
	r.reservedUsername = username
	return r.reserveErr
}
func (r *accountsRepository) CompleteOnboarding(_ context.Context, _ string, remoteID, subscription string, _ time.Time) (model.User, error) {
	r.completedRemoteID, r.completedSubscription = remoteID, subscription
	return r.user, r.completeErr
}
func (r *accountsRepository) BeginRemnawaveRecovery(_ context.Context, _ string, reason string, _ time.Time) (model.User, error) {
	r.recoveryStarted, r.recoveryReason = true, reason
	return r.recoveryUser, nil
}
func (r *accountsRepository) SaveJoinInvite(_ context.Context, userID, kind string, chatID int64, link string, expires time.Time) (database.JoinInvite, error) {
	invite := database.JoinInvite{ID: "saved-" + kind, UserID: userID, ChatKind: kind, ChatID: chatID, InviteLink: link, ExpiresAt: expires}
	r.savedInvites = append(r.savedInvites, invite)
	return invite, r.saveInviteErr
}
func (r *accountsRepository) JoinInviteByLink(context.Context, string) (database.JoinInvite, error) {
	return r.invite, r.inviteLookupErr
}
func (r *accountsRepository) MarkJoinInviteUsed(context.Context, string, time.Time) error {
	r.inviteMarked = true
	return r.markInviteErr
}

type accountsValidator struct {
	profile model.TelegramProfile
	err     error
}

func (v *accountsValidator) Validate(string) (model.TelegramProfile, error) { return v.profile, v.err }

type accountsTelegram struct {
	memberships   map[string]bool
	createErr     error
	membershipErr error
	approveErr    error
	revokeErr     error
	inviteNames   []string
	revokedLinks  []string
	approveCalls  int
}

func (t *accountsTelegram) CreateJoinRequestInvite(_ context.Context, chatID, name string, _ time.Time) (string, error) {
	t.inviteNames = append(t.inviteNames, name)
	return "https://invite.test/" + chatID, t.createErr
}
func (t *accountsTelegram) GetMembership(_ context.Context, chatID string, _ int64) (bool, error) {
	return t.memberships[chatID], t.membershipErr
}
func (t *accountsTelegram) ApproveJoinRequest(context.Context, string, int64) error {
	t.approveCalls++
	return t.approveErr
}
func (t *accountsTelegram) RevokeInviteLink(_ context.Context, _ string, link string) error {
	t.revokedLinks = append(t.revokedLinks, link)
	return t.revokeErr
}

type accountsFindResponse struct {
	remote RemoteUser
	exists bool
	err    error
}

var errAccountsDuplicate = errors.New("duplicate")

type accountsRemnawave struct {
	findResponses    []accountsFindResponse
	findIndex        int
	telegramResponse accountsFindResponse
	createRemote     RemoteUser
	createErr        error
	created          []RemoteCreateUser
	linkedResponse   accountsFindResponse
}

func (r *accountsRemnawave) FindUserByUsername(context.Context, string) (RemoteUser, bool, error) {
	if r.findIndex >= len(r.findResponses) {
		return RemoteUser{}, false, nil
	}
	response := r.findResponses[r.findIndex]
	r.findIndex++
	return response.remote, response.exists, response.err
}
func (r *accountsRemnawave) FindUserByTelegramID(context.Context, int64) (RemoteUser, bool, error) {
	return r.telegramResponse.remote, r.telegramResponse.exists, r.telegramResponse.err
}
func (r *accountsRemnawave) FindUserByID(context.Context, string) (RemoteUser, bool, error) {
	return r.linkedResponse.remote, r.linkedResponse.exists, r.linkedResponse.err
}
func (r *accountsRemnawave) CreateUser(_ context.Context, input RemoteCreateUser) (RemoteUser, error) {
	r.created = append(r.created, input)
	return r.createRemote, r.createErr
}
func (r *accountsRemnawave) IsDuplicateError(err error) bool {
	return errors.Is(err, errAccountsDuplicate)
}

type accountsSettings struct {
	values map[string]string
	errs   map[string]error
}

func (s *accountsSettings) Plaintext(_ context.Context, key string) (string, error) {
	if s.errs != nil && s.errs[key] != nil {
		return "", s.errs[key]
	}
	return s.values[key], nil
}

func validAccountsSettings() *accountsSettings {
	return &accountsSettings{values: map[string]string{"telegram.group_chat_id": "-1001", "telegram.channel_chat_id": "-1002"}}
}

func newAccountsServiceForTest(repository Repository, validator InitDataValidator, telegram TelegramClient, remnawave RemnawaveClient, settings Settings, adminID int64) *Service {
	service := NewService(repository, validator, telegram, remnawave, settings, adminID, 7*24*time.Hour)
	service.now = func() time.Time { return time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC) }
	return service
}
