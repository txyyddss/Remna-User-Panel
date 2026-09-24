package rollover

import "math/big"

// EligibleUnused applies the strict threshold to one whole purchase term.
func EligibleUnused(allocated, used int64, minimumRemainingBPS int) int64 {
	if allocated <= 0 || used < 0 || used > allocated || minimumRemainingBPS < 0 || minimumRemainingBPS > 10000 {
		return 0
	}
	remaining := allocated - used
	if !strictlyAboveBPS(remaining, allocated, minimumRemainingBPS) {
		return 0
	}
	return remaining
}

// CreditMinor rounds a proportional net-paid credit to the nearest TXB cent.
func CreditMinor(paid, eligible, allocated int64) int64 {
	if paid <= 0 || eligible <= 0 || allocated <= 0 {
		return 0
	}
	if eligible >= allocated {
		return paid
	}
	product := new(big.Int).Mul(big.NewInt(paid), big.NewInt(eligible))
	product.Add(product, big.NewInt(allocated/2))
	product.Quo(product, big.NewInt(allocated))
	if !product.IsInt64() {
		return paid
	}
	return product.Int64()
}
