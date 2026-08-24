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
