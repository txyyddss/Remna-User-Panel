package connectivity

import (
	"context"
	"errors"
	"regexp"
)

var safeErrorCode = regexp.MustCompile(`^CONNECTIVITY_[A-Z0-9_]{1,80}$`)

// CodeError carries a stable public code while preserving an internal cause.
// Error never exposes the cause, which can contain upstream credentials.
type CodeError struct {
	Code string
	Err  error
}

// Error returns only the sanitized diagnostic code.
func (err *CodeError) Error() string { return ErrorCode(err) }

// Unwrap retains cancellation and provider classification for internal handling.
func (err *CodeError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

// ErrorCode projects any internal error to a credential-free public code.
func ErrorCode(err error) string {
	if err == nil {
		return ""
	}
	var coded *CodeError
	if errors.As(err, &coded) && coded != nil && safeErrorCode.MatchString(coded.Code) {
		return coded.Code
	}
	if errors.Is(err, context.Canceled) {
		return "CONNECTIVITY_INTERRUPTED"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "CONNECTIVITY_PROBE_TIMEOUT"
	}
	return "CONNECTIVITY_UNAVAILABLE"
}
