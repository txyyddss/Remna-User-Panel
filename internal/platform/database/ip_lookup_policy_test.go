package database

import (
	"context"
	"testing"
	"time"
)

func TestIPLookupCacheIsFreeWithRemainingAllowance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, service, user, _, _ := ipLookupFixture(t, 2)
	first := submitIP(t, service, user, "warm", false)
	finishIP(t, store, first.ID, false)
	before, err := store.Balance(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	state, err := service.State(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	q, err := service.Quote(ctx, user, "150.249.241.62", false)
	if err != nil {
		t.Fatal(err)
	}
	if q.Charge.Minor != "0" || q.UseQuota || q.CacheReportID == "" {
		t.Fatalf("cached quote=%+v", q)
	}
	op := submitIP(t, service, user, "free-cache", false)
	check, err := service.Check(ctx, user, op.ID)
	if err != nil || !check.Cached || check.UsedQuota || check.Charge.Minor != "0" {
		t.Fatalf("cached check=%+v %v", check, err)
	}
	afterState, err := service.State(ctx, user)
	if err != nil || afterState.Allowance.Remaining != state.Allowance.Remaining {
		t.Fatalf("cache consumed allowance: %+v %v", afterState, err)
	}
	after, err := store.Balance(ctx, user)
	if err != nil || after.Minor != before.Minor {
		t.Fatal("cache deducted TXB")
	}
}

func TestIPLookupAnyProviderOutageRefundsPartialOrRejectedReport(t *testing.T) {
	t.Parallel()
	for _, quota := range []int{0, 1} {
		for _, rejected := range []bool{false, true} {
			t.Run(string(rune('a'+quota))+map[bool]string{false: "partial", true: "rejected"}[rejected], func(t *testing.T) {
				ctx := context.Background()
				store, service, user, _, now := ipLookupFixture(t, quota)
				before, err := store.Balance(ctx, user)
				if err != nil {
					t.Fatal(err)
				}
				op := submitIP(t, service, user, "outage", false)
				r, _, err := store.IPLookupRun(ctx, op.ID)
				if err != nil {
					t.Fatal(err)
				}
				r.RefundRequired = true
				r.Status = "partial"
				r.Verdict = "inconclusive"
				if rejected {
					r.Status = "succeeded"
					r.Verdict = "unsuitable"
				}
				if err := store.FinishIPLookupRun(ctx, r, now); err != nil {
					t.Fatal(err)
				}
				if err := store.FinishIPLookupRun(ctx, r, now); err != nil {
					t.Fatal(err)
				}
				check, err := service.Check(ctx, user, op.ID)
				if err != nil || !check.Refunded || check.Charge.Minor != "0" || check.Report == nil {
					t.Fatalf("outage result=%+v %v", check, err)
				}
				state, err := service.State(ctx, user)
				if err != nil || state.Allowance.Remaining != quota {
					t.Fatal("outage consumed quota")
				}
				after, err := store.Balance(ctx, user)
				if err != nil || before.Minor != after.Minor {
					t.Fatal("outage charged TXB")
				}
			})
		}
	}
}

func TestIPLookupMaintenancePreservesReceiptsAndRefundsStalledRuns(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, service, user, _, now := ipLookupFixture(t, 0)
	before, err := store.Balance(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	op := submitIP(t, service, user, "stalled", false)
	if _, err := store.DB().ExecContext(ctx, `UPDATE provider_operations SET created_at=? WHERE id=?`, stamp(now.Add(-48*time.Hour)), op.ID); err != nil {
		t.Fatal(err)
	}
	tx, err := store.DB().BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := pruneProviderOperationsTx(ctx, tx, now.Add(-24*time.Hour), now, map[string]int64{}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	check, err := service.Check(ctx, user, op.ID)
	if err != nil || !check.Refunded || check.Operation.Status != "failed" {
		t.Fatalf("maintenance lost receipt/refund=%+v %v", check, err)
	}
	after, err := store.Balance(ctx, user)
	if err != nil || after.Minor != before.Minor {
		t.Fatal("maintenance stranded debit")
	}
	q, err := service.Quote(ctx, user, "150.249.241.62", false)
	if err != nil {
		t.Fatal(err)
	}
	q.ExpiresAt = time.Now().Add(time.Minute).Unix()
	if q.CacheReportID != "" {
		t.Fatal("stalled empty report became cache")
	}
	// A new command must be able to reserve another run for the same IP.
	if _, err := store.CreateIPLookupCheck(ctx, user, "replacement", q, now); err != nil {
		t.Fatal(err)
	}
}

func TestIPLookupCompletedReceiptSurvivesMaintenance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, service, user, _, now := ipLookupFixture(t, 0)
	op := submitIP(t, service, user, "complete", false)
	finishIP(t, store, op.ID, false)
	tx, err := store.DB().BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := pruneProviderOperationsTx(ctx, tx, now.Add(-24*time.Hour), now, map[string]int64{}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	check, err := service.Check(ctx, user, op.ID)
	if err != nil || check.Report == nil {
		t.Fatalf("completed receipt deleted: %+v %v", check, err)
	}
}

func TestIPLookupLatestCacheHandlesFractionalTimestampOrdering(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, service, user, _, now := ipLookupFixture(t, 0)
	first := submitIP(t, service, user, "whole-second", false)
	r, _, err := store.IPLookupRun(ctx, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	r.Status, r.Verdict = "succeeded", "suitable"
	wholeSecond := now.Truncate(time.Second)
	if err := store.FinishIPLookupRun(ctx, r, wholeSecond); err != nil {
		t.Fatal(err)
	}
	refresh := submitIP(t, service, user, "fractional-refresh", true)
	newReport, _, err := store.IPLookupRun(ctx, refresh.ID)
	if err != nil {
		t.Fatal(err)
	}
	newReport.Status, newReport.Verdict = "succeeded", "suitable"
	if err := store.FinishIPLookupRun(ctx, newReport, wholeSecond.Add(time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	q, err := service.Quote(ctx, user, "150.249.241.62", false)
	if err != nil || q.CacheReportID != newReport.ID {
		t.Fatalf("latest report selected incorrectly: %+v %v", q, err)
	}
}
