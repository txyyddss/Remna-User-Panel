package rollover

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

func TestHandleOutboxOrdersQuiesceTrafficAndFinalization(t *testing.T) {
	t.Parallel()
	transient := errors.New("temporary upstream failure")
	tests := []struct {
		name          string
		status        string
		quiesceErr    error
		trafficErr    error
		wantEvents    []string
		wantException string
		wantError     bool
	}{
		{name: "happy path", status: "pending", wantEvents: []string{"quiesce", "mark", "traffic", "finalize"}},
		{name: "quiesce 404 becomes zero exception", status: "pending", quiesceErr: ErrRemoteUserMissing,
			wantEvents: []string{"quiesce", "mark", "finalize"}, wantException: "remnawave_user_missing"},
		{name: "quiesce 5xx remains retryable", status: "pending", quiesceErr: transient,
			wantEvents: []string{"quiesce"}, wantError: true},
		{name: "traffic 404 becomes zero exception", status: "pending", trafficErr: ErrRemoteUserMissing,
			wantEvents: []string{"quiesce", "mark", "traffic", "finalize"}, wantException: "remnawave_user_missing"},
		{name: "traffic 5xx after durable quiesce remains retryable without requiescing", status: "processing", trafficErr: transient,
			wantEvents: []string{"traffic"}, wantError: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			events := make([]string, 0, 4)
			remoteID := "remote-1"
			repository := &rolloverRepository{events: &events, rollover: model.PurchaseRollover{PurchaseID: "purchase-1", Status: test.status},
				user: model.User{ID: "user-1", RemnaUserID: &remoteID}}
			remote := &rolloverRemote{events: &events, quiesceErr: test.quiesceErr, trafficErr: test.trafficErr, limit: 1_000, used: 250}
			service := NewService(repository, remote)
			service.now = func() time.Time { return time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC) }
			err := service.HandleOutbox(context.Background(), model.OutboxJob{Kind: "rollover_finalize", AggregateID: "purchase-1"})
			if (err != nil) != test.wantError {
				t.Fatalf("HandleOutbox() error = %v, wantError=%t", err, test.wantError)
			}
			if !reflect.DeepEqual(events, test.wantEvents) {
				t.Fatalf("events = %v, want %v", events, test.wantEvents)
			}
			if repository.exception != test.wantException {
				t.Fatalf("exception = %q, want %q", repository.exception, test.wantException)
			}
		})
	}
}

func TestHandleOutboxLocalIdentityMissingFinalizesWithoutRemoteAccess(t *testing.T) {
	t.Parallel()
	events := make([]string, 0, 2)
	repository := &rolloverRepository{events: &events, rollover: model.PurchaseRollover{PurchaseID: "purchase-1", Status: "pending"}, user: model.User{ID: "user-1"}}
	service := NewService(repository, &rolloverRemote{events: &events})
	if err := service.HandleOutbox(context.Background(), model.OutboxJob{Kind: "rollover_finalize", AggregateID: "purchase-1"}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(events, []string{"mark", "finalize"}) || repository.exception != "local_identity_missing" {
		t.Fatalf("events=%v exception=%q", events, repository.exception)
	}
}

type rolloverRepository struct {
	events    *[]string
	rollover  model.PurchaseRollover
	user      model.User
	exception string
}

func (repository *rolloverRepository) RolloverByPurchase(context.Context, string) (model.PurchaseRollover, error) {
	return repository.rollover, nil
}

func (repository *rolloverRepository) UserForPurchase(context.Context, string) (model.User, error) {
	return repository.user, nil
}

func (repository *rolloverRepository) MarkRolloverProcessing(context.Context, string, time.Time) error {
	*repository.events = append(*repository.events, "mark")
	repository.rollover.Status = "processing"
	return nil
}

func (repository *rolloverRepository) FinalizeRollover(_ context.Context, _ string, _, _ int64, exception string, _ time.Time) (model.PurchaseRollover, error) {
	*repository.events = append(*repository.events, "finalize")
	repository.exception = exception
	return repository.rollover, nil
}

type rolloverRemote struct {
	events     *[]string
	quiesceErr error
	trafficErr error
	limit      int64
	used       int64
}

func (remote *rolloverRemote) QuiesceForRollover(context.Context, string) error {
	*remote.events = append(*remote.events, "quiesce")
	return remote.quiesceErr
}

func (remote *rolloverRemote) TrafficForRollover(context.Context, string) (int64, int64, error) {
	*remote.events = append(*remote.events, "traffic")
	return remote.limit, remote.used, remote.trafficErr
}
