package databaseadmin

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

var decimalPattern = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]+)?|\.[0-9]+)$`)

type preparedMutation struct {
	schema         tableSchema
	current        *rowData
	columns        []schemaColumn
	arguments      []any
	afterRaw       []any
	changedColumns []string
}

// ReviewMutation creates a one-use, ten-minute review for the exact redacted
// diff. Review tokens are persisted so an insert cannot be replayed after a
// process restart.
func (s *Service) ReviewMutation(ctx context.Context, actorUserID string, request MutationRequest) (MutationReview, error) {
	request.ReviewHash = ""
	request.Confirmation = ""
	if err := validateAction(request.Action); err != nil {
		return MutationReview{}, err
	}
	reason, err := normalizeReason(request.Reason)
	if err != nil {
		return MutationReview{}, err
	}
	request.Reason = reason
	schema, err := s.schema(ctx, request.Table)
	if err != nil {
		return MutationReview{}, err
	}
	prepared, err := s.prepare(ctx, s.db, schema, request, false)
	if err != nil {
		return MutationReview{}, err
	}
	digest, err := mutationDigest(actorUserID, request)
	if err != nil {
		return MutationReview{}, err
	}
	token, err := ids.Token(32)
	if err != nil {
		return MutationReview{}, fmt.Errorf("create database review token: %w", err)
	}
	now := s.now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return MutationReview{}, fmt.Errorf("begin database review: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM database_admin_reviews WHERE consumed_at IS NOT NULL OR expires_at<?`, now.Format(time.RFC3339Nano)); err != nil {
		return MutationReview{}, fmt.Errorf("prune database reviews: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO database_admin_reviews(id,actor_user_id,action,table_name,request_hash,expires_at,created_at) VALUES(?,?,?,?,?,?,?)`,
		token, actorUserID, request.Action, schema.Name, digest, now.Add(10*time.Minute).Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); err != nil {
		return MutationReview{}, fmt.Errorf("store database review: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return MutationReview{}, fmt.Errorf("commit database review: %w", err)
	}
	review := MutationReview{
		Action: request.Action, Table: schema.Name, Key: cloneKey(request.Key), ChangedColumns: prepared.changedColumns,
		ReviewHash: token, RequiredConfirmation: requiredConfirmation(request.Action, schema.Name),
		RescueBackupRequired: true, Warning: bypassWarning,
	}
	if prepared.current != nil {
		review.Before = cloneValues(prepared.current.record.Values)
	}
	if request.Action != "delete" {
		keyRaw := keyRawFromRecord(schema, prepared.afterRaw, prepared.current)
		record, err := s.makeRecord(schema, prepared.afterRaw, keyRaw)
		if err != nil {
			return MutationReview{}, err
		}
		review.After = record.Values
		if request.Action == "insert" {
			// Omitted insert columns are populated by SQLite defaults at apply
			// time. Do not misrepresent them as NULL in the review; show only
			// the explicit typed values the administrator is submitting.
			for _, column := range schema.ColumnsRaw {
				if _, supplied := request.Values[column.Name]; !supplied {
					delete(review.After, column.Name)
				}
			}
		}
	}
	return review, nil
}

// ApplyMutation consumes an exact review, creates a rescue backup, rechecks the
// optimistic record hash, and performs the edit plus audit in one transaction.
func (s *Service) ApplyMutation(ctx context.Context, actorUserID string, request MutationRequest) (MutationResult, error) {
	s.mutation.Lock()
	defer s.mutation.Unlock()
	if err := validateAction(request.Action); err != nil {
		return MutationResult{}, err
	}
	reason, err := normalizeReason(request.Reason)
	if err != nil {
		return MutationResult{}, err
	}
	request.Reason = reason
	if request.Confirmation != requiredConfirmation(request.Action, request.Table) {
		return MutationResult{}, ErrConfirmation
	}
	if strings.TrimSpace(request.ReviewHash) == "" {
		return MutationResult{}, ErrReviewConflict
	}
	schema, err := s.schema(ctx, request.Table)
	if err != nil {
		return MutationResult{}, err
	}
	digest, err := mutationDigest(actorUserID, request)
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.checkReview(ctx, actorUserID, request, digest); err != nil {
		return MutationResult{}, err
	}
	if s.backups == nil {
		return MutationResult{}, errors.New("database rescue backup service is unavailable")
	}
	rescue, err := s.backups.Run(ctx)
	if err != nil {
		return MutationResult{}, fmt.Errorf("create database editor rescue backup: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin direct database edit: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := consumeReview(ctx, tx, actorUserID, request, digest, s.now().UTC()); err != nil {
		return MutationResult{}, err
	}
	prepared, err := s.prepare(ctx, tx, schema, request, true)
	if err != nil {
		return MutationResult{}, err
	}
	result, err := s.execute(ctx, tx, request.Action, prepared)
	if err != nil {
		return MutationResult{}, err
	}
	result.RescueBackupID = rescue.ID
	if err := s.auditMutation(ctx, tx, actorUserID, request, prepared, result, rescue.ID); err != nil {
		return MutationResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return MutationResult{}, fmt.Errorf("commit direct database edit: %w", err)
	}
	s.logger.Warn("direct database edit applied",
		"actor_user_id", actorUserID, "table", schema.Name, "action", request.Action,
		"changed_columns", prepared.changedColumns, "reason", "[redacted]", "values", "[redacted]",
		"rescue_backup_id", rescue.ID, "bypasses_domain_hooks", true)
	return result, nil
}

func validateAction(action string) error {
	switch action {
	case "insert", "update", "delete":
		return nil
	default:
		return fmt.Errorf("%w: action must be insert, update, or delete", ErrInvalidValue)
	}
}

func requiredConfirmation(action, table string) string {
	if action == "delete" {
		return "DELETE " + table
	}
	return "EDIT " + table
}

func mutationDigest(actor string, request MutationRequest) (string, error) {
	request.ReviewHash = ""
	request.Confirmation = ""
	payload := struct {
		Actor   string          `json:"actor"`
		Request MutationRequest `json:"request"`
	}{Actor: actor, Request: request}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode database mutation review: %w", err)
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

func (s *Service) checkReview(ctx context.Context, actor string, request MutationRequest, digest string) error {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM database_admin_reviews
		WHERE id=? AND actor_user_id=? AND action=? AND table_name=? AND request_hash=? AND consumed_at IS NULL AND expires_at>=?`,
		request.ReviewHash, actor, request.Action, request.Table, digest, s.now().UTC().Format(time.RFC3339Nano)).Scan(&count)
	if err != nil {
		return fmt.Errorf("verify database review: %w", err)
	}
	if count != 1 {
		return ErrReviewConflict
	}
	return nil
}

func consumeReview(ctx context.Context, tx *sql.Tx, actor string, request MutationRequest, digest string, now time.Time) error {
	result, err := tx.ExecContext(ctx, `UPDATE database_admin_reviews SET consumed_at=?
		WHERE id=? AND actor_user_id=? AND action=? AND table_name=? AND request_hash=? AND consumed_at IS NULL AND expires_at>=?`,
		now.Format(time.RFC3339Nano), request.ReviewHash, actor, request.Action, request.Table, digest, now.Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("consume database review: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count consumed database review: %w", err)
	}
	if affected != 1 {
		return ErrReviewConflict
	}
	return nil
}

func (s *Service) prepare(ctx context.Context, querier queryRower, schema tableSchema, request MutationRequest, encryptSettings bool) (preparedMutation, error) {
	prepared := preparedMutation{schema: schema}
	switch request.Action {
	case "update", "delete":
		current, err := s.readRecord(ctx, querier, schema, request.Key)
		if err != nil {
			return preparedMutation{}, err
		}
		if request.ExpectedHash == "" || current.record.RecordHash != request.ExpectedHash {
			return preparedMutation{}, ErrOptimisticConflict
		}
		prepared.current = &current
	case "insert":
		if len(request.Key) != 0 || request.ExpectedHash != "" {
			return preparedMutation{}, fmt.Errorf("%w: insert must not contain an existing key or record hash", ErrInvalidValue)
		}
	}
	if request.Action == "delete" {
		if len(request.Values) != 0 {
			return preparedMutation{}, fmt.Errorf("%w: delete must not contain values", ErrInvalidValue)
		}
		for _, column := range schema.ColumnsRaw {
			prepared.changedColumns = append(prepared.changedColumns, column.Name)
		}
		return prepared, nil
	}
	columns, arguments, afterRaw, changed, err := s.prepareValues(schema, prepared.current, request.Values, request.Action == "insert", encryptSettings)
	if err != nil {
		return preparedMutation{}, err
	}
	prepared.columns, prepared.arguments, prepared.afterRaw, prepared.changedColumns = columns, arguments, afterRaw, changed
	return prepared, nil
}

func (s *Service) readRecord(ctx context.Context, querier queryRower, schema tableSchema, key map[string]string) (rowData, error) {
	keyValues, err := s.decodePublicKey(schema, key)
	if err != nil {
		return rowData{}, err
	}
	query := `SELECT ` + selectList(schema) + ` FROM ` + quoteIdentifier(schema.Name) + ` WHERE ` + keyTuple(schema) + `=(` + placeholders(len(keyValues)) + `) LIMIT 1`
	data, err := s.scanRecord(schema, querier.QueryRowContext(ctx, query, keyValues...))
	if errors.Is(err, sql.ErrNoRows) || strings.Contains(errString(err), "sql: no rows") {
		return rowData{}, ErrRecordNotFound
	}
	return data, err
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (s *Service) prepareValues(schema tableSchema, current *rowData, requested map[string]Value, insert, encryptSettings bool) ([]schemaColumn, []any, []any, []string, error) {
	if len(requested) == 0 && !insert {
		return nil, nil, nil, nil, fmt.Errorf("%w: update contains no values", ErrInvalidValue)
	}
	byName := make(map[string]schemaColumn, len(schema.ColumnsRaw))
	for _, column := range schema.ColumnsRaw {
		byName[column.Name] = column
	}
	parsed := make(map[string]any, len(requested))
	for name, value := range requested {
		if !identifierPattern.MatchString(name) {
			return nil, nil, nil, nil, fmt.Errorf("%w: unsafe column name", ErrInvalidValue)
		}
		column, ok := byName[name]
		if !ok || !column.Editable {
			return nil, nil, nil, nil, fmt.Errorf("%w: column %s is not editable", ErrInvalidValue, name)
		}
		parsedValue, err := parseEditorValue(column, value)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		parsed[name] = parsedValue
	}
	afterRaw := make([]any, len(schema.ColumnsRaw))
	if current != nil {
		copy(afterRaw, current.raw)
	}
	if insert {
		for index, column := range schema.ColumnsRaw {
			if value, ok := parsed[column.Name]; ok {
				afterRaw[index] = value
				continue
			}
			if column.Hidden != 0 || column.Default.Valid || column.Nullable {
				continue
			}
			return nil, nil, nil, nil, fmt.Errorf("%w: required column %s is missing", ErrInvalidValue, column.Name)
		}
	} else {
		for index, column := range schema.ColumnsRaw {
			if value, ok := parsed[column.Name]; ok {
				afterRaw[index] = value
			}
		}
	}
	if schema.Name == "settings" {
		if err := s.prepareEncryptedSetting(schema, current, requested, parsed, afterRaw, insert, encryptSettings); err != nil {
			return nil, nil, nil, nil, err
		}
	}
	columns := make([]schemaColumn, 0, len(parsed))
	arguments := make([]any, 0, len(parsed))
	changed := make([]string, 0, len(parsed))
	for index, column := range schema.ColumnsRaw {
		value, supplied := parsed[column.Name]
		if !supplied {
			continue
		}
		columns = append(columns, column)
		arguments = append(arguments, value)
		if insert || current == nil || canonicalRaw(current.raw[index]) != canonicalRaw(afterRaw[index]) || column.Sensitive {
			changed = append(changed, column.Name)
		}
	}
	if len(changed) == 0 {
		if insert && len(parsed) == 0 {
			return columns, arguments, afterRaw, changed, nil
		}
		return nil, nil, nil, nil, fmt.Errorf("%w: mutation does not change any values", ErrInvalidValue)
	}
	sort.Strings(changed)
	return columns, arguments, afterRaw, changed, nil
}

func (s *Service) prepareEncryptedSetting(schema tableSchema, current *rowData, requested map[string]Value, parsed map[string]any, afterRaw []any, insert, encrypt bool) error {
	keyIndex, valueIndex, encryptedIndex := -1, -1, -1
	for index, column := range schema.ColumnsRaw {
		switch column.Name {
		case "key":
			keyIndex = index
		case "value":
			valueIndex = index
		case "encrypted":
			encryptedIndex = index
		}
	}
	if keyIndex < 0 || valueIndex < 0 || encryptedIndex < 0 {
		return errors.New("settings schema is incomplete")
	}
	wasEncrypted := current != nil && integerBool(current.raw[encryptedIndex])
	willBeEncrypted := integerBool(afterRaw[encryptedIndex])
	if wasEncrypted && !willBeEncrypted {
		return fmt.Errorf("%w: encrypted settings cannot be converted to plaintext through the database editor", ErrInvalidValue)
	}
	value, valueSupplied := requested["value"]
	keyChanged := current != nil && canonicalRaw(current.raw[keyIndex]) != canonicalRaw(afterRaw[keyIndex])
	if willBeEncrypted && !valueSupplied && (insert || !wasEncrypted || keyChanged) {
		return fmt.Errorf("%w: an encrypted setting needs a write-only value when its key or encryption state changes", ErrInvalidValue)
	}
	if !willBeEncrypted || !valueSupplied {
		return nil
	}
	plaintext, ok := value.Text()
	if !ok {
		return fmt.Errorf("%w: encrypted setting replacement must be plaintext", ErrInvalidValue)
	}
	key, ok := afterRaw[keyIndex].(string)
	if !ok || strings.TrimSpace(key) == "" {
		return fmt.Errorf("%w: encrypted setting key is required", ErrInvalidValue)
	}
	stored := "[write-only replacement]"
	if encrypt {
		if s.vault == nil {
			return errors.New("vault is unavailable for encrypted setting replacement")
		}
		ciphertext, err := s.vault.Encrypt(key, plaintext)
		if err != nil {
			return fmt.Errorf("encrypt setting replacement: %w", err)
		}
		stored = ciphertext
	}
	parsed["value"] = stored
	afterRaw[valueIndex] = stored
	return nil
}

func integerBool(value any) bool {
	switch typed := value.(type) {
	case int64:
		return typed == 1
	case bool:
		return typed
	case string:
		return typed == "1" || strings.EqualFold(typed, "true")
	default:
		return false
	}
}

func parseEditorValue(column schemaColumn, value Value) (any, error) {
	if value.kind == valueNull {
		if !column.Nullable {
			return nil, fmt.Errorf("%w: %s cannot be NULL", ErrInvalidValue, column.Name)
		}
		return nil, nil
	}
	if column.Boolean {
		if boolean, ok := value.Bool(); ok {
			if boolean {
				return int64(1), nil
			}
			return int64(0), nil
		}
		if text, ok := value.Text(); ok {
			switch strings.ToLower(strings.TrimSpace(text)) {
			case "1", "true":
				return int64(1), nil
			case "0", "false":
				return int64(0), nil
			}
		}
		return nil, fmt.Errorf("%w: %s must be boolean", ErrInvalidValue, column.Name)
	}
	switch column.Affinity {
	case affinityInteger:
		text, ok := value.Text()
		if !ok {
			return nil, fmt.Errorf("%w: %s must be a decimal integer string", ErrInvalidValue, column.Name)
		}
		parsed, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: %s must be a decimal integer string", ErrInvalidValue, column.Name)
		}
		return parsed, nil
	case affinityReal:
		text, ok := value.Text()
		if !ok || !decimalPattern.MatchString(text) {
			return nil, fmt.Errorf("%w: %s must be a finite decimal string", ErrInvalidValue, column.Name)
		}
		parsed, err := strconv.ParseFloat(text, 64)
		if err != nil || math.IsInf(parsed, 0) || math.IsNaN(parsed) {
			return nil, fmt.Errorf("%w: %s must be a finite decimal string", ErrInvalidValue, column.Name)
		}
		return parsed, nil
	case affinityNumeric:
		text, ok := value.Text()
		if !ok || !decimalPattern.MatchString(text) {
			return nil, fmt.Errorf("%w: %s must be a decimal string", ErrInvalidValue, column.Name)
		}
		return text, nil
	case affinityBlob:
		blob, ok := value.Blob()
		if !ok {
			return nil, fmt.Errorf("%w: %s must use blobBase64", ErrInvalidValue, column.Name)
		}
		return blob, nil
	default:
		text, ok := value.Text()
		if !ok {
			return nil, fmt.Errorf("%w: %s must be text", ErrInvalidValue, column.Name)
		}
		return text, nil
	}
}

func keyRawFromRecord(schema tableSchema, raw []any, current *rowData) []any {
	if schema.UsesRowID {
		if current != nil {
			return append([]any(nil), current.keyRaw...)
		}
		return []any{int64(0)}
	}
	keys := make([]any, 0, len(schema.KeyColumns))
	for _, key := range schema.KeyColumns {
		for index, column := range schema.ColumnsRaw {
			if column.Name == key.Name {
				keys = append(keys, raw[index])
				break
			}
		}
	}
	return keys
}

func (s *Service) execute(ctx context.Context, tx *sql.Tx, action string, prepared preparedMutation) (MutationResult, error) {
	schema := prepared.schema
	switch action {
	case "insert":
		query := `INSERT INTO ` + quoteIdentifier(schema.Name)
		if len(prepared.columns) == 0 {
			query += ` DEFAULT VALUES`
		} else {
			names := make([]string, 0, len(prepared.columns))
			for _, column := range prepared.columns {
				names = append(names, quoteIdentifier(column.Name))
			}
			query += ` (` + strings.Join(names, `,`) + `) VALUES (` + placeholders(len(names)) + `)`
		}
		query += ` RETURNING ` + selectList(schema)
		data, err := s.scanRecord(schema, tx.QueryRowContext(ctx, query, prepared.arguments...))
		if err != nil {
			return MutationResult{}, fmt.Errorf("insert %s record: %w", schema.Name, err)
		}
		return MutationResult{Row: &data.record}, nil
	case "update":
		sets := make([]string, 0, len(prepared.columns))
		for _, column := range prepared.columns {
			sets = append(sets, quoteIdentifier(column.Name)+`=?`)
		}
		query := `UPDATE ` + quoteIdentifier(schema.Name) + ` SET ` + strings.Join(sets, `,`) + ` WHERE ` + keyTuple(schema) + `=(` + placeholders(len(prepared.current.keyRaw)) + `) RETURNING ` + selectList(schema)
		args := append(append([]any(nil), prepared.arguments...), prepared.current.keyRaw...)
		data, err := s.scanRecord(schema, tx.QueryRowContext(ctx, query, args...))
		if err != nil {
			return MutationResult{}, fmt.Errorf("update %s record: %w", schema.Name, err)
		}
		return MutationResult{Row: &data.record}, nil
	case "delete":
		query := `DELETE FROM ` + quoteIdentifier(schema.Name) + ` WHERE ` + keyTuple(schema) + `=(` + placeholders(len(prepared.current.keyRaw)) + `)`
		result, err := tx.ExecContext(ctx, query, prepared.current.keyRaw...)
		if err != nil {
			return MutationResult{}, fmt.Errorf("delete %s record: %w", schema.Name, err)
		}
		affected, err := result.RowsAffected()
		if err != nil || affected != 1 {
			return MutationResult{}, ErrOptimisticConflict
		}
		return MutationResult{Deleted: true}, nil
	default:
		return MutationResult{}, ErrInvalidValue
	}
}

func (s *Service) auditMutation(ctx context.Context, tx *sql.Tx, actor string, request MutationRequest, prepared preparedMutation, result MutationResult, rescueBackupID string) error {
	auditID, err := ids.New()
	if err != nil {
		return fmt.Errorf("create database edit audit identifier: %w", err)
	}
	detail := map[string]any{
		"action": request.Action, "table": request.Table, "key": request.Key,
		"changedColumns": prepared.changedColumns, "reason": request.Reason, "rescueBackupId": rescueBackupID,
		"beforeHash": "", "afterHash": "", "values": "[redacted]", "warning": bypassWarning,
	}
	if prepared.current != nil {
		detail["beforeHash"] = prepared.current.record.RecordHash
	}
	if result.Row != nil {
		detail["afterHash"] = result.Row.RecordHash
	}
	encoded, err := json.Marshal(detail)
	if err != nil {
		return fmt.Errorf("encode database edit audit: %w", err)
	}
	targetID := request.Table
	if len(request.Key) > 0 {
		keyData, _ := json.Marshal(request.Key)
		targetID = request.Table + ":" + string(keyData)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO audit_events(id,actor_user_id,action,target_type,target_id,detail,created_at) VALUES(?,?,?,?,?,?,?)`,
		auditID, actor, "database_record_"+request.Action, "database_table", targetID, string(encoded), s.now().UTC().Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("record database edit audit: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM audit_events WHERE id IN (SELECT id FROM audit_events ORDER BY created_at DESC,id DESC LIMIT -1 OFFSET 200)`); err != nil {
		return fmt.Errorf("retain database edit audits: %w", err)
	}
	return nil
}

func cloneKey(source map[string]string) map[string]string {
	if source == nil {
		return nil
	}
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func cloneValues(source map[string]Value) map[string]Value {
	if source == nil {
		return nil
	}
	result := make(map[string]Value, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
