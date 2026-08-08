package activity

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// CryptoRandom uses crypto/rand for unbiased game and draw outcomes.
type CryptoRandom struct{}

// Int63n returns an unbiased integer in [0, upperBound).
func (CryptoRandom) Int63n(upperBound int64) (int64, error) {
	if upperBound <= 0 {
		return 0, fmt.Errorf("%w: random upper bound must be positive", ErrInvalidInput)
	}
	value, err := rand.Int(rand.Reader, big.NewInt(upperBound))
	if err != nil {
		return 0, fmt.Errorf("read cryptographic random value: %w", err)
	}
	return value.Int64(), nil
}
