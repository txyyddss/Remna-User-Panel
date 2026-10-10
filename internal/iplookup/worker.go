package iplookup

import (
	"context"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/providerops"
)

// WorkerRepository persists progress before each external request and atomically settles costs.
type WorkerRepository interface {
	BeginProviderOperationAttempt(context.Context, string, time.Time) (providerops.Operation, error)
	IPLookupRun(context.Context, string) (Report, Config, error)
	SaveIPLookupProgress(context.Context, Report) error
	FinishIPLookupRun(context.Context, Report, time.Time) error
}

// Provider is a queued adapter with a stable registry identity.
type Provider interface {
	ID() string
	Lookup(context.Context, string, ProviderConfig) (ProviderResult, error)
}

// Worker resumes report stages in the single existing durable operation lane.
type Worker struct {
	repository WorkerRepository
	providers  map[string]Provider
	now        func() time.Time
}

// NewWorker registers adapters without coupling verdict or billing to their HTTP schemas.
func NewWorker(repo WorkerRepository, providers []Provider) *Worker {
	w := &Worker{repository: repo, providers: map[string]Provider{}, now: time.Now}
	for _, p := range providers {
		w.providers[p.ID()] = p
	}
	return w
}

// HandleProviderOperation persists completed stages and does not retry interrupted HTTP calls.
func (w *Worker) HandleProviderOperation(ctx context.Context, operation providerops.Operation, _ model.OutboxJob) error {
	if providerops.Terminal(providerops.Status(operation.Receipt.Status)) {
		return nil
	}
	if operation.Receipt.Status == "queued" {
		if _, err := w.repository.BeginProviderOperationAttempt(ctx, operation.Receipt.ID, w.now().UTC()); err != nil {
			return err
		}
	}
	r, c, err := w.repository.IPLookupRun(ctx, operation.Receipt.ID)
	if err != nil {
		return err
	}
	if r.Status != "processing" {
		return w.repository.FinishIPLookupRun(ctx, r, w.now().UTC())
	}
	for index := range r.Providers {
		p := &r.Providers[index]
		if p.Status == "disabled" || p.Status == "skipped" {
			continue
		}
		if len(RiskReasons(*p)) > 0 {
			skipLater(&r, index)
			break
		}
		if p.Status == "success" || p.Status == "partial" || p.Status == "error" {
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if p.Status == "processing" {
			p.Status, p.ErrorCode = "error", "IP_LOOKUP_INTERRUPTED"
		} else {
			p.Status = "processing"
			if err := w.repository.SaveIPLookupProgress(ctx, r); err != nil {
				return err
			}
			adapter := w.providers[p.ID]
			if adapter == nil {
				p.Status, p.ErrorCode = "error", "IP_LOOKUP_PROVIDER_UNAVAILABLE"
			} else {
				result, callErr := adapter.Lookup(ctx, r.IP, c.Providers[index])
				if callErr != nil {
					p.Status, p.ErrorCode = "error", ErrorCode(callErr)
				} else {
					*p = result
				}
			}
		}
		p.CheckedAt = w.now().UTC().Format(time.RFC3339Nano)
		if err := w.repository.SaveIPLookupProgress(ctx, r); err != nil {
			return err
		}
		if len(RiskReasons(*p)) > 0 {
			skipLater(&r, index)
			break
		}
	}
	r = Aggregate(r, w.now().UTC())
	return w.repository.FinishIPLookupRun(ctx, r, w.now().UTC())
}

func skipLater(r *Report, index int) {
	for i := index + 1; i < len(r.Providers); i++ {
		if r.Providers[i].Status == "queued" {
			r.Providers[i].Status = "skipped"
		}
	}
}
