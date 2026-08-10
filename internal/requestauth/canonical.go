// Package requestauth authenticates browser requests bound to an opaque session.
package requestauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
)

const (
	// ClientKeyCookie is readable by the same-origin browser signing client.
	ClientKeyCookie = "txc_request_key"
	// TimestampHeader carries Unix seconds for the signed request.
	TimestampHeader = "X-TXC-Timestamp"
	// NonceHeader carries a unique base64url value for the signed request.
	NonceHeader = "X-TXC-Nonce"
	// SignatureHeader carries a lowercase hexadecimal HMAC-SHA256.
	SignatureHeader    = "X-TXC-Signature"
	keyDerivationLabel = "txc-request-v1\x00"
)

// CanonicalTarget returns the escaped path and exact raw query covered by a signature.
func CanonicalTarget(value *url.URL) string {
	path := "/"
	if value != nil && value.EscapedPath() != "" {
		path = value.EscapedPath()
	}
	if value != nil && value.RawQuery != "" {
		return path + "?" + value.RawQuery
	}
	return path
}

// Sign computes the browser-compatible signature for one canonical request.
func Sign(clientKey, method, target, timestamp, nonce string, body []byte) (string, error) {
	key, err := base64.RawURLEncoding.DecodeString(clientKey)
	if err != nil || len(key) != sha256.Size {
		return "", errors.New("client key must be unpadded base64url-encoded 32 bytes")
	}
	digest := sha256.Sum256(body)
	canonical := strings.Join([]string{
		strings.ToUpper(method), target, timestamp, nonce, hex.EncodeToString(digest[:]),
	}, "\n")
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(canonical))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func deriveClientKey(master []byte, sessionToken string) []byte {
	mac := hmac.New(sha256.New, master)
	_, _ = mac.Write([]byte(keyDerivationLabel))
	_, _ = mac.Write([]byte(sessionToken))
	return mac.Sum(nil)
}
