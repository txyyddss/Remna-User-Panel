package entitlements

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

func TestWorkerDrainAppliesEntitlementAndCompletesJob(t *testing.T) {
	t.Parallel()

	remoteID := "remote-1"
	job := model.OutboxJob{ID: "job-1", Kind: "remna_apply_entitlement", Payload: `{"purchaseId":"purchase-1"}`, Attempts: 1}
	repository := &entitlementRepository{
		jobs:     []*model.OutboxJob{&job},
		purchase: model.Purchase{ID: "purchase-1", Status: "activating", TrafficLimitBytes: 1234, ResetStrategy: "MONTH", SquadUUIDs: []string{"squad-b", "squad-a"}, ValidUntil: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)},
		user:     model.User{ID: "user-1", RemnaUserID: &remoteID},
	}
	remote := &entitlementRemnawave{}
	worker := newEntitlementWorkerForTest(repository, remote)

	if err := worker.Drain(context.Background(), 1); err != nil {
		t.Fatalf("Drain(): %v", err)
	}
	if remote.applyUserID != remoteID || remote.applyLimit != 1234 || remote.applyResetStrategy != "MONTH" || remote.removeCalls != 1 || remote.resetCalls != 1 {
		t.Fatalf("ApplyEntitlement call = %+v", remote)
	}
	if len(remote.applySquads) != 2 || repository.markedPurchaseID != "purchase-1" || repository.markedSuccess == nil || !*repository.markedSuccess {
		t.Fatalf("apply squads/mark = %v, %q/%v", remote.applySquads, repository.markedPurchaseID, repository.markedSuccess)
	}
	if repository.completedJobID != "job-1" || repository.completedAttempts != 1 || repository.completedErr != nil {
		t.Fatalf("completion = %q/%d/%v", repository.completedJobID, repository.completedAttempts, repository.completedErr)
	}
}

func TestWorkerDrainRetriesAndMarksTerminalFailure(t *testing.T) {
	t.Parallel()

	remoteID := "remote-1"
	job := model.OutboxJob{ID: "job-1", Kind: "remna_apply_entitlement", Payload: `{"purchaseId":"purchase-1"}`, Attempts: 10}
	repository := &entitlementRepository{
		jobs:     []*model.OutboxJob{&job},
		purchase: model.Purchase{ID: "purchase-1", Status: "active", ValidUntil: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)},
		user:     model.User{ID: "user-1", RemnaUserID: &remoteID},
	}
	remote := &entitlementRemnawave{applyErr: errors.New("upstream unavailable")}
	worker := newEntitlementWorkerForTest(repository, remote)

	if err := worker.Drain(context.Background(), 1); err != nil {
		t.Fatalf("Drain(): %v", err)
	}
	if repository.completedErr == nil {
		t.Fatal("failed job was completed without its processing error")
	}
	if repository.markedSuccess == nil || *repository.markedSuccess {
		t.Fatalf("terminal purchase success marker = %v, want false", repository.markedSuccess)
	}
}

func TestWorkerRetryDoesNotRedispatchTrafficReset(t *testing.T) {
	t.Parallel()

	remoteID := "remote-1"
	repository := &entitlementRepository{
		purchase:   model.Purchase{ID: "purchase-1", Status: "activating", ValidUntil: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)},
		user:       model.User{ID: "user-1", RemnaUserID: &remoteID},
		resetPhase: "reset",
	}
	remote := &entitlementRemnawave{}
	if err := newEntitlementWorkerForTest(repository, remote).process(context.Background(), model.OutboxJob{Kind: "remna_apply_entitlement", Payload: `{"purchaseId":"purchase-1"}`, Attempts: 2}); err != nil {
		t.Fatalf("process(): %v", err)
	}
	if remote.applyCalls != 1 || remote.removeCalls != 0 || remote.resetCalls != 0 {
		t.Fatalf("retry calls apply/remove/reset = %d/%d/%d, want 1/0/0", remote.applyCalls, remote.removeCalls, remote.resetCalls)
	}
}

func TestWorkerResumesDurableTrafficResetPhases(t *testing.T) {
	t.Parallel()

	remoteID := "remote-1"
	future := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		phase      string
		wantRemove int
		wantReset  int
		wantApply  int
	}{
		{phase: "pending", wantRemove: 1, wantReset: 1, wantApply: 1},
		{phase: "quiesced", wantReset: 1, wantApply: 1},
		{phase: "reset", wantApply: 1},
	} {
		t.Run(test.phase, func(t *testing.T) {
			t.Parallel()
			repository := &entitlementRepository{purchase: model.Purchase{ID: "purchase", Status: "activating", ValidUntil: future}, user: model.User{RemnaUserID: &remoteID}, resetPhase: test.phase}
			remote := &entitlementRemnawave{}
			if err := newEntitlementWorkerForTest(repository, remote).process(context.Background(), model.OutboxJob{Kind: "remna_apply_entitlement", Payload: `{"purchaseId":"purchase"}`}); err != nil {
				t.Fatalf("process(): %v", err)
			}
			if remote.removeCalls != test.wantRemove || remote.resetCalls != test.wantReset || remote.applyCalls != test.wantApply {
				t.Fatalf("calls remove/reset/apply = %d/%d/%d, want %d/%d/%d", remote.removeCalls, remote.resetCalls, remote.applyCalls, test.wantRemove, test.wantReset, test.wantApply)
			}
		})
	}
}

func TestWorkerDrainInfrastructureAndContextErrors(t *testing.T) {
	t.Parallel()

	testError := errors.New("infrastructure failure")
	tests := []struct {
		name      string
		limit     int
		configure func(*entitlementRepository)
		cancel    bool
		wantError bool
	}{
		{name: "empty default batch", limit: 0},
		{name: "claim error", limit: 1, configure: func(repository *entitlementRepository) { repository.claimErr = testError }, wantError: true},
		{name: "complete error", limit: 1, configure: func(repository *entitlementRepository) {
			job := model.OutboxJob{ID: "unknown", Kind: "unknown"}
			repository.jobs = []*model.OutboxJob{&job}
			repository.completeErr = testError
		}, wantError: true},
		{name: "cancelled context", limit: 1, cancel: true, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := &entitlementRepository{}
			if test.configure != nil {
				test.configure(repository)
			}
			ctx, cancel := context.WithCancel(context.Background())
			if test.cancel {
				cancel()
			} else {
				defer cancel()
			}
			err := newEntitlementWorkerForTest(repository, &entitlementRemnawave{}).Drain(ctx, test.limit)
			if test.wantError && err == nil {
				t.Fatal("Drain() unexpectedly succeeded")
			}
			if !test.wantError && err != nil {
				t.Fatalf("Drain(): %v", err)
			}
		})
	}
}

func TestWorkerProcessApplyBranches(t *testing.T) {
	t.Parallel()

	remoteID := "remote-1"
	testError := errors.New("test failure")
	future := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		repository *entitlementRepository
		remote     *entitlementRemnawave
		wantError  bool
		wantApply  bool
	}{
		{name: "cancelled purchase", repository: &entitlementRepository{purchase: model.Purchase{Status: "cancelled"}}, remote: &entitlementRemnawave{}},
		{name: "expired purchase", repository: &entitlementRepository{purchase: model.Purchase{Status: "expired"}}, remote: &entitlementRemnawave{}},
		{name: "purchase lookup", repository: &entitlementRepository{purchaseErr: testError}, remote: &entitlementRemnawave{}, wantError: true},
		{name: "user lookup", repository: &entitlementRepository{purchase: model.Purchase{ID: "purchase", ValidUntil: future}, userForPurchaseErr: testError}, remote: &entitlementRemnawave{}, wantError: true},
		{name: "missing remote identity", repository: &entitlementRepository{purchase: model.Purchase{ID: "purchase", ValidUntil: future}, user: model.User{}}, remote: &entitlementRemnawave{}, wantError: true},
		{name: "reset phase lookup", repository: &entitlementRepository{purchase: model.Purchase{ID: "purchase", ValidUntil: future}, user: model.User{RemnaUserID: &remoteID}, resetPhaseErr: testError}, remote: &entitlementRemnawave{}, wantError: true},
		{name: "quiesce", repository: &entitlementRepository{purchase: model.Purchase{ID: "purchase", ValidUntil: future}, user: model.User{RemnaUserID: &remoteID}}, remote: &entitlementRemnawave{removeErr: testError}, wantError: true},
		{name: "reset traffic", repository: &entitlementRepository{purchase: model.Purchase{ID: "purchase", ValidUntil: future}, user: model.User{RemnaUserID: &remoteID}, resetPhase: "quiesced"}, remote: &entitlementRemnawave{resetErr: testError}, wantError: true},
		{name: "unknown reset phase", repository: &entitlementRepository{purchase: model.Purchase{ID: "purchase", ValidUntil: future}, user: model.User{RemnaUserID: &remoteID}, resetPhase: "mystery"}, remote: &entitlementRemnawave{}, wantError: true},
		{name: "remote apply", repository: &entitlementRepository{purchase: model.Purchase{ID: "purchase", ValidUntil: future}, user: model.User{RemnaUserID: &remoteID}}, remote: &entitlementRemnawave{applyErr: testError}, wantError: true, wantApply: true},
		{name: "mark result", repository: &entitlementRepository{purchase: model.Purchase{ID: "purchase", ValidUntil: future}, user: model.User{RemnaUserID: &remoteID}, markErr: testError}, remote: &entitlementRemnawave{}, wantError: true, wantApply: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := newEntitlementWorkerForTest(test.repository, test.remote).process(context.Background(), model.OutboxJob{Kind: "remna_apply_entitlement", Payload: `{"purchaseId":"purchase"}`})
			if test.wantError && err == nil {
				t.Fatal("process() unexpectedly succeeded")
			}
			if !test.wantError && err != nil {
				t.Fatalf("process(): %v", err)
			}
			if got := test.remote.applyCalls > 0; got != test.wantApply {
				t.Fatalf("ApplyEntitlement called = %t, want %t", got, test.wantApply)
			}
		})
	}
}

func TestWorkerProcessExpiresDuePurchase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expireErr error
		wantError bool
	}{
		{name: "success"},
		{name: "repository error", expireErr: errors.New("expire failure"), wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := &entitlementRepository{purchase: model.Purchase{ID: "purchase", Status: "active", ValidUntil: time.Date(2026, 8, 7, 11, 0, 0, 0, time.UTC)}, expireErr: test.expireErr}
			err := newEntitlementWorkerForTest(repository, &entitlementRemnawave{}).process(context.Background(), model.OutboxJob{Kind: "remna_apply_entitlement", Payload: `{"purchaseId":"purchase"}`})
			if test.wantError && err == nil {
				t.Fatal("process() unexpectedly succeeded")
			}
			if !test.wantError && err != nil {
				t.Fatalf("process(): %v", err)
			}
			if repository.expiredPurchaseID != "purchase" {
				t.Fatalf("expired purchase ID = %q", repository.expiredPurchaseID)
			}
		})
	}
}

func TestWorkerProcessSynchronizesDesiredOrEmptyState(t *testing.T) {
	t.Parallel()

	remoteID := "remote-1"
	desired := model.Purchase{TrafficLimitBytes: 4321, ResetStrategy: "WEEK", SquadUUIDs: []string{"squad"}}
	testError := errors.New("test failure")
	tests := []struct {
		name       string
		repository *entitlementRepository
		remote     *entitlementRemnawave
		wantError  bool
		wantApply  bool
		wantRemove bool
	}{
		{name: "user lookup", repository: &entitlementRepository{userByIDErr: testError}, remote: &entitlementRemnawave{}, wantError: true},
		{name: "no remote identity", repository: &entitlementRepository{user: model.User{}}, remote: &entitlementRemnawave{}},
		{name: "desired lookup", repository: &entitlementRepository{user: model.User{RemnaUserID: &remoteID}, desiredErr: testError}, remote: &entitlementRemnawave{}, wantError: true},
		{name: "remove empty", repository: &entitlementRepository{user: model.User{RemnaUserID: &remoteID}}, remote: &entitlementRemnawave{}, wantRemove: true},
		{name: "remove error", repository: &entitlementRepository{user: model.User{RemnaUserID: &remoteID}}, remote: &entitlementRemnawave{removeErr: testError}, wantRemove: true, wantError: true},
		{name: "apply desired", repository: &entitlementRepository{user: model.User{RemnaUserID: &remoteID}, desired: &desired}, remote: &entitlementRemnawave{}, wantApply: true},
		{name: "apply desired error", repository: &entitlementRepository{user: model.User{RemnaUserID: &remoteID}, desired: &desired}, remote: &entitlementRemnawave{applyErr: testError}, wantApply: true, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := newEntitlementWorkerForTest(test.repository, test.remote).process(context.Background(), model.OutboxJob{Kind: "remna_sync_user", Payload: `{"userId":"user-1"}`})
			if test.wantError && err == nil {
				t.Fatal("process() unexpectedly succeeded")
			}
			if !test.wantError && err != nil {
				t.Fatalf("process(): %v", err)
			}
			if got := test.remote.applyCalls > 0; got != test.wantApply {
				t.Fatalf("ApplyEntitlement called = %t, want %t", got, test.wantApply)
			}
			if got := test.remote.removeCalls > 0; got != test.wantRemove {
				t.Fatalf("RemoveEntitlement called = %t, want %t", got, test.wantRemove)
			}
			if test.wantApply && test.remote.resetCalls != 0 {
				t.Fatal("sync-user apply unexpectedly reset traffic")
			}
		})
	}
}

func TestWorkerProcessRejectsUnknownJob(t *testing.T) {
	t.Parallel()

	err := newEntitlementWorkerForTest(&entitlementRepository{}, &entitlementRemnawave{}).process(context.Background(), model.OutboxJob{Kind: "mystery"})
	if err == nil {
		t.Fatal("process() unexpectedly accepted unknown job")
	}
}

type entitlementRepository struct {
	jobs               []*model.OutboxJob
	claimIndex         int
	claimErr           error
	completeErr        error
	purchase           model.Purchase
	purchaseErr        error
	user               model.User
	userForPurchaseErr error
	userByIDErr        error
	desired            *model.Purchase
	desiredErr         error
	resetPhase         string
	resetPhaseErr      error
	advanceResetErr    error
	resetTransitions   [][2]string
	markErr            error
	markedPurchaseID   string
	markedSuccess      *bool
	expiredPurchaseID  string
	expireErr          error
	completedJobID     string
	completedAttempts  int
	completedErr       error
}

func (r *entitlementRepository) ClaimOutboxJob(context.Context, time.Time) (*model.OutboxJob, error) {
	if r.claimErr != nil {
		return nil, r.claimErr
	}
	if r.claimIndex >= len(r.jobs) {
		return nil, nil
	}
	job := r.jobs[r.claimIndex]
	r.claimIndex++
	return job, nil
}
func (r *entitlementRepository) CompleteOutboxJob(_ context.Context, jobID string, attempts int, jobErr error, _ time.Time) error {
	r.completedJobID, r.completedAttempts, r.completedErr = jobID, attempts, jobErr
	return r.completeErr
}
func (r *entitlementRepository) PurchaseByID(context.Context, string) (model.Purchase, error) {
	return r.purchase, r.purchaseErr
}
func (r *entitlementRepository) UserForPurchase(context.Context, string) (model.User, error) {
	return r.user, r.userForPurchaseErr
}
func (r *entitlementRepository) UserByID(context.Context, string) (model.User, error) {
	return r.user, r.userByIDErr
}
func (r *entitlementRepository) DesiredEntitlement(context.Context, string, time.Time) (*model.Purchase, error) {
	return r.desired, r.desiredErr
}
func (r *entitlementRepository) PurchaseTrafficResetPhase(context.Context, string) (string, error) {
	if r.resetPhaseErr != nil {
		return "", r.resetPhaseErr
	}
	if r.resetPhase == "" {
		return "pending", nil
	}
	return r.resetPhase, nil
}
func (r *entitlementRepository) AdvancePurchaseTrafficReset(_ context.Context, _ string, from, to string, _ time.Time) error {
	if r.advanceResetErr != nil {
		return r.advanceResetErr
	}
	r.resetTransitions = append(r.resetTransitions, [2]string{from, to})
	r.resetPhase = to
	return nil
}
func (r *entitlementRepository) MarkPurchaseSyncResult(_ context.Context, purchaseID string, success bool, _ time.Time) error {
	r.markedPurchaseID = purchaseID
	r.markedSuccess = new(bool)
	*r.markedSuccess = success
	return r.markErr
}
func (r *entitlementRepository) ExpirePurchase(_ context.Context, purchaseID string, _ time.Time) error {
	r.expiredPurchaseID = purchaseID
	return r.expireErr
}

type entitlementRemnawave struct {
	applyCalls         int
	applyUserID        string
	applyLimit         int64
	applyResetStrategy string
	applySquads        []string
	applyErr           error
	removeCalls        int
	removeErr          error
	resetCalls         int
	resetErr           error
}

func (r *entitlementRemnawave) ApplyEntitlement(_ context.Context, userID string, limit int64, strategy string, squads []string) error {
	r.applyCalls++
	r.applyUserID, r.applyLimit, r.applyResetStrategy = userID, limit, strategy
	r.applySquads = append([]string(nil), squads...)
	return r.applyErr
}
func (r *entitlementRemnawave) ResetTraffic(context.Context, string) error {
	r.resetCalls++
	return r.resetErr
}
func (r *entitlementRemnawave) RemoveEntitlement(context.Context, string) error {
	r.removeCalls++
	return r.removeErr
}

func newEntitlementWorkerForTest(repository Repository, remnawave RemnawaveClient) *Worker {
	worker := NewWorker(repository, remnawave)
	worker.now = func() time.Time { return time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC) }
	return worker
}
