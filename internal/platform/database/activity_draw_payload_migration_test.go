package database

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestDrawPayloadMigrationRetriesOnlyRevisionDecodeFailures(t *testing.T) {
	t.Parallel()
	ctx, store := context.Background(), newTestStore(t)
	const migration = "053_retry_draw_payload_decode.sql"
	const previousTime = "2100-01-01T00:00:00Z"
	decodeError := func(kind string) string {
		return fmt.Sprintf("handle outbox kind %s: decode %s job payload: json: cannot unmarshal number into Go value of type string", kind, kind)
	}
	tests := []struct {
		id, kind, payload, status, failure string
		retry                              bool
	}{
		{"update", "draw_telegram_update", "", "failed", decodeError("draw_telegram_update"), true},
		{"z-duplicate update", "draw_telegram_update", `{"drawId":"update","revision":2}`, "failed", decodeError("draw_telegram_update"), false},
		{"a-unrelated duplicate", "draw_telegram_update", `{"drawId":"update","revision":2}`, "failed", "telegram request timed out", false},
		{"settle", "draw_raffle_settle", "", "failed", decodeError("draw_raffle_settle"), true},
		{"delayed settle", "draw_raffle_settle", "", "pending", decodeError("draw_raffle_settle"), true},
		{"delayed update", "draw_telegram_update", "", "pending", decodeError("draw_telegram_update"), true},
		{"in flight", "draw_raffle_settle", "", "processing", decodeError("draw_raffle_settle"), false},
		{"failed with pending", "draw_raffle_settle", `{"drawId":"delayed settle","revision":2}`, "failed", decodeError("draw_raffle_settle"), false},
		{"failed with processing", "draw_raffle_settle", `{"drawId":"in flight","revision":2}`, "failed", decodeError("draw_raffle_settle"), false},
		{"telegram failure", "draw_telegram_update", "", "failed", "telegram editMessageText failed (http=403 api=403): Forbidden", false},
		{"settlement failure", "draw_raffle_settle", "", "failed", "handle outbox kind draw_raffle_settle: conflict", false},
		{"other kind", "draw_telegram_publish", "", "failed", decodeError("draw_telegram_publish"), false},
		{"wrong error kind", "draw_telegram_update", "", "failed", decodeError("draw_raffle_settle"), false},
		{"numeric target", "draw_telegram_update", `{"drawId":123,"revision":2}`, "failed", decodeError("draw_telegram_update"), false},
		{"blank target", "draw_telegram_update", `{"drawId":" ","revision":2}`, "failed", decodeError("draw_telegram_update"), false},
		{"missing target", "draw_telegram_update", `{"revision":2}`, "failed", decodeError("draw_telegram_update"), false},
		{"no revision", "draw_telegram_update", `{"drawId":"draw-1"}`, "failed", decodeError("draw_telegram_update"), false},
		{"fresh pending", "draw_telegram_update", "", "pending", "", false},
		{"done", "draw_raffle_settle", "", "done", decodeError("draw_raffle_settle"), false},
	}
	for _, test := range tests {
		payload := test.payload
		if payload == "" {
			payload = fmt.Sprintf(`{"drawId":%q,"revision":2}`, test.id)
		}
		_, err := store.DB().ExecContext(ctx, `INSERT INTO outbox_jobs
  (id,kind,payload,status,attempts,available_at,last_error,created_at,updated_at)
  VALUES(?,?,?,?,10,?,?,?,?)`, test.id, test.kind, payload, test.status, previousTime, test.failure, previousTime, previousTime)
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := store.DB().ExecContext(ctx, `DELETE FROM schema_migrations WHERE version=?`, migration); err != nil {
		t.Fatal(err)
	}
	if err := migrate(ctx, store.DB()); err != nil {
		t.Fatal(err)
	}
	for _, test := range tests {
		t.Run(test.id, func(t *testing.T) {
			var status, failure, available, created, updated string
			var attempts int
			err := store.DB().QueryRowContext(ctx, `SELECT status,attempts,last_error,available_at,created_at,updated_at
  FROM outbox_jobs WHERE id=?`, test.id).Scan(&status, &attempts, &failure, &available, &created, &updated)
			if err != nil {
				t.Fatal(err)
			}
			if created != previousTime {
				t.Fatalf("created_at changed: %q", created)
			}
			if !test.retry {
				if status != test.status || attempts != 10 || failure != test.failure || available != previousTime || updated != previousTime {
					t.Fatalf("unrelated job changed: status=%q attempts=%d failure=%q available=%q updated=%q", status, attempts, failure, available, updated)
				}
				return
			}
			if status != "pending" || attempts != 0 || failure != "" || available != updated {
				t.Fatalf("job not reset: status=%q attempts=%d failure=%q available=%q updated=%q", status, attempts, failure, available, updated)
			}
			due, err := time.Parse(time.RFC3339Nano, available)
			if err != nil || due.After(time.Now().UTC()) {
				t.Fatalf("retry is not immediately due: available=%q, err=%v", available, err)
			}
		})
	}
}
