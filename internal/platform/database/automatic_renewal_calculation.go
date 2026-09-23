package database

import "github.com/txyyddss/Remna-User-Panel/internal/model"

func calculatedRolloverCredit(rollover model.PurchaseRollover) int64 {
	if rollover.AllocatedBytes == nil || rollover.EligibleUnusedBytes == nil {
		return 0
	}
	return proportionalFloor(rollover.NetPaidTXBMinor, *rollover.EligibleUnusedBytes, *rollover.AllocatedBytes)
}

func renewalFundsCover(balance, credit, price int64, rolloverCountsTowardBalance bool) bool {
	if !rolloverCountsTowardBalance {
		credit = 0
	}
	return price >= 0 && balance >= 0 && (balance >= price || (credit >= 0 && credit >= price-balance))
}
