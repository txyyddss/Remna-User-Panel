package databaseadmin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/secret"
)

func TestSchemaRowsAndCursorAreTypedAndAllowlisted(t *testing.T) {
	t.Parallel()
	service, db, _, _, actor := newTestService(t)
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `CREATE TABLE rowid_records(label TEXT NOT NULL,amount INTEGER NOT NULL,enabled INTEGER NOT NULL DEFAULT 0)`); err != nil {
		t.Fatalf("create rowid table: %v", err)
	}
	for _, values := range []struct {
		label   string
		amount  int64
		enabled int
	}{{"first", 9007199254740993, 1}, {"second", 2, 0}, {"third", 3, 1}} {
		if _, err := db.ExecContext(ctx, `INSERT INTO rowid_records(label,amount,enabled) VALUES(?,?,?)`, values.label, values.amount, values.enabled); err != nil {
			t.Fatalf("insert rowid record: %v", err)
		}
	}
	tables, err := service.Tables(ctx)
	if err != nil {
		t.Fatalf("Tables(): %v", err)
	}
	foundRowID, foundMigration := false, false
	for _, table := range tables {
		if table.Name == "rowid_records" {
			foundRowID = table.SupportsRowID && table.HighRisk
		}
		if table.Name == "schema_migrations" || strings.HasPrefix(table.Name, "sqlite_") {
			foundMigration = true
		}
	}
	if !foundRowID || foundMigration {
		t.Fatalf("rowid table found=%v, excluded table leaked=%v", foundRowID, foundMigration)
	}
	page, err := service.Records(ctx, "rowid_records", "", 2)
	if err != nil {
		t.Fatalf("Records(first): %v", err)
	}
	if len(page.Items) != 2 || page.NextCursor == nil {
		t.Fatalf("first page = %+v", page)
	}
	if amount, ok := page.Items[0].Values["amount"].Text(); !ok || amount != "9007199254740993" {
		t.Fatalf("integer value = %q, string=%v", amount, ok)
	}
	if enabled, ok := page.Items[0].Values["enabled"].Bool(); !ok || !enabled {
		t.Fatalf("boolean value = %v, bool=%v", enabled, ok)
	}
	page2, err := service.Records(ctx, "rowid_records", *page.NextCursor, 2)
	if err != nil || len(page2.Items) != 1 || page2.NextCursor != nil {
		t.Fatalf("Records(second) = %+v, %v", page2, err)
	}
	if _, err := service.Records(ctx, `users"; DROP TABLE users;--`, "", 50); !errors.Is(err, ErrTableNotFound) {
		t.Fatalf("identifier injection error = %v", err)
	}

	// Ensure the generated actor remains usable; this also guards against the
	// schema introspection accidentally mutating application state.
	if _, err := db.ExecContext(ctx, `SELECT id FROM users WHERE id=?`, actor); err != nil {
		t.Fatalf("actor query: %v", err)
	}
}

func TestDatabaseValueRequiresDecimalStringForNumericJSON(t *testing.T) {
	t.Parallel()
	var value Value
	if err := json.Unmarshal([]byte(`9007199254740993`), &value); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("raw JSON integer error = %v", err)
	}
	if err := json.Unmarshal([]byte(`"9007199254740993"`), &value); err != nil {
		t.Fatalf("decimal-string JSON value: %v", err)
	}
	if text, ok := value.Text(); !ok || text != "9007199254740993" {
		t.Fatalf("decimal-string value=%q, string=%v", text, ok)
	}
}

func TestRowIDTableRemainsAddressableWhenOneAliasIsShadowed(t *testing.T) {
	t.Parallel()
	service, db, backup, _, actor := newTestService(t)
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `CREATE TABLE rowid_shadow(rowid TEXT NOT NULL,label TEXT NOT NULL)`); err != nil {
		t.Fatalf("create rowid-shadow table: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO rowid_shadow(rowid,label) VALUES('display-only','first')`); err != nil {
		t.Fatalf("insert rowid-shadow record: %v", err)
	}
	page, err := service.Records(ctx, "rowid_shadow", "", 10)
	if err != nil {
		t.Fatalf("Records(rowid_shadow): %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].Key["_rowid_"] != "1" {
		t.Fatalf("rowid-shadow page = %+v", page)
	}
	request := MutationRequest{Action: "update", Table: "rowid_shadow", Key: page.Items[0].Key,
		ExpectedHash: page.Items[0].RecordHash, Values: map[string]Value{"label": StringValue("reviewed")},
		Reason: "Verify the remaining rowid alias"}
	review, err := service.ReviewMutation(ctx, actor, request)
	if err != nil {
		t.Fatalf("ReviewMutation(rowid shadow): %v", err)
	}
	request.ReviewHash, request.Confirmation = review.ReviewHash, review.RequiredConfirmation
	result, err := service.ApplyMutation(ctx, actor, request)
	if err != nil {
		t.Fatalf("ApplyMutation(rowid shadow): %v", err)
	}
	if backup.calls != 1 || result.Row == nil {
		t.Fatalf("rowid-shadow mutation result=%+v backup calls=%d", result, backup.calls)
	}
}

func TestRowIDInsertStillRequiresDeclaredNotNullColumns(t *testing.T) {
	t.Parallel()
	service, db, _, _, actor := newTestService(t)
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `CREATE TABLE rowid_insert_required(label TEXT NOT NULL,enabled INTEGER NOT NULL DEFAULT 0)`); err != nil {
		t.Fatalf("create rowid insert table: %v", err)
	}
	request := MutationRequest{Action: "insert", Table: "rowid_insert_required",
		Values: map[string]Value{"enabled": BooleanValue(true)}, Reason: "Verify required field enforcement"}
	if _, err := service.ReviewMutation(ctx, actor, request); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("missing required rowid-table column error = %v", err)
	}
}

func TestDefaultValuesInsertIsReviewedAndTransactional(t *testing.T) {
	t.Parallel()
	service, db, backup, _, actor := newTestService(t)
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `CREATE TABLE default_value_records(enabled INTEGER NOT NULL DEFAULT 1)`); err != nil {
		t.Fatalf("create default-value table: %v", err)
	}
	request := MutationRequest{Action: "insert", Table: "default_value_records", Reason: "Create the default configuration row"}
	review, err := service.ReviewMutation(ctx, actor, request)
	if err != nil {
		t.Fatalf("ReviewMutation(default values): %v", err)
	}
	if len(review.After) != 0 {
		t.Fatalf("default-value review claimed unresolved values: %+v", review.After)
	}
	request.ReviewHash, request.Confirmation = review.ReviewHash, review.RequiredConfirmation
	result, err := service.ApplyMutation(ctx, actor, request)
	if err != nil {
		t.Fatalf("ApplyMutation(default values): %v", err)
	}
	if backup.calls != 1 || result.Row == nil {
		t.Fatalf("default-value result=%+v backup calls=%d", result, backup.calls)
	}
	if enabled, ok := result.Row.Values["enabled"].Bool(); !ok || !enabled {
		t.Fatalf("default enabled=%v, boolean=%v", enabled, ok)
	}
}

func TestNumericPrimaryKeyCursorPreservesExactLargeInteger(t *testing.T) {
	t.Parallel()
	service, db, _, _, actor := newTestService(t)
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `CREATE TABLE numeric_key_records(id NUMERIC PRIMARY KEY NOT NULL,label TEXT NOT NULL)`); err != nil {
		t.Fatalf("create numeric-key table: %v", err)
	}
	for _, id := range []int64{9007199254740993, 9007199254740995} {
		if _, err := db.ExecContext(ctx, `INSERT INTO numeric_key_records(id,label) VALUES(?,?)`, id, "record"); err != nil {
			t.Fatalf("insert numeric key %d: %v", id, err)
		}
	}
	first, err := service.Records(ctx, "numeric_key_records", "", 1)
	if err != nil || len(first.Items) != 1 || first.NextCursor == nil {
		t.Fatalf("first numeric-key page = %+v, %v", first, err)
	}
	if first.Items[0].Key["id"] != "9007199254740993" {
		t.Fatalf("first numeric key = %q", first.Items[0].Key["id"])
	}
	second, err := service.Records(ctx, "numeric_key_records", *first.NextCursor, 1)
	if err != nil || len(second.Items) != 1 {
		t.Fatalf("second numeric-key page = %+v, %v", second, err)
	}
	if second.Items[0].Key["id"] != "9007199254740995" {
		t.Fatalf("second numeric key = %q", second.Items[0].Key["id"])
	}

	request := MutationRequest{Action: "update", Table: "numeric_key_records", Key: first.Items[0].Key,
		ExpectedHash: first.Items[0].RecordHash, Values: map[string]Value{"label": StringValue("reviewed")},
		Reason: "Verify exact numeric key handling"}
	if _, err := service.ReviewMutation(ctx, actor, request); err != nil {
		t.Fatalf("ReviewMutation(large numeric key): %v", err)
	}
}

func TestReviewedUpdateCreatesRescueAuditAndCannotReplay(t *testing.T) {
	t.Parallel()
	service, db, backup, _, actor := newTestService(t)
	ctx := context.Background()
	page, err := service.Records(ctx, "users", "", 10)
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("Records(users) = %+v, %v", page, err)
	}
	row := page.Items[0]
	request := MutationRequest{Action: "update", Table: "users", Key: row.Key, ExpectedHash: row.RecordHash,
		Values: map[string]Value{"telegram_first_name": StringValue("Reviewed")}, Reason: "Correct the imported display name"}
	review, err := service.ReviewMutation(ctx, actor, request)
	if err != nil {
		t.Fatalf("ReviewMutation(): %v", err)
	}
	if review.ReviewHash == "" || review.RequiredConfirmation != "EDIT users" || len(review.ChangedColumns) != 1 {
		t.Fatalf("review = %+v", review)
	}
	request.ReviewHash, request.Confirmation = review.ReviewHash, review.RequiredConfirmation
	result, err := service.ApplyMutation(ctx, actor, request)
	if err != nil {
		t.Fatalf("ApplyMutation(): %v", err)
	}
	if backup.calls != 1 || result.RescueBackupID == "" || result.Row == nil {
		t.Fatalf("backup calls=%d, result=%+v", backup.calls, result)
	}
	var firstName string
	if err := db.QueryRowContext(ctx, `SELECT telegram_first_name FROM users WHERE id=?`, actor).Scan(&firstName); err != nil || firstName != "Reviewed" {
		t.Fatalf("updated first name=%q, error=%v", firstName, err)
	}
	var detail string
	if err := db.QueryRowContext(ctx, `SELECT detail FROM audit_events WHERE action='database_record_update' ORDER BY created_at DESC LIMIT 1`).Scan(&detail); err != nil {
		t.Fatalf("read audit: %v", err)
	}
	if strings.Contains(detail, "Reviewed") || !strings.Contains(detail, `"values":"[redacted]"`) {
		t.Fatalf("audit detail was not redacted: %s", detail)
	}
	if _, err := service.ApplyMutation(ctx, actor, request); !errors.Is(err, ErrReviewConflict) {
		t.Fatalf("review replay error = %v", err)
	}
}

func TestOptimisticConflictRollsBackReviewAndEdit(t *testing.T) {
	t.Parallel()
	service, db, _, _, actor := newTestService(t)
	ctx := context.Background()
	page, err := service.Records(ctx, "users", "", 10)
	if err != nil {
		t.Fatal(err)
	}
	row := page.Items[0]
	request := MutationRequest{Action: "update", Table: "users", Key: row.Key, ExpectedHash: row.RecordHash,
		Values: map[string]Value{"telegram_last_name": StringValue("planned")}, Reason: "Repair imported account metadata"}
	review, err := service.ReviewMutation(ctx, actor, request)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE users SET telegram_last_name='concurrent' WHERE id=?`, actor); err != nil {
		t.Fatal(err)
	}
	request.ReviewHash, request.Confirmation = review.ReviewHash, review.RequiredConfirmation
	if _, err := service.ApplyMutation(ctx, actor, request); !errors.Is(err, ErrOptimisticConflict) {
		t.Fatalf("optimistic error = %v", err)
	}
	var lastName string
	if err := db.QueryRowContext(ctx, `SELECT telegram_last_name FROM users WHERE id=?`, actor).Scan(&lastName); err != nil || lastName != "concurrent" {
		t.Fatalf("last name=%q, error=%v", lastName, err)
	}
	var consumed sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT consumed_at FROM database_admin_reviews WHERE id=?`, review.ReviewHash).Scan(&consumed); err != nil || consumed.Valid {
		t.Fatalf("review consumed=%v, error=%v", consumed.Valid, err)
	}
}

func TestEncryptedSettingIsMaskedAndReplacedThroughVault(t *testing.T) {
	t.Parallel()
	service, db, _, vault, actor := newTestService(t)
	ctx := context.Background()
	initial, err := vault.Encrypt("provider.secret", "old-secret")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.ExecContext(ctx, `INSERT INTO settings(key,value,encrypted,updated_at,updated_by) VALUES(?,?,?,?,?)`, "provider.secret", initial, 1, now, actor); err != nil {
		t.Fatal(err)
	}
	page, err := service.Records(ctx, "settings", "", 100)
	if err != nil {
		t.Fatal(err)
	}
	var row Record
	for _, candidate := range page.Items {
		if candidate.Key["key"] == "provider.secret" {
			row = candidate
		}
	}
	encoded, err := json.Marshal(row.Values["value"])
	if err != nil || strings.Contains(string(encoded), initial) || string(encoded) != `"********"` {
		t.Fatalf("masked value=%s, error=%v", encoded, err)
	}
	request := MutationRequest{Action: "update", Table: "settings", Key: row.Key, ExpectedHash: row.RecordHash,
		Values: map[string]Value{"value": StringValue("new-secret")}, Reason: "Rotate the provider credential"}
	review, err := service.ReviewMutation(ctx, actor, request)
	if err != nil {
		t.Fatal(err)
	}
	request.ReviewHash, request.Confirmation = review.ReviewHash, review.RequiredConfirmation
	if _, err := service.ApplyMutation(ctx, actor, request); err != nil {
		t.Fatal(err)
	}
	var stored string
	if err := db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key='provider.secret'`).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored == "new-secret" || strings.Contains(stored, "new-secret") {
		t.Fatalf("setting stored plaintext: %q", stored)
	}
	plaintext, err := vault.Decrypt("provider.secret", stored)
	if err != nil || plaintext != "new-secret" {
		t.Fatalf("decrypted setting=%q, error=%v", plaintext, err)
	}
}

func newTestService(t *testing.T) (*Service, *sql.DB, *fakeBackupper, *secret.Vault, string) {
	t.Helper()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "database-admin.db")
	db, err := database.Open(ctx, path)
	if err != nil {
		t.Fatalf("database.Open(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	store := database.NewStore(db)
	user, _, err := store.UpsertTelegramUser(ctx, model.TelegramProfile{ID: 424242, FirstName: "Admin"}, true)
	if err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	vault, err := secret.NewVault([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("secret.NewVault(): %v", err)
	}
	backup := &fakeBackupper{}
	return NewService(db, backup, vault, slog.Default()), db, backup, vault, user.ID
}

type fakeBackupper struct{ calls int }

func (b *fakeBackupper) Run(_ context.Context) (model.BackupRun, error) {
	b.calls++
	return model.BackupRun{ID: "rescue-" + string(rune('0'+b.calls)), Status: "complete"}, nil
}
