package model

import (
	"errors"
	"math/big"
	"strings"
)

// ParseTXBMajor converts a human major-unit decimal to integer hundredths. It
// rejects exponent notation and values with more than two fractional digits.
func ParseTXBMajor(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "+") || strings.ContainsAny(raw, "eE") {
		return 0, errors.New("TXB must be a fixed decimal")
	}
	negative := strings.HasPrefix(raw, "-")
	if negative {
		raw = strings.TrimPrefix(raw, "-")
	}
	parts := strings.Split(raw, ".")
	if len(parts) > 2 || parts[0] == "" || (len(parts) == 2 && (parts[1] == "" || len(parts[1]) > 2)) {
		return 0, errors.New("TXB must have at most two fractional digits")
	}
	for _, part := range parts {
		for _, digit := range part {
			if digit < '0' || digit > '9' {
				return 0, errors.New("TXB contains an invalid digit")
			}
		}
	}
	fraction := "00"
	if len(parts) == 2 {
		fraction = parts[1] + strings.Repeat("0", 2-len(parts[1]))
	}
	minor := new(big.Int)
	if _, ok := minor.SetString(parts[0]+fraction, 10); !ok {
		return 0, errors.New("TXB amount is invalid")
	}
	if negative {
		minor.Neg(minor)
	}
	if !minor.IsInt64() {
		return 0, errors.New("TXB amount is out of range")
	}
	return minor.Int64(), nil
}
