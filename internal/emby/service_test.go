package emby

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestHardenPolicyPreservesRemoteAccessAndForcesRestrictions(t *testing.T) {
	t.Parallel()
	current := Policy{
		"EnableRemoteAccess": json.RawMessage(`true`),
		"FutureField":        json.RawMessage(`{"nested":"kept"}`),
		"IsHidden":           json.RawMessage(`false`),
	}
	rating := int32(13)
	hardened := HardenPolicy(current, Preferences{MaxParentalRating: &rating, LibraryIDs: []string{"movies"}})
	checks := map[string]string{
		"EnableRemoteAccess":              "true",
		"FutureField":                     `{"nested":"kept"}`,
		"IsHidden":                        "true",
		"IsHiddenRemotely":                "true",
		"EnableRemoteControlOfOtherUsers": "false",
		"EnableSharedDeviceControl":       "false",
		"EnableAudioPlaybackTranscoding":  "false",
		"EnableVideoPlaybackTranscoding":  "false",
		"EnablePlaybackRemuxing":          "false",
		"EnableSyncTranscoding":           "false",
		"EnableMediaConversion":           "false",
		"EnableContentDownloading":        "false",
		"EnableSubtitleDownloading":       "false",
		"EnableAllFolders":                "false",
		"MaxParentalRating":               "13",
		"EnabledFolders":                  `["movies"]`,
	}
	for key, want := range checks {
		if got := string(hardened[key]); got != want {
			t.Errorf("policy[%s] = %s, want %s", key, got, want)
		}
	}
	if string(current["IsHidden"]) != "false" {
		t.Fatal("HardenPolicy mutated the fetched policy")
	}
}

func TestSetupUsesServerPriceAndProvisioningContext(t *testing.T) {
	t.Parallel()
	repository := &serviceRepository{baseUsername: "ada"}
	remote := newServiceRemote()
	secrets := &serviceSecrets{}
	service := NewService(repository, remote, fixedPrice(275), secrets)
	rating := int32(13)
	account, created, err := service.Setup(context.Background(), "user-1", "private password", Preferences{MaxParentalRating: &rating, LibraryIDs: []string{"movies", "movies"}})
	if err != nil || !created || account.SetupPriceTXBMinor != 275 {
		t.Fatalf("Setup() = (%+v, %v, %v)", account, created, err)
	}
	if secrets.context != "emby.provisioning.password:user-1" || secrets.plaintext != "private password" {
		t.Fatalf("sealed secret = context %q plaintext %q", secrets.context, secrets.plaintext)
	}
	if repository.record.BaseUsername != "ada" || repository.record.SetupPriceTXBMinor != 275 || !reflect.DeepEqual(repository.record.Preferences.LibraryIDs, []string{"movies"}) {
		t.Fatalf("queued record = %+v", repository.record)
	}
}

func TestProvisionCollisionHardeningAndAmbiguousReconciliation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		record            ProvisioningRecord
		seedBaseCollision bool
		seedCandidate     bool
		wantCreate        bool
	}{
		{
			name: "base collision receives stable suffix",
			record: ProvisioningRecord{Account: Account{ID: "account-1", UserID: "user-1", BaseUsername: "ada", Status: StatusQueued,
				Preferences: Preferences{LibraryIDs: []string{"movies"}}}, PasswordCiphertext: "sealed", PasswordContext: "emby.provisioning.password:user-1"},
			seedBaseCollision: true, wantCreate: true,
		},
		{
			name: "ambiguous create is reconciled by persisted exact name",
			record: ProvisioningRecord{Account: Account{ID: "account-2", UserID: "user-2", BaseUsername: "river", CandidateUsername: "river", Status: StatusQueued,
				Preferences: Preferences{LibraryIDs: []string{"movies"}}}, PasswordCiphertext: "sealed", PasswordContext: "emby.provisioning.password:user-2", CreateAttempted: true},
			seedCandidate: true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := &serviceRepository{record: test.record}
			remote := newServiceRemote()
			if test.seedBaseCollision {
				remote.addUser(RemoteUser{ID: "existing", Name: test.record.BaseUsername, Policy: Policy{}})
			}
			if test.seedCandidate {
				remote.addUser(RemoteUser{ID: "reconciled", Name: test.record.CandidateUsername, Policy: basePolicy()})
			}
			service := NewService(repository, remote, fixedPrice(100), &serviceSecrets{})
			if err := service.HandleProvisionJob(context.Background(), test.record.ID); err != nil {
				t.Fatalf("Provision() error = %v", err)
			}
			if repository.record.Status != StatusActive || repository.record.PasswordCiphertext != "" || repository.record.RemoteUserID == "" {
				t.Fatalf("provisioned record = %+v", repository.record)
			}
			if test.wantCreate != (remote.createCalls == 1) {
				t.Fatalf("create calls = %d, wantCreate=%v", remote.createCalls, test.wantCreate)
			}
			if test.seedBaseCollision && repository.record.CandidateUsername != SuffixedUsername("ada", "user-1") {
				t.Fatalf("candidate = %q", repository.record.CandidateUsername)
			}
			if string(remote.updatedPolicy["EnableRemoteAccess"]) != "true" || string(remote.updatedPolicy["EnableAllFolders"]) != "false" ||
				string(remote.updatedPolicy["FutureField"]) != `"kept"` {
				t.Fatalf("updated policy = %#v", remote.updatedPolicy)
			}
			if string(remote.password) != "temporary" {
				t.Fatalf("password = %q", remote.password)
			}
		})
	}
}

func TestLinkedPreferencesAndPasswordNeverPersistPlaintext(t *testing.T) {
	t.Parallel()
	repository := &serviceRepository{record: ProvisioningRecord{Account: Account{
		ID: "account", UserID: "user", BaseUsername: "ada", RemoteUserID: "remote", RemoteUsername: "ada", Status: StatusActive,
	}}}
	remote := newServiceRemote()
	remote.addUser(RemoteUser{ID: "remote", Name: "ada", Policy: basePolicy()})
	service := NewService(repository, remote, fixedPrice(100), &serviceSecrets{})
	rating := int32(13)
	updated, err := service.UpdatePreferences(context.Background(), "user", Preferences{MaxParentalRating: &rating, LibraryIDs: []string{"movies"}})
	if err != nil || updated.Preferences.MaxParentalRating == nil || *updated.Preferences.MaxParentalRating != 13 {
		t.Fatalf("UpdatePreferences() = (%+v, %v)", updated, err)
	}
	if string(remote.updatedPolicy["EnableRemoteAccess"]) != "true" || string(remote.updatedPolicy["EnableContentDownloading"]) != "false" {
		t.Fatalf("restricted policy = %#v", remote.updatedPolicy)
	}
	if err := service.ChangePassword(context.Background(), "user", "current", "replacement"); err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
	if string(remote.currentPassword) != "current" || string(remote.password) != "replacement" || repository.touchCalls != 1 {
		t.Fatalf("password call current=%q next=%q touches=%d", remote.currentPassword, remote.password, repository.touchCalls)
	}
	account, err := service.Account(context.Background(), "user")
	if err != nil || account.ID != "account" {
		t.Fatalf("Account() = (%+v, %v)", account, err)
	}
	accounts, err := service.ListAccounts(context.Background(), 10)
	if err != nil || len(accounts) != 1 || accounts[0].ID != "account" {
		t.Fatalf("ListAccounts() = (%+v, %v)", accounts, err)
	}
	retried, err := service.RetryProvisioning(context.Background(), "account")
	if err != nil || retried.ID != "account" {
		t.Fatalf("RetryProvisioning() = (%+v, %v)", retried, err)
	}
}

func TestVaultAdapterAndPriceFunc(t *testing.T) {
	t.Parallel()
	cipher := &serviceCipher{}
	box := NewSecretBox(cipher)
	sealed, err := box.Seal("emby.provisioning.password:user", []byte("secret"))
	if err != nil || sealed != "encrypted" || cipher.encryptContext != "emby.provisioning.password:user" {
		t.Fatalf("Seal() = (%q, %v), cipher=%+v", sealed, err, cipher)
	}
	opened, err := box.Open("emby.provisioning.password:user", sealed)
	if err != nil || string(opened) != "secret" {
		t.Fatalf("Open() = (%q, %v)", opened, err)
	}
	price, err := PriceFunc(func(context.Context) (int64, error) { return 250, nil }).EmbySetupPriceTXBMinor(context.Background())
	if err != nil || price != 250 {
		t.Fatalf("PriceFunc() = (%d, %v)", price, err)
	}
}

func TestProvisionFailureClassification(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		remoteErr  error
		wantRefund int
		wantRetry  int
		wantError  bool
	}{
		{name: "terminal provider rejection refunds", remoteErr: terminalServiceError, wantRefund: 1},
		{name: "transient outage requeues", remoteErr: transientServiceError, wantRetry: 1, wantError: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := &serviceRepository{record: ProvisioningRecord{Account: Account{ID: "account", UserID: "user", BaseUsername: "ada", Status: StatusQueued},
				PasswordCiphertext: "sealed", PasswordContext: "emby.provisioning.password:user"}}
			remote := newServiceRemote()
			remote.findErr = test.remoteErr
			service := NewService(repository, remote, fixedPrice(100), &serviceSecrets{})
			err := service.Provision(context.Background(), "account")
			if (err != nil) != test.wantError || repository.refundCalls != test.wantRefund || repository.retryCalls != test.wantRetry {
				t.Fatalf("Provision() error=%v refund=%d retry=%d", err, repository.refundCalls, repository.retryCalls)
			}
		})
	}
}

func TestPasswordErrorsCannotPersistOrReturnEchoedPlaintext(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name       string
		cause      error
		wantRefund bool
	}{
		{name: "terminal", cause: fmt.Errorf("%w: rejected temporary", terminalServiceError), wantRefund: true},
		{name: "transient", cause: fmt.Errorf("%w: rejected temporary", transientServiceError)},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := &serviceRepository{record: ProvisioningRecord{Account: Account{
				ID: "account", UserID: "user", BaseUsername: "ada", Status: StatusQueued,
			}, PasswordCiphertext: "sealed", PasswordContext: "emby.provisioning.password:user"}}
			remote := newServiceRemote()
			remote.passwordErr = test.cause
			service := NewService(repository, remote, fixedPrice(100), &serviceSecrets{})
			err := service.Provision(context.Background(), "account")
			if test.wantRefund {
				if err != nil || repository.refundCalls != 1 || repository.retryCalls != 0 {
					t.Fatalf("Provision() error=%v refund=%d retry=%d", err, repository.refundCalls, repository.retryCalls)
				}
			} else if !errors.Is(err, transientServiceError) || repository.retryCalls != 1 || repository.refundCalls != 0 {
				t.Fatalf("Provision() error=%v refund=%d retry=%d", err, repository.refundCalls, repository.retryCalls)
			}
			persisted := repository.refundReason
			if repository.retryError != nil {
				persisted += " " + repository.retryError.Error()
			}
			if strings.Contains(persisted, "temporary") || (err != nil && strings.Contains(err.Error(), "temporary")) {
				t.Fatalf("password leaked through provisioning error: returned=%q persisted=%q", err, persisted)
			}
		})
	}
}

func TestProvisionRecoveryAndTerminalStateBranches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		record      ProvisioningRecord
		seed        *RemoteUser
		secretError error
		wantRefund  int
		wantActive  bool
	}{
		{
			name: "stored remote id resumes policy work",
			record: ProvisioningRecord{Account: Account{ID: "one", UserID: "user-one", BaseUsername: "ada", RemoteUserID: "remote-one", CandidateUsername: "ada", Status: StatusQueued},
				PasswordCiphertext: "sealed", PasswordContext: "emby.provisioning.password:user-one"},
			seed: &RemoteUser{ID: "remote-one", Name: "ada", Policy: basePolicy()}, wantActive: true,
		},
		{
			name: "restored server reconciles changed remote id by candidate",
			record: ProvisioningRecord{Account: Account{ID: "two", UserID: "user-two", BaseUsername: "river", RemoteUserID: "old-id", CandidateUsername: "river", Status: StatusQueued},
				PasswordCiphertext: "sealed", PasswordContext: "emby.provisioning.password:user-two"},
			seed: &RemoteUser{ID: "new-id", Name: "river", Policy: basePolicy()}, wantActive: true,
		},
		{
			name: "missing secret refunds", record: ProvisioningRecord{Account: Account{ID: "three", UserID: "user-three", Status: StatusQueued}}, wantRefund: 1,
		},
		{
			name:        "undecryptable secret refunds",
			record:      ProvisioningRecord{Account: Account{ID: "four", UserID: "user-four", Status: StatusQueued}, PasswordCiphertext: "sealed", PasswordContext: "emby.provisioning.password:user-four"},
			secretError: errors.New("authentication failed"), wantRefund: 1,
		},
		{name: "active replay is complete", record: ProvisioningRecord{Account: Account{ID: "five", UserID: "user-five", Status: StatusActive}}, wantActive: true},
		{name: "failed replay is complete", record: ProvisioningRecord{Account: Account{ID: "six", UserID: "user-six", Status: StatusFailed}}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := &serviceRepository{record: test.record}
			remote := newServiceRemote()
			if test.seed != nil {
				remote.addUser(*test.seed)
			}
			secrets := &serviceSecrets{openErr: test.secretError}
			service := NewService(repository, remote, fixedPrice(100), secrets)
			if err := service.Provision(context.Background(), test.record.ID); err != nil {
				t.Fatalf("Provision() error = %v", err)
			}
			if repository.refundCalls != test.wantRefund {
				t.Fatalf("refund calls = %d, want %d", repository.refundCalls, test.wantRefund)
			}
			if test.wantActive && test.record.Status == StatusQueued && repository.record.Status != StatusActive {
				t.Fatalf("status = %s, want active", repository.record.Status)
			}
		})
	}
}

func TestProvisionReconcilesCommittedCreateTimeout(t *testing.T) {
	t.Parallel()
	repository := &serviceRepository{record: ProvisioningRecord{Account: Account{ID: "account", UserID: "user", BaseUsername: "ada", Status: StatusQueued},
		PasswordCiphertext: "sealed", PasswordContext: "emby.provisioning.password:user"}}
	remote := newServiceRemote()
	remote.createErr, remote.createCommits = transientServiceError, true
	service := NewService(repository, remote, fixedPrice(100), &serviceSecrets{})
	if err := service.Provision(context.Background(), "account"); err != nil {
		t.Fatalf("Provision() error = %v", err)
	}
	if repository.record.Status != StatusActive || remote.createCalls != 1 {
		t.Fatalf("record=%+v createCalls=%d", repository.record, remote.createCalls)
	}
}

func TestSetupAndLinkedValidationBranches(t *testing.T) {
	t.Parallel()
	activeRepository := &serviceRepository{record: ProvisioningRecord{Account: Account{ID: "existing", UserID: "user", Status: StatusActive}}}
	activeRemote := newServiceRemote()
	service := NewService(activeRepository, activeRemote, fixedPrice(100), &serviceSecrets{})
	if account, created, err := service.Setup(context.Background(), "user", "password", Preferences{}); err != nil || created || account.ID != "existing" {
		t.Fatalf("Setup(existing) = (%+v, %v, %v)", account, created, err)
	}

	validationRepository := &serviceRepository{baseUsername: "ada"}
	validationService := NewService(validationRepository, newServiceRemote(), fixedPrice(100), &serviceSecrets{})
	tests := []struct {
		name        string
		password    string
		preferences Preferences
	}{
		{name: "empty password"},
		{name: "unknown library", password: "password", preferences: Preferences{LibraryIDs: []string{"missing"}}},
		{name: "unknown rating", password: "password", preferences: Preferences{MaxParentalRating: int32Pointer(99)}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if _, _, err := validationService.Setup(context.Background(), "user", test.password, test.preferences); err == nil {
				t.Fatal("Setup() error = nil")
			}
		})
	}

	inactiveRepository := &serviceRepository{record: ProvisioningRecord{Account: Account{ID: "inactive", UserID: "user", Status: StatusQueued}}}
	inactiveService := NewService(inactiveRepository, newServiceRemote(), fixedPrice(100), &serviceSecrets{})
	if _, err := inactiveService.UpdatePreferences(context.Background(), "user", Preferences{}); !errors.Is(err, ErrRemoteAccountMissing) {
		t.Fatalf("UpdatePreferences(inactive) = %v", err)
	}
	if err := inactiveService.ChangePassword(context.Background(), "user", "", "new"); !errors.Is(err, ErrRemoteAccountMissing) {
		t.Fatalf("ChangePassword(inactive) = %v", err)
	}
	if err := inactiveService.ChangePassword(context.Background(), "user", "", ""); !errors.Is(err, ErrInvalidSetup) {
		t.Fatalf("ChangePassword(invalid) = %v", err)
	}
}

var (
	terminalServiceError  = errors.New("terminal provider error")
	transientServiceError = errors.New("transient provider error")
	missingServiceError   = errors.New("provider not found")
)

type fixedPrice int64

func (p fixedPrice) EmbySetupPriceTXBMinor(context.Context) (int64, error) { return int64(p), nil }

type serviceSecrets struct {
	context   string
	plaintext string
	openErr   error
}

func (s *serviceSecrets) Seal(context string, plaintext []byte) (string, error) {
	s.context, s.plaintext = context, string(plaintext)
	return "sealed", nil
}
func (s *serviceSecrets) Open(string, string) ([]byte, error) {
	if s.openErr != nil {
		return nil, s.openErr
	}
	return []byte("temporary"), nil
}

type serviceRepository struct {
	baseUsername string
	record       ProvisioningRecord
	refundCalls  int
	retryCalls   int
	touchCalls   int
	retryError   error
	refundReason string
}

func (r *serviceRepository) EmbyBaseUsername(context.Context, string) (string, error) {
	return r.baseUsername, nil
}
func (r *serviceRepository) QueueEmbySetup(_ context.Context, input QueueSetupInput, now time.Time) (Account, bool, error) {
	r.record = ProvisioningRecord{Account: Account{ID: input.ID, UserID: input.UserID, BaseUsername: input.BaseUsername,
		Status: StatusQueued, SetupPriceTXBMinor: input.SetupPriceTXBMinor, SetupAttempt: 1, Preferences: input.Preferences,
		CreatedAt: now, UpdatedAt: now}, PasswordCiphertext: input.PasswordCiphertext, PasswordContext: input.PasswordContext}
	return r.record.Account, true, nil
}
func (r *serviceRepository) EmbyAccountForUser(context.Context, string) (Account, error) {
	if r.record.ID == "" {
		return Account{}, ErrNotFound
	}
	return r.record.Account, nil
}
func (r *serviceRepository) ListEmbyAccounts(context.Context, int) ([]Account, error) {
	if r.record.ID == "" {
		return []Account{}, nil
	}
	return []Account{r.record.Account}, nil
}
func (r *serviceRepository) EmbyProvisioningByID(context.Context, string) (ProvisioningRecord, error) {
	return r.record, nil
}
func (r *serviceRepository) RetryEmbyProvisioning(context.Context, string, time.Time) (Account, error) {
	return r.record.Account, nil
}
func (r *serviceRepository) BeginEmbyProvisioning(context.Context, string, time.Time) (ProvisioningRecord, error) {
	if r.record.Status == StatusQueued {
		r.record.Status = StatusProvisioning
	}
	return r.record, nil
}
func (r *serviceRepository) SetEmbyCandidate(_ context.Context, _ string, candidate string, _ time.Time) error {
	r.record.CandidateUsername = candidate
	return nil
}
func (r *serviceRepository) MarkEmbyCreateAttempted(context.Context, string, time.Time) error {
	r.record.CreateAttempted = true
	return nil
}
func (r *serviceRepository) SetEmbyRemoteIdentity(_ context.Context, _ string, id, name string, _ time.Time) error {
	r.record.RemoteUserID, r.record.RemoteUsername, r.record.CandidateUsername = id, name, name
	return nil
}
func (r *serviceRepository) RequeueEmbyProvisioning(_ context.Context, _ string, provisionErr error, _ time.Time) error {
	r.retryCalls++
	r.retryError = provisionErr
	r.record.Status = StatusQueued
	return nil
}
func (r *serviceRepository) MarkEmbyProvisioned(context.Context, string, time.Time) error {
	r.record.Status, r.record.PasswordCiphertext, r.record.PasswordContext = StatusActive, "", ""
	return nil
}
func (r *serviceRepository) FailAndRefundEmbySetup(_ context.Context, _ string, reason string, _ time.Time) (Account, error) {
	r.refundCalls++
	r.refundReason = reason
	r.record.Status, r.record.PasswordCiphertext, r.record.PasswordContext = StatusFailed, "", ""
	return r.record.Account, nil
}
func (r *serviceRepository) UpdateEmbyPreferences(_ context.Context, _ string, preferences Preferences, _ time.Time) (Account, error) {
	r.record.Preferences = preferences
	return r.record.Account, nil
}
func (r *serviceRepository) TouchEmbyAccount(context.Context, string, time.Time) error {
	r.touchCalls++
	return nil
}

type serviceRemote struct {
	byName          map[string]RemoteUser
	byID            map[string]RemoteUser
	findErr         error
	createCalls     int
	createdName     string
	createErr       error
	createCommits   bool
	currentPassword []byte
	password        []byte
	passwordErr     error
	updatedPolicy   Policy
}

func newServiceRemote() *serviceRemote {
	return &serviceRemote{byName: make(map[string]RemoteUser), byID: make(map[string]RemoteUser)}
}
func (r *serviceRemote) addUser(user RemoteUser) {
	r.byName[user.Name], r.byID[user.ID] = user, user
}
func (r *serviceRemote) FindUserByName(_ context.Context, name string) (RemoteUser, bool, error) {
	if r.findErr != nil {
		return RemoteUser{}, false, r.findErr
	}
	user, exists := r.byName[name]
	return user, exists, nil
}
func (r *serviceRemote) CreateUser(_ context.Context, name string) (RemoteUser, error) {
	r.createCalls++
	r.createdName = name
	user := RemoteUser{ID: "created", Name: name, Policy: basePolicy()}
	if r.createErr == nil || r.createCommits {
		r.addUser(user)
	}
	if r.createErr != nil {
		return RemoteUser{}, r.createErr
	}
	return user, nil
}
func (r *serviceRemote) GetUser(_ context.Context, id string) (RemoteUser, error) {
	user, exists := r.byID[id]
	if !exists {
		return RemoteUser{}, missingServiceError
	}
	return user, nil
}

func (r *serviceRemote) SetPassword(_ context.Context, _ string, current, password []byte) error {
	r.currentPassword = append([]byte(nil), current...)
	r.password = append([]byte(nil), password...)
	return r.passwordErr
}
func (r *serviceRemote) UpdatePolicy(_ context.Context, _ string, policy Policy) error {
	r.updatedPolicy = policy.Clone()
	return nil
}
func (*serviceRemote) ListSelectableFolders(context.Context) ([]Folder, error) {
	return []Folder{{ID: "movies", Name: "Movies"}}, nil
}
func (*serviceRemote) ListParentalRatings(context.Context) ([]ParentalRating, error) {
	return []ParentalRating{{Name: "PG-13", Value: 13}}, nil
}
func (*serviceRemote) IsNotFound(err error) bool { return errors.Is(err, missingServiceError) }
func (*serviceRemote) IsTerminal(err error) bool { return errors.Is(err, terminalServiceError) }

func basePolicy() Policy {
	return Policy{"EnableRemoteAccess": json.RawMessage(`true`), "FutureField": json.RawMessage(`"kept"`)}
}

func int32Pointer(value int32) *int32 { return &value }

type serviceCipher struct {
	encryptContext string
	plaintext      string
}

func (c *serviceCipher) Encrypt(context, plaintext string) (string, error) {
	c.encryptContext, c.plaintext = context, plaintext
	return "encrypted", nil
}

func (c *serviceCipher) Decrypt(context, ciphertext string) (string, error) {
	if context != c.encryptContext || ciphertext != "encrypted" {
		return "", errors.New("cipher context mismatch")
	}
	return c.plaintext, nil
}
