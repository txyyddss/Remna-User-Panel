package iplookup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

// IDs defines the immutable execution order of the built-in providers.
var IDs = func() []string {
	ids := []string{}
	for _, provider := range Registry {
		ids = append(ids, provider.ID)
	}
	return ids
}()

// CodeError exposes only a stable member-safe error identifier.
type CodeError struct{ Code string }

func (e *CodeError) Error() string { return e.Code }

// CredentialKey binds encryption to the exact provider setting.
func CredentialKey(id string) string { return "ip_lookup." + id + ".credential" }

// DefaultConfig leaves prices unconfigured and all providers disabled.
func DefaultConfig() Config {
	c := Config{Providers: []ProviderConfig{}}
	for _, id := range IDs {
		c.Providers = append(c.Providers, ProviderConfig{ID: id})
	}
	return c
}

// DecodeConfig validates and canonicalizes the provider order.
func DecodeConfig(raw string) (Config, error) {
	if raw == "" {
		return DefaultConfig(), nil
	}
	var c Config
	if json.Unmarshal([]byte(raw), &c) != nil {
		return c, &CodeError{"IP_LOOKUP_INVALID_CONFIG"}
	}
	for _, fee := range []string{c.LookupFeeTXB, c.RefreshFeeTXB} {
		if fee != "" {
			if amount, err := model.ParseTXBMajor(fee); err != nil || amount < 0 || amount > 1_000_000_000_000 {
				return c, &CodeError{"IP_LOOKUP_INVALID_CONFIG"}
			}
		}
	}
	byID := map[string]ProviderConfig{}
	for _, p := range c.Providers {
		known := false
		for _, id := range IDs {
			if p.ID == id {
				known = true
			}
		}
		if _, duplicate := byID[p.ID]; duplicate || !known {
			return c, &CodeError{"IP_LOOKUP_INVALID_CONFIG"}
		}
		if p.ID == "maxmind" && p.Enabled {
			if p.AccountID == "" || strings.Trim(p.AccountID, "0123456789") != "" {
				return c, &CodeError{"IP_LOOKUP_INVALID_CONFIG"}
			}
		}
		if p.ID == "scamalytics" && p.Enabled {
			if p.AccountID == "" || len(p.AccountID) > 80 || strings.Trim(p.AccountID, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-") != "" {
				return c, &CodeError{"IP_LOOKUP_INVALID_CONFIG"}
			}
		}
		byID[p.ID] = p
	}
	c.Providers = []ProviderConfig{}
	for _, id := range IDs {
		p := byID[id]
		p.ID = id
		c.Providers = append(c.Providers, p)
	}
	if c.Enabled {
		if c.LookupFeeTXB == "" || c.RefreshFeeTXB == "" {
			return c, &CodeError{"IP_LOOKUP_CONFIGURATION_REQUIRED"}
		}
		enabled := false
		for _, p := range c.Providers {
			enabled = enabled || p.Enabled
		}
		if !enabled {
			return c, &CodeError{"IP_LOOKUP_CONFIGURATION_REQUIRED"}
		}
	}
	return c, nil
}

// ConfigHash freezes public configuration in a signed quote.
func ConfigHash(c Config) string {
	b, _ := json.Marshal(c)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// ErrorCode prevents HTTP clients and queue logs from leaking credential-bearing URLs.
func ErrorCode(err error) string {
	var coded *CodeError
	if errors.As(err, &coded) {
		return coded.Code
	}
	return "IP_LOOKUP_FAILED"
}
