// Package databaseadmin provides a schema-aware administrative editor without
// exposing raw SQL or trusting browser-supplied identifiers and types.
package databaseadmin

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

var (
	// ErrTableNotFound indicates that the table is excluded or is not part of
	// the application schema discovered from sqlite_schema.
	ErrTableNotFound = errors.New("database table not found")
	// ErrRecordNotFound indicates that the declared key no longer identifies a row.
	ErrRecordNotFound = errors.New("database record not found")
	// ErrOptimisticConflict indicates that a row changed after it was displayed.
	ErrOptimisticConflict = errors.New("database record changed")
	// ErrReviewConflict indicates that a review is missing, expired, consumed,
	// or does not cover the exact requested mutation.
	ErrReviewConflict = errors.New("database mutation review conflicts")
	// ErrInvalidValue indicates that a typed editor value does not match its
	// SQLite column affinity or nullability.
	ErrInvalidValue = errors.New("invalid database editor value")
	// ErrConfirmation indicates that the typed destructive confirmation differs.
	ErrConfirmation = errors.New("database mutation confirmation does not match")
)

const bypassWarning = "Direct edits bypass domain synchronization hooks. Use a domain-specific admin action whenever possible."

// Backupper creates a verified rescue snapshot immediately before a direct edit.
type Backupper interface {
	Run(context.Context) (model.BackupRun, error)
}

// Vault seals write-only settings and opaque cursors/keys.
type Vault interface {
	Encrypt(string, string) (string, error)
	Decrypt(string, string) (string, error)
}

// Service owns schema introspection, typed CRUD, review consumption, rescue
// backups, optimistic hashes, and redacted audit logging.
type Service struct {
	db       *sql.DB
	backups  Backupper
	vault    Vault
	logger   *slog.Logger
	now      func() time.Time
	mutation sync.Mutex
}

// NewService constructs a database editor. A nil logger is replaced with a
// discard logger; backups and vault are required when applying mutations.
func NewService(db *sql.DB, backups Backupper, vault Vault, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Service{db: db, backups: backups, vault: vault, logger: logger, now: time.Now}
}

// Column describes a schema-declared editable field.
type Column struct {
	Name               string `json:"name"`
	DeclaredType       string `json:"declaredType"`
	Nullable           bool   `json:"nullable"`
	PrimaryKeyPosition int    `json:"primaryKeyPosition"`
	Editable           bool   `json:"editable"`
	Sensitive          bool   `json:"sensitive"`
}

// Table describes an allowlisted application table.
type Table struct {
	Name          string   `json:"name"`
	Columns       []Column `json:"columns"`
	HighRisk      bool     `json:"highRisk"`
	SupportsRowID bool     `json:"supportsRowId"`
	Warning       string   `json:"warning"`
}

type valueKind uint8

const (
	valueNull valueKind = iota
	valueString
	valueBoolean
	valueBlob
	valueMasked
)

// Value is the JSON wire union used by typed record editors. SQLite integers
// and real/numeric values intentionally use the string variant, preserving
// decimal representation without JavaScript precision loss.
type Value struct {
	kind valueKind
	text string
	bool bool
	blob []byte
}

// NullValue creates a SQL NULL editor value.
func NullValue() Value { return Value{kind: valueNull} }

// StringValue creates a text or decimal-string editor value.
func StringValue(value string) Value { return Value{kind: valueString, text: value} }

// BooleanValue creates a boolean editor value stored as SQLite 0 or 1.
func BooleanValue(value bool) Value { return Value{kind: valueBoolean, bool: value} }

// BlobValue creates a binary editor value encoded as blobBase64 on the wire.
func BlobValue(value []byte) Value {
	return Value{kind: valueBlob, blob: append([]byte(nil), value...)}
}

func maskedValue() Value { return Value{kind: valueMasked} }

// IsNull reports whether the value represents SQL NULL.
func (v Value) IsNull() bool { return v.kind == valueNull }

// Text returns a string/decimal-string value.
func (v Value) Text() (string, bool) { return v.text, v.kind == valueString }

// Bool returns a boolean value.
func (v Value) Bool() (bool, bool) { return v.bool, v.kind == valueBoolean }

// Blob returns a defensive copy of a blob value.
func (v Value) Blob() ([]byte, bool) { return append([]byte(nil), v.blob...), v.kind == valueBlob }

// MarshalJSON emits the public DatabaseValue union.
func (v Value) MarshalJSON() ([]byte, error) {
	switch v.kind {
	case valueNull:
		return []byte("null"), nil
	case valueString:
		return json.Marshal(v.text)
	case valueBoolean:
		return json.Marshal(v.bool)
	case valueBlob:
		return json.Marshal(map[string]string{"blobBase64": base64.StdEncoding.EncodeToString(v.blob)})
	case valueMasked:
		return json.Marshal("********")
	default:
		return nil, errors.New("unknown database value kind")
	}
}

// UnmarshalJSON accepts null, a string, a boolean, or a single
// {"blobBase64":"..."} object. Numeric SQLite values must arrive as decimal
// strings so a JavaScript client cannot round a 64-bit integer before the
// server validates it.
func (v *Value) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var raw any
	if err := decoder.Decode(&raw); err != nil {
		return fmt.Errorf("decode database value: %w", err)
	}
	switch typed := raw.(type) {
	case nil:
		*v = NullValue()
	case string:
		*v = StringValue(typed)
	case json.Number:
		return fmt.Errorf("%w: numeric values must use decimal strings", ErrInvalidValue)
	case bool:
		*v = BooleanValue(typed)
	case map[string]any:
		if len(typed) != 1 {
			return fmt.Errorf("%w: blob value contains unknown fields", ErrInvalidValue)
		}
		encoded, ok := typed["blobBase64"].(string)
		if !ok {
			return fmt.Errorf("%w: blobBase64 must be a string", ErrInvalidValue)
		}
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return fmt.Errorf("%w: decode blobBase64: %v", ErrInvalidValue, err)
		}
		*v = BlobValue(decoded)
	default:
		return fmt.Errorf("%w: unsupported JSON value", ErrInvalidValue)
	}
	return nil
}

// Record is a row with its declared key and optimistic content hash.
type Record struct {
	Key        map[string]string `json:"key"`
	Values     map[string]Value  `json:"values"`
	RecordHash string            `json:"recordHash"`
}

// Page is a stable primary-key or rowid cursor page.
type Page struct {
	Items      []Record `json:"items"`
	NextCursor *string  `json:"nextCursor"`
}

// QueryFilter is one allowlisted, typed predicate. Value is omitted for null
// operators and otherwise uses the existing precision-safe DatabaseValue union.
type QueryFilter struct {
	Column   string `json:"column"`
	Operator string `json:"operator"`
	Value    *Value `json:"value,omitempty"`
}

// QueryRequest combines broad search with at most five typed column filters.
type QueryRequest struct {
	Search  string        `json:"search"`
	Filters []QueryFilter `json:"filters"`
	Cursor  string        `json:"cursor,omitempty"`
	Limit   int           `json:"limit,omitempty"`
}

// MutationRequest is the exact reviewed mutation. Key is omitted for inserts;
// ExpectedHash is required for updates and deletes.
type MutationRequest struct {
	Action       string            `json:"action"`
	Table        string            `json:"table"`
	Key          map[string]string `json:"key,omitempty"`
	Values       map[string]Value  `json:"values,omitempty"`
	ExpectedHash string            `json:"recordHash,omitempty"`
	Reason       string            `json:"reason"`
	ReviewHash   string            `json:"reviewHash,omitempty"`
	Confirmation string            `json:"confirmation,omitempty"`
}

// MutationReview is the redacted before/after diff an administrator must
// inspect before applying a direct edit.
type MutationReview struct {
	Action               string            `json:"action"`
	Table                string            `json:"table"`
	Key                  map[string]string `json:"key,omitempty"`
	Before               map[string]Value  `json:"before,omitempty"`
	After                map[string]Value  `json:"after,omitempty"`
	ChangedColumns       []string          `json:"changedColumns"`
	ReviewHash           string            `json:"reviewHash"`
	RequiredConfirmation string            `json:"requiredConfirmation"`
	RescueBackupRequired bool              `json:"rescueBackupRequired"`
	Warning              string            `json:"warning"`
}

// MutationResult reports the resulting row or a successful deletion and the
// rescue snapshot created immediately before the transaction.
type MutationResult struct {
	Row            *Record `json:"row,omitempty"`
	Deleted        bool    `json:"deleted"`
	RescueBackupID string  `json:"rescueBackupId"`
}

func normalizeReason(reason string) (string, error) {
	reason = strings.TrimSpace(reason)
	if len(reason) < 4 || len(reason) > 500 {
		return "", fmt.Errorf("%w: reason must contain 4 to 500 characters", ErrInvalidValue)
	}
	return reason, nil
}
