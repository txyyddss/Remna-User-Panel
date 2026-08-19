package connections

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"strings"
)

func (s *Signer) ipDigest(rawIP string) (string, string, error) {
	parsed := net.ParseIP(strings.TrimSpace(rawIP))
	if s == nil || parsed == nil {
		return "", "", errors.New("invalid connection IP address")
	}
	canonical := parsed.String()
	mac := hmac.New(sha256.New, s.key)
	_, _ = mac.Write([]byte("connection-ip-block:v1:" + canonical))
	return canonical, hex.EncodeToString(mac.Sum(nil)), nil
}

func ipBlockSecretContext(userID, nodeUUID, digest string) string {
	return "connection-ip-block:" + userID + ":" + nodeUUID + ":" + digest
}

// dropSecretContext remains stable for legacy connection_drop jobs queued before this upgrade.
func dropSecretContext(userID, fingerprint string) string {
	return "connection-drop:" + userID + ":" + fingerprint
}
