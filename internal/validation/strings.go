// Package validation provides transport-safe validation for untrusted strings.
package validation

import (
	"errors"
	"fmt"
	"regexp"
	"unicode/utf8"
)

var (
	// JSON text may contain the JSON whitespace characters, while request-line
	// and header values must remain on one line.
	multilineTextPattern = regexp.MustCompile(`^[^\x00-\x08\x0B\x0C\x0E-\x1F\x7F]*$`)
	singleLinePattern    = regexp.MustCompile(`^[^\x00-\x1F\x7F]*$`)
	headerNamePattern    = regexp.MustCompile("^[!#$%&'*+.^_`|~0-9A-Za-z-]+$")
	escapedPathPattern   = regexp.MustCompile(`^/[A-Za-z0-9._~!$&'()*+,;=:@%/\-]*$`)
	decodedPathPattern   = regexp.MustCompile(`^[^\x00-\x1F\x7F?#]*$`)
	methodPattern        = regexp.MustCompile(`^[A-Z]{3,16}$`)
)

var (
	// ErrInvalidUTF8 marks a string that is not valid UTF-8.
	ErrInvalidUTF8 = errors.New("invalid UTF-8")
	// ErrControlCharacter marks a disallowed control character.
	ErrControlCharacter = errors.New("control character is not allowed")
	// ErrTooLong marks an input that exceeds its transport limit.
	ErrTooLong = errors.New("input is too long")
	// ErrPattern marks an input outside its field's regular-language grammar.
	ErrPattern = errors.New("input does not match the required pattern")
)

// FieldError identifies the input class without echoing the rejected value.
type FieldError struct {
	Field string
	Err   error
}

// Error implements error.
func (e *FieldError) Error() string { return fmt.Sprintf("validate %s: %v", e.Field, e.Err) }

// Unwrap exposes the validation category.
func (e *FieldError) Unwrap() error { return e.Err }

// Text validates printable UTF-8. Multiline values may use tab, CR, and LF.
func Text(field, value string, maxBytes int, multiline bool) error {
	if !utf8.ValidString(value) {
		return &FieldError{Field: field, Err: ErrInvalidUTF8}
	}
	if maxBytes >= 0 && len(value) > maxBytes {
		return &FieldError{Field: field, Err: ErrTooLong}
	}
	pattern := singleLinePattern
	if multiline {
		pattern = multilineTextPattern
	}
	if !pattern.MatchString(value) {
		return &FieldError{Field: field, Err: ErrControlCharacter}
	}
	return nil
}
