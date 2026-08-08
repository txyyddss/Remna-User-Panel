package databaseadmin

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type queryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type rowData struct {
	raw    []any
	keyRaw []any
	record Record
}

// Records returns a cursor page ordered by the declared primary key, or rowid
// only when the table has no declared key. Browser-supplied table names are
// resolved through schema introspection before any identifier is composed.
func (s *Service) Records(ctx context.Context, table, cursor string, limit int) (Page, error) {
	schema, err := s.schema(ctx, table)
	if err != nil {
		return Page{}, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	selectSQL := selectList(schema)
	query := `SELECT ` + selectSQL + ` FROM ` + quoteIdentifier(schema.Name)
	args := make([]any, 0)
	if strings.TrimSpace(cursor) != "" {
		keyValues, err := s.decodeCursor(schema, cursor)
		if err != nil {
			return Page{}, fmt.Errorf("%w: invalid record cursor", ErrInvalidValue)
		}
		query += ` WHERE ` + keyTuple(schema) + ` > (` + placeholders(len(keyValues)) + `)`
		args = append(args, keyValues...)
	}
	query += ` ORDER BY ` + keyOrder(schema) + ` LIMIT ?`
	args = append(args, limit+1)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return Page{}, fmt.Errorf("list %s records: %w", schema.Name, err)
	}
	defer rows.Close()
	data := make([]rowData, 0, limit+1)
	for rows.Next() {
		item, err := s.scanRecord(schema, rows)
		if err != nil {
			return Page{}, err
		}
		data = append(data, item)
	}
	if err := rows.Err(); err != nil {
		return Page{}, fmt.Errorf("iterate %s records: %w", schema.Name, err)
	}
	page := Page{Items: make([]Record, 0, min(limit, len(data)))}
	visible := len(data)
	if visible > limit {
		visible = limit
	}
	for index := 0; index < visible; index++ {
		page.Items = append(page.Items, data[index].record)
	}
	if len(data) > limit && visible > 0 {
		next, err := s.encodeCursor(schema, data[visible-1].keyRaw)
		if err != nil {
			return Page{}, err
		}
		page.NextCursor = &next
	}
	return page, nil
}

func selectList(schema tableSchema) string {
	parts := make([]string, 0, len(schema.ColumnsRaw)+1)
	for _, column := range schema.ColumnsRaw {
		parts = append(parts, quoteIdentifier(column.Name))
	}
	if schema.UsesRowID {
		parts = append(parts, quoteIdentifier(schema.RowIDName))
	}
	return strings.Join(parts, `,`)
}

func keyTuple(schema tableSchema) string {
	if schema.UsesRowID {
		return `(` + quoteIdentifier(schema.RowIDName) + `)`
	}
	parts := make([]string, 0, len(schema.KeyColumns))
	for _, column := range schema.KeyColumns {
		parts = append(parts, quoteIdentifier(column.Name))
	}
	return `(` + strings.Join(parts, `,`) + `)`
}

func keyOrder(schema tableSchema) string {
	if schema.UsesRowID {
		return quoteIdentifier(schema.RowIDName) + ` ASC`
	}
	parts := make([]string, 0, len(schema.KeyColumns))
	for _, column := range schema.KeyColumns {
		parts = append(parts, quoteIdentifier(column.Name)+` ASC`)
	}
	return strings.Join(parts, `,`)
}

func placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.TrimSuffix(strings.Repeat("?,", count), ",")
}

func (s *Service) scanRecord(schema tableSchema, row interface{ Scan(...any) error }) (rowData, error) {
	values := make([]any, len(schema.ColumnsRaw)+boolInt(schema.UsesRowID))
	destinations := make([]any, len(values))
	for index := range values {
		destinations[index] = &values[index]
	}
	if err := row.Scan(destinations...); err != nil {
		return rowData{}, fmt.Errorf("scan %s record: %w", schema.Name, err)
	}
	keyRaw := make([]any, 0, max(1, len(schema.KeyColumns)))
	if schema.UsesRowID {
		keyRaw = append(keyRaw, values[len(values)-1])
	} else {
		for _, key := range schema.KeyColumns {
			for index, column := range schema.ColumnsRaw {
				if column.Name == key.Name {
					keyRaw = append(keyRaw, values[index])
					break
				}
			}
		}
	}
	record, err := s.makeRecord(schema, values[:len(schema.ColumnsRaw)], keyRaw)
	if err != nil {
		return rowData{}, err
	}
	return rowData{raw: values[:len(schema.ColumnsRaw)], keyRaw: keyRaw, record: record}, nil
}

func (s *Service) makeRecord(schema tableSchema, raw, keyRaw []any) (Record, error) {
	record := Record{Key: make(map[string]string), Values: make(map[string]Value, len(schema.ColumnsRaw))}
	for index, column := range schema.ColumnsRaw {
		if column.Sensitive {
			record.Values[column.Name] = maskedValue()
			continue
		}
		value, err := publicValue(column, raw[index])
		if err != nil {
			return Record{}, fmt.Errorf("render %s.%s: %w", schema.Name, column.Name, err)
		}
		record.Values[column.Name] = value
	}
	if schema.UsesRowID {
		encoded, err := canonicalKeyValue(schemaColumn{Column: Column{Name: "_rowid_"}, Affinity: affinityInteger}, keyRaw[0])
		if err != nil {
			return Record{}, err
		}
		record.Key["_rowid_"] = encoded
	} else {
		for index, column := range schema.KeyColumns {
			encoded, err := canonicalKeyValue(column, keyRaw[index])
			if err != nil {
				return Record{}, err
			}
			if column.Sensitive {
				if s.vault == nil {
					return Record{}, errors.New("vault is required to conceal a sensitive record key")
				}
				sealed, err := s.vault.Encrypt(keyContext(schema.Name, column.Name), encoded)
				if err != nil {
					return Record{}, fmt.Errorf("seal sensitive record key: %w", err)
				}
				encoded = "sealed:" + sealed
			}
			record.Key[column.Name] = encoded
		}
	}
	hash, err := recordHash(schema, raw, keyRaw)
	if err != nil {
		return Record{}, err
	}
	record.RecordHash = hash
	return record, nil
}

func publicValue(column schemaColumn, raw any) (Value, error) {
	if raw == nil {
		return NullValue(), nil
	}
	if column.Boolean {
		switch value := raw.(type) {
		case int64:
			if value == 0 || value == 1 {
				return BooleanValue(value == 1), nil
			}
		case bool:
			return BooleanValue(value), nil
		}
	}
	switch value := raw.(type) {
	case int64:
		return StringValue(strconv.FormatInt(value, 10)), nil
	case float64:
		return StringValue(strconv.FormatFloat(value, 'g', -1, 64)), nil
	case string:
		if column.Affinity == affinityBlob {
			return BlobValue([]byte(value)), nil
		}
		return StringValue(value), nil
	case []byte:
		if column.Affinity == affinityBlob {
			return BlobValue(value), nil
		}
		return StringValue(string(value)), nil
	case bool:
		return BooleanValue(value), nil
	default:
		return Value{}, fmt.Errorf("unsupported SQLite value %T", raw)
	}
}

func recordHash(schema tableSchema, raw, keyRaw []any) (string, error) {
	payload := struct {
		Table   string   `json:"table"`
		Columns []string `json:"columns"`
		Values  []string `json:"values"`
		Keys    []string `json:"keys"`
	}{Table: schema.Name}
	for index, column := range schema.ColumnsRaw {
		payload.Columns = append(payload.Columns, column.Name)
		payload.Values = append(payload.Values, canonicalRaw(raw[index]))
	}
	for _, value := range keyRaw {
		payload.Keys = append(payload.Keys, canonicalRaw(value))
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode record hash: %w", err)
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

func canonicalRaw(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null:"
	case int64:
		return "integer:" + strconv.FormatInt(typed, 10)
	case float64:
		return "real:" + strconv.FormatFloat(typed, 'g', -1, 64)
	case []byte:
		return "blob:" + base64.StdEncoding.EncodeToString(typed)
	case string:
		return "text:" + typed
	case bool:
		return "boolean:" + strconv.FormatBool(typed)
	default:
		return fmt.Sprintf("%T:%v", value, value)
	}
}

func canonicalKeyValue(column schemaColumn, raw any) (string, error) {
	if raw == nil {
		return "", fmt.Errorf("%w: primary key %s is NULL", ErrInvalidValue, column.Name)
	}
	switch value := raw.(type) {
	case int64:
		return strconv.FormatInt(value, 10), nil
	case float64:
		return strconv.FormatFloat(value, 'g', -1, 64), nil
	case string:
		return value, nil
	case []byte:
		return base64.StdEncoding.EncodeToString(value), nil
	default:
		return "", fmt.Errorf("%w: unsupported primary key type %T", ErrInvalidValue, raw)
	}
}

func (s *Service) decodePublicKey(schema tableSchema, key map[string]string) ([]any, error) {
	if schema.UsesRowID {
		if len(key) != 1 {
			return nil, fmt.Errorf("%w: rowid key shape differs", ErrInvalidValue)
		}
		value, ok := key["_rowid_"]
		if !ok {
			return nil, fmt.Errorf("%w: rowid key is required", ErrInvalidValue)
		}
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: rowid must be a decimal integer", ErrInvalidValue)
		}
		return []any{parsed}, nil
	}
	if len(key) != len(schema.KeyColumns) {
		return nil, fmt.Errorf("%w: primary key shape differs", ErrInvalidValue)
	}
	values := make([]any, 0, len(schema.KeyColumns))
	for _, column := range schema.KeyColumns {
		encoded, ok := key[column.Name]
		if !ok {
			return nil, fmt.Errorf("%w: primary key %s is required", ErrInvalidValue, column.Name)
		}
		if column.Sensitive {
			if s.vault == nil || !strings.HasPrefix(encoded, "sealed:") {
				return nil, fmt.Errorf("%w: sensitive primary key is not sealed", ErrInvalidValue)
			}
			plain, err := s.vault.Decrypt(keyContext(schema.Name, column.Name), strings.TrimPrefix(encoded, "sealed:"))
			if err != nil {
				return nil, fmt.Errorf("%w: sensitive primary key cannot be opened", ErrInvalidValue)
			}
			encoded = plain
		}
		parsed, err := parseKeyValue(column, encoded)
		if err != nil {
			return nil, err
		}
		values = append(values, parsed)
	}
	return values, nil
}

func parseKeyValue(column schemaColumn, encoded string) (any, error) {
	switch column.Affinity {
	case affinityInteger:
		value, err := strconv.ParseInt(encoded, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: key %s must be a decimal integer", ErrInvalidValue, column.Name)
		}
		return value, nil
	case affinityReal:
		value, err := strconv.ParseFloat(encoded, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: key %s must be numeric", ErrInvalidValue, column.Name)
		}
		return value, nil
	case affinityNumeric:
		// Keep NUMERIC keys as validated decimal text. Binding them as float64
		// would round exact 64-bit SQLite integers above 2^53 and could duplicate
		// cursor pages or address the wrong record. SQLite applies the column's
		// NUMERIC affinity to this bound text for comparison.
		if !decimalPattern.MatchString(encoded) {
			return nil, fmt.Errorf("%w: key %s must be numeric", ErrInvalidValue, column.Name)
		}
		return encoded, nil
	case affinityBlob:
		value, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("%w: key %s must be base64", ErrInvalidValue, column.Name)
		}
		return value, nil
	default:
		return encoded, nil
	}
}

func keyContext(table, column string) string { return "databaseadmin:key:" + table + ":" + column }

func (s *Service) encodeCursor(schema tableSchema, raw []any) (string, error) {
	values := make([]string, 0, len(raw))
	if schema.UsesRowID {
		value, err := canonicalKeyValue(schemaColumn{Column: Column{Name: "_rowid_"}, Affinity: affinityInteger}, raw[0])
		if err != nil {
			return "", err
		}
		values = append(values, value)
	} else {
		for index, column := range schema.KeyColumns {
			value, err := canonicalKeyValue(column, raw[index])
			if err != nil {
				return "", err
			}
			values = append(values, value)
		}
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return "", fmt.Errorf("encode record cursor: %w", err)
	}
	if s.vault != nil {
		sealed, err := s.vault.Encrypt("databaseadmin:cursor:"+schema.Name, string(encoded))
		if err != nil {
			return "", fmt.Errorf("seal record cursor: %w", err)
		}
		return "sealed:" + sealed, nil
	}
	return "plain:" + base64.RawURLEncoding.EncodeToString(encoded), nil
}

func (s *Service) decodeCursor(schema tableSchema, cursor string) ([]any, error) {
	var encoded []byte
	switch {
	case strings.HasPrefix(cursor, "sealed:") && s.vault != nil:
		plain, err := s.vault.Decrypt("databaseadmin:cursor:"+schema.Name, strings.TrimPrefix(cursor, "sealed:"))
		if err != nil {
			return nil, err
		}
		encoded = []byte(plain)
	case strings.HasPrefix(cursor, "plain:") && s.vault == nil:
		value, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(cursor, "plain:"))
		if err != nil {
			return nil, err
		}
		encoded = value
	default:
		return nil, errors.New("cursor format is invalid")
	}
	var values []string
	if err := json.Unmarshal(encoded, &values); err != nil {
		return nil, err
	}
	expected := len(schema.KeyColumns)
	if schema.UsesRowID {
		expected = 1
	}
	if len(values) != expected {
		return nil, errors.New("cursor key shape differs")
	}
	parsed := make([]any, 0, expected)
	if schema.UsesRowID {
		value, err := strconv.ParseInt(values[0], 10, 64)
		if err != nil {
			return nil, err
		}
		return []any{value}, nil
	}
	for index, column := range schema.KeyColumns {
		value, err := parseKeyValue(column, values[index])
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, value)
	}
	return parsed, nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
