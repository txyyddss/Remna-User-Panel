// Package secret encrypts dashboard-managed credentials at rest.
package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

const version = "v1"

// Vault uses AES-256-GCM with a process bootstrap key.
type Vault struct {
	aead cipher.AEAD
}

// NewVault creates a credential vault from a 32-byte key.
func NewVault(key []byte) (*Vault, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("vault key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}
	return &Vault{aead: aead}, nil
}

// Encrypt seals plaintext and binds it to the setting name supplied as associated data.
func (v *Vault) Encrypt(settingName, plaintext string) (string, error) {
	nonce := make([]byte, v.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("create encryption nonce: %w", err)
	}
	ciphertext := v.aead.Seal(nil, nonce, []byte(plaintext), []byte(settingName))
	payload := append(nonce, ciphertext...)
	return version + ":" + base64.RawStdEncoding.EncodeToString(payload), nil
}

// Decrypt opens a value previously produced by Encrypt.
func (v *Vault) Decrypt(settingName, encoded string) (string, error) {
	parts := strings.SplitN(encoded, ":", 2)
	if len(parts) != 2 || parts[0] != version {
		return "", fmt.Errorf("unsupported encrypted value version")
	}
	payload, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("decode encrypted value: %w", err)
	}
	if len(payload) < v.aead.NonceSize() {
		return "", fmt.Errorf("encrypted value is truncated")
	}
	nonce, ciphertext := payload[:v.aead.NonceSize()], payload[v.aead.NonceSize():]
	plaintext, err := v.aead.Open(nil, nonce, ciphertext, []byte(settingName))
	if err != nil {
		return "", fmt.Errorf("decrypt value: %w", err)
	}
	return string(plaintext), nil
}
