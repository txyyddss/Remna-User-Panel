package databaseadmin

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
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

func TestMutationReasonContractRequiresFourCharacters(t *testing.T) {
	t.Parallel()

	if _, err := normalizeReason("abc"); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("normalizeReason(three characters) error = %v, want ErrInvalidValue", err)
	}
	if reason, err := normalizeReason(" abcd "); err != nil || reason != "abcd" {
		t.Fatalf("normalizeReason(four characters) = (%q, %v), want (abcd, nil)", reason, err)
	}
}
