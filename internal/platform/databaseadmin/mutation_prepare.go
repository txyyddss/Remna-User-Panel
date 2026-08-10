package databaseadmin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
)

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
