package ezpay

import (
	"crypto/md5" // #nosec G501 -- MD5 is mandated by the EZPay wire protocol.
	"encoding/hex"
	"errors"
	"math/big"
	"net/url"
	"sort"
	"strings"
)

func Sign(values url.Values, key string) string {
	keys := make([]string, 0, len(values))
	for name := range values {
		if name == "sign" || name == "sign_type" || values.Get(name) == "" {
			continue
		}
		keys = append(keys, name)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, name := range keys {
		parts = append(parts, name+"="+values.Get(name))
	}
	sum := md5.Sum([]byte(strings.Join(parts, "&") + key)) // #nosec G401 -- required for protocol compatibility.
	return hex.EncodeToString(sum[:])
}

func validateCallbackURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.User != nil {
		return errors.New("URL must be absolute http(s) without credentials")
	}
	return nil
}

func validatePositiveDecimal(value string) error {
	if len(value) > 64 {
		return errors.New("value is too long")
	}
	if !decimalPattern.MatchString(value) {
		return errors.New("value must be an unsigned base-10 decimal")
	}
	parsed, ok := new(big.Rat).SetString(value)
	if !ok || parsed.Sign() <= 0 {
		return errors.New("value must be positive")
	}
	return nil
}

