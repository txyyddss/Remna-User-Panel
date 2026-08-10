package databaseadmin

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

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
