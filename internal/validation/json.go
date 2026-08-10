package validation

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

const maxJSONDepth = 64

// JSONDocument validates UTF-8, structure, object keys, and every JSON string.
// Printable passwords, Telegram initData, and multiline or CSV text remain valid.
func JSONDocument(document []byte) error {
	if !utf8.Valid(document) {
		return &FieldError{Field: "JSON document", Err: ErrInvalidUTF8}
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return fmt.Errorf("decode JSON validation document: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("JSON document contains multiple values")
	}
	return validateJSONValue(value, 0)
}

func validateJSONValue(value any, depth int) error {
	if depth > maxJSONDepth {
		return &FieldError{Field: "JSON document", Err: ErrTooLong}
	}
	switch typed := value.(type) {
	case string:
		return Text("JSON string", typed, -1, true)
	case []any:
		for _, item := range typed {
			if err := validateJSONValue(item, depth+1); err != nil {
				return err
			}
		}
	case map[string]any:
		for key, item := range typed {
			if err := Text("JSON object key", key, 256, false); err != nil {
				return err
			}
			if err := validateJSONValue(item, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}
