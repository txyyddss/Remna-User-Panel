package database

import (
	"database/sql"
	"errors"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"math/big"
)

func scanRollover(row rowScanner) (model.PurchaseRollover, error) {
	var value model.PurchaseRollover
	var allocated, used, eligible, remaining sql.NullInt64
	var created, updated string
	var completed sql.NullString
	if err := row.Scan(&value.PurchaseID, &value.Status, &value.TrafficLimitBytes, &allocated, &used, &eligible, &remaining, &value.MinimumRemainingBPS,
		&value.MaximumTXBMinor, &value.NetPaidTXBMinor, &value.CreditedTXBMinor, &value.ExceptionCode, &value.Attempts, &created, &updated, &completed, &value.AlgorithmVersion); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PurchaseRollover{}, ErrNotFound
		}
		return model.PurchaseRollover{}, err
	}
	if used.Valid {
		value.UsedTrafficBytes = &used.Int64
	}
	if allocated.Valid {
		value.AllocatedBytes = &allocated.Int64
	}
	if eligible.Valid {
		value.EligibleUnusedBytes = &eligible.Int64
	}
	if remaining.Valid {
		value.RemainingBytes = &remaining.Int64
	}
	var err error
	if value.CreatedAt, err = parseStamp(created); err != nil {
		return model.PurchaseRollover{}, err
	}
	if value.UpdatedAt, err = parseStamp(updated); err != nil {
		return model.PurchaseRollover{}, err
	}
	if completed.Valid {
		parsed, parseErr := parseStamp(completed.String)
		if parseErr != nil {
			return model.PurchaseRollover{}, parseErr
		}
		value.CompletedAt = &parsed
	}
	return value, nil
}

func strictlyAboveBPS(remaining, limit int64, threshold int) bool {
	left := new(big.Int).Mul(big.NewInt(remaining), big.NewInt(10000))
	right := new(big.Int).Mul(big.NewInt(limit), big.NewInt(int64(threshold)))
	return left.Cmp(right) > 0
}

func proportionalFloor(paid, remaining, limit int64) int64 {
	if paid <= 0 || remaining <= 0 || limit <= 0 {
		return 0
	}
	value := new(big.Int).Mul(big.NewInt(paid), big.NewInt(remaining))
	value.Quo(value, big.NewInt(limit))
	if !value.IsInt64() {
		return paid
	}
	return value.Int64()
}
