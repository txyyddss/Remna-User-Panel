package database

import (
	"context"
	"testing"
	"time"
)

func TestCreateIncidentHonorsWarningCooldown(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 78_001)
	now := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	if _, err := store.DB().ExecContext(ctx, `UPDATE abuse_punishment_rules SET enabled=1 WHERE action='warning'`); err != nil {
		t.Fatal(err)
	}
	policy, err := store.Policy(ctx)
	if err != nil {
		t.Fatal(err)
	}
	policy.WarningCooldownMinutes = 30
	if created, err := store.CreateIncident(ctx, user.ID, now, 10, 10, []string{"global"}, nil, policy, now); err != nil || !created {
		t.Fatalf("first warning = (%t, %v), want (true, nil)", created, err)
	}
	if created, err := store.CreateIncident(ctx, user.ID, now.Add(30*time.Second), 10, 10, []string{"global"}, nil, policy, now.Add(time.Minute)); err != nil || created {
		t.Fatalf("cooldown warning = (%t, %v), want (false, nil)", created, err)
	}
	if created, err := store.CreateIncident(ctx, user.ID, now.Add(31*time.Minute), 10, 10, []string{"global"}, nil, policy, now.Add(31*time.Minute)); err != nil || !created {
		t.Fatalf("warning after cooldown = (%t, %v), want (true, nil)", created, err)
	}
}

func TestPruneAbuseRecordRequiresPunishmentDeliveryAndRestore(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 78_002)
	now := time.Date(2026, 8, 24, 13, 0, 0, 0, time.UTC)
	if _, err := store.DB().ExecContext(ctx, `UPDATE abuse_punishment_rules SET enabled=CASE WHEN action='temporary_ban' THEN 1 ELSE 0 END`); err != nil {
		t.Fatal(err)
	}
	policy, err := store.Policy(ctx)
	if err != nil {
		t.Fatal(err)
	}
	createdAt := now.Add(-8 * 24 * time.Hour)
	if created, createErr := store.CreateIncident(ctx, user.ID, createdAt, 10, 10, []string{"global"}, nil, policy, createdAt); createErr != nil || !created {
		t.Fatalf("CreateIncident() = (%t,%v)", created, createErr)
	}
	var recordID string
	if err = store.DB().QueryRowContext(ctx, `SELECT id FROM abuse_records WHERE user_id=?`, user.ID).Scan(&recordID); err != nil {
		t.Fatal(err)
	}
	pruneAbuseDetails(t, store, now)
	assertAbuseRecordCount(t, store, recordID, 1)
	if err = store.MarkPunishmentCompleted(ctx, recordID, now); err != nil {
		t.Fatal(err)
	}
	pruneAbuseDetails(t, store, now)
	assertAbuseRecordCount(t, store, recordID, 1)
	markAllAbuseDeliveries(t, store, recordID, now)
	pruneAbuseDetails(t, store, now)
	assertAbuseRecordCount(t, store, recordID, 1)
	if err = store.CompleteAbuseRestore(ctx, user.ID, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	pruneAbuseDetails(t, store, now.Add(time.Hour))
	assertAbuseRecordCount(t, store, recordID, 0)
	var facts int
	if err = store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM abuse_incident_facts WHERE incident_id=?`, recordID).Scan(&facts); err != nil || facts != 0 {
		t.Fatalf("remaining facts = %d, error %v", facts, err)
	}
}

func markAllAbuseDeliveries(t *testing.T, store *Store, recordID string, now time.Time) {
	t.Helper()
	if _, err := store.DB().Exec(`UPDATE abuse_notification_deliveries SET delivered_at=? WHERE record_id=?`, stamp(now), recordID); err != nil {
		t.Fatal(err)
	}
}

func pruneAbuseDetails(t *testing.T, store *Store, now time.Time) {
	t.Helper()
	tx, err := store.DB().Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err = store.PruneAbuseRecordsTx(context.Background(), tx, now, map[string]int64{}); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func assertAbuseRecordCount(t *testing.T, store *Store, recordID string, want int) {
	t.Helper()
	var got int
	if err := store.DB().QueryRow(`SELECT COUNT(*) FROM abuse_records WHERE id=?`, recordID).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("record count = %d, want %d", got, want)
	}
}
