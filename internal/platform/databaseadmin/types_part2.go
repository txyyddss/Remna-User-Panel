package databaseadmin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

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

