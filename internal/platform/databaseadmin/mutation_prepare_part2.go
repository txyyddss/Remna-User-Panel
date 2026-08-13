package databaseadmin

import (
	"errors"
	"fmt"
	"strings"
)

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
