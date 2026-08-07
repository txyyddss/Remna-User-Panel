// Package ids creates random UUID-compatible identifiers and opaque tokens.
package ids

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// New returns a random RFC 4122 version 4 UUID string.
func New() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("read random UUID: %w", err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}

// Token returns a URL-safe cryptographically random token containing byteCount bytes of entropy.
func Token(byteCount int) (string, error) {
	if byteCount < 16 {
		return "", fmt.Errorf("token byte count must be at least 16")
	}
	value := make([]byte, byteCount)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("read random token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}
