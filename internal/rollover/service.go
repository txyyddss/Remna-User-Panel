// Package rollover settles unused-traffic credits before a term can reset.
package rollover

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/outbox"
)

// ErrRemoteUserMissing is an authoritative upstream 404, not a transient error.
var ErrRemoteUserMissing = errors.New("Remnawave user is missing")

type Repository interface {
	RolloverByPurchase(context.Context, string) (model.PurchaseRollover, error)
	UserForPurchase(context.Context, string) (model.User, error)
	MarkRolloverProcessing(context.Context, string, time.Time) error
	FinalizeRollover(context.Context, string, int64, int64, string, time.Time) (model.PurchaseRollover, error)
}

type usageFinalizer interface {
	FinalizeRolloverUsage(context.Context, string, model.RolloverUsageSummary, string, time.Time) (model.PurchaseRollover, error)
}

type purchaseLoader interface {
	PurchaseByID(context.Context, string) (model.Purchase, error)
}

type Remote interface {
	QuiesceForRollover(context.Context, string) error
	TrafficForRollover(context.Context, string) (limitBytes, usedBytes int64, err error)
}

type DailyUsage struct {
	Date  time.Time
	Bytes int64
}
type UsageSnapshot struct {
	LimitBytes  int64
	Strategy    string
	LastResetAt *time.Time
	Daily       []DailyUsage
}
type usageSnapshotRemote interface {
	UsageSnapshotForRollover(context.Context, string, time.Time, time.Time) (UsageSnapshot, error)
}

type Service struct {
	repository Repository
	remote     Remote
	now        func() time.Time
}

func NewService(repository Repository, remote Remote) *Service {
	return &Service{repository: repository, remote: remote, now: time.Now}
}

func (s *Service) HandleOutbox(ctx context.Context, job model.OutboxJob) error {
	if job.Kind != "rollover_finalize" {
		return fmt.Errorf("unsupported rollover job %q", job.Kind)
	}
	purchaseID, err := outbox.TargetID(job, "purchaseId")
	if err != nil {
		return err
	}
	rollover, err := s.repository.RolloverByPurchase(ctx, purchaseID)
	if err != nil {
		return err
	}
	if rollover.Status == "credited" || rollover.Status == "zero" || rollover.Status == "exception" {
		return nil
	}
	user, err := s.repository.UserForPurchase(ctx, purchaseID)
	if err != nil {
		return err
	}
	if user.RemnaUserID == nil {
		if rollover.Status == "pending" {
			if err := s.repository.MarkRolloverProcessing(ctx, purchaseID, s.now().UTC()); err != nil {
				return err
			}
		}
		_, err := s.repository.FinalizeRollover(ctx, purchaseID, 0, 0, "local_identity_missing", s.now().UTC())
		return err
	}
	if rollover.Status == "pending" {
		if err := s.remote.QuiesceForRollover(ctx, *user.RemnaUserID); err != nil {
			if errors.Is(err, ErrRemoteUserMissing) {
				if markErr := s.repository.MarkRolloverProcessing(ctx, purchaseID, s.now().UTC()); markErr != nil {
					return markErr
				}
				_, finalizeErr := s.repository.FinalizeRollover(ctx, purchaseID, 0, 0, "remnawave_user_missing", s.now().UTC())
				return finalizeErr
			}
			return fmt.Errorf("quiesce rollover: %w", err)
		}
		if err := s.repository.MarkRolloverProcessing(ctx, purchaseID, s.now().UTC()); err != nil {
			return err
		}
	}
	if extended, ok := s.remote.(usageSnapshotRemote); ok {
		if loader, loadOK := s.repository.(purchaseLoader); loadOK {
			purchase, loadErr := loader.PurchaseByID(ctx, purchaseID)
			if loadErr != nil {
				return loadErr
			}
			snapshot, snapshotErr := extended.UsageSnapshotForRollover(ctx, *user.RemnaUserID, purchase.ValidFrom, purchase.ValidUntil)
			if errors.Is(snapshotErr, ErrRemoteUserMissing) {
				_, snapshotErr = s.repository.FinalizeRollover(ctx, purchaseID, 0, 0, "remnawave_user_missing", s.now().UTC())
				return snapshotErr
			}
			if snapshotErr != nil {
				return fmt.Errorf("fetch rollover traffic: %w", snapshotErr)
			}
			if finalizer, finalizerOK := s.repository.(usageFinalizer); finalizerOK {
				summary := CalculateUsage(purchase, rollover.MinimumRemainingBPS, snapshot)
				_, finalErr := finalizer.FinalizeRolloverUsage(ctx, purchaseID, summary, "", s.now().UTC())
				return finalErr
			}
		}
	}
	limit, used, err := s.remote.TrafficForRollover(ctx, *user.RemnaUserID)
	if errors.Is(err, ErrRemoteUserMissing) {
		_, err = s.repository.FinalizeRollover(ctx, purchaseID, 0, 0, "remnawave_user_missing", s.now().UTC())
		return err
	}
	if err != nil {
		return fmt.Errorf("fetch rollover traffic: %w", err)
	}
	_, err = s.repository.FinalizeRollover(ctx, purchaseID, limit, used, "", s.now().UTC())
	return err
}

const UsageAlgorithmVersion = "cadence-v1"

// CalculateUsage derives cadence allowances from daily upstream data without
// retaining the raw provider series.
func CalculateUsage(purchase model.Purchase, threshold int, snapshot UsageSnapshot) model.RolloverUsageSummary {
	if snapshot.LimitBytes <= 0 || purchase.ValidUntil.Before(purchase.ValidFrom) {
		return model.RolloverUsageSummary{AlgorithmVersion: UsageAlgorithmVersion}
	}
	start, end := purchase.ValidFrom.UTC(), purchase.ValidUntil.UTC()
	anchor := start
	if snapshot.LastResetAt != nil && snapshot.LastResetAt.Before(end) {
		anchor = snapshot.LastResetAt.UTC()
	}
	periods := cadencePeriods(anchor, start, end, snapshot.Strategy)
	usageByDay := make(map[string]int64, len(snapshot.Daily))
	for _, item := range snapshot.Daily {
		if item.Bytes > 0 {
			usageByDay[item.Date.UTC().Format(time.DateOnly)] += item.Bytes
		}
	}
	var allocated, used, eligible int64
	for _, period := range periods {
		full := cadenceAdvance(period.start, snapshot.Strategy).Sub(period.start)
		if full <= 0 {
			full = end.Sub(start)
		}
		overlap := overlapDuration(period.start, period.end, start, end)
		if overlap <= 0 {
			continue
		}
		allowance := proportionalBytes(snapshot.LimitBytes, overlap.Nanoseconds(), full.Nanoseconds())
		periodUsed := int64(0)
		for day := dayStart(period.start); day.Before(period.end); day = day.AddDate(0, 0, 1) {
			dayEnd := day.AddDate(0, 0, 1)
			portion := overlapDuration(day, dayEnd, period.start, period.end)
			if portion <= 0 {
				continue
			}
			periodUsed += proportionalBytes(usageByDay[day.Format(time.DateOnly)], portion.Nanoseconds(), (24 * time.Hour).Nanoseconds())
		}
		if allowance < 0 {
			allowance = 0
		}
		if periodUsed > allowance {
			periodUsed = allowance
		}
		unused := allowance - periodUsed
		if allowance > 0 && unused*10000 > allowance*int64(threshold) {
			allocated += allowance
			used += periodUsed
			eligible += unused
		}
	}
	return model.RolloverUsageSummary{AllocatedBytes: allocated, UsedBytes: used, EligibleUnusedBytes: eligible, AlgorithmVersion: UsageAlgorithmVersion}
}

type cadencePeriod struct{ start, end time.Time }

func cadencePeriods(anchor, start, end time.Time, strategy string) []cadencePeriod {
	if strategy == "NO_RESET" {
		return []cadencePeriod{{start: start, end: end}}
	}
	periodStart := anchor
	for periodStart.After(start) {
		periodStart = cadencePrevious(periodStart, strategy)
	}
	result := make([]cadencePeriod, 0)
	for periodStart.Before(end) {
		next := cadenceAdvance(periodStart, strategy)
		result = append(result, cadencePeriod{periodStart, next})
		periodStart = next
	}
	return result
}
func cadenceAdvance(value time.Time, strategy string) time.Time {
	switch strategy {
	case "DAY":
		return value.AddDate(0, 0, 1)
	case "WEEK":
		return value.AddDate(0, 0, 7)
	case "MONTH", "MONTH_ROLLING":
		return value.AddDate(0, 1, 0)
	default:
		return value.AddDate(0, 0, 1)
	}
}
func cadencePrevious(value time.Time, strategy string) time.Time {
	switch strategy {
	case "DAY":
		return value.AddDate(0, 0, -1)
	case "WEEK":
		return value.AddDate(0, 0, -7)
	case "MONTH", "MONTH_ROLLING":
		return value.AddDate(0, -1, 0)
	default:
		return value.AddDate(0, 0, -1)
	}
}
func dayStart(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
func overlapDuration(left, right, start, end time.Time) time.Duration {
	if left.Before(start) {
		left = start
	}
	if right.After(end) {
		right = end
	}
	if !right.After(left) {
		return 0
	}
	return right.Sub(left)
}

func proportionalBytes(value, numerator, denominator int64) int64 {
	if value <= 0 || numerator <= 0 || denominator <= 0 {
		return 0
	}
	result := new(big.Int).Mul(big.NewInt(value), big.NewInt(numerator))
	result.Quo(result, big.NewInt(denominator))
	if !result.IsInt64() {
		return value
	}
	return result.Int64()
}
