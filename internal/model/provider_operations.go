package model

import "time"

// OperationReceipt is the public, non-sensitive state of a durable provider operation.
type OperationReceipt struct {
	ID          string     `json:"id"`
	Kind        string     `json:"kind"`
	Status      string     `json:"status"`
	ErrorCode   *string    `json:"errorCode"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	CompletedAt *time.Time `json:"completedAt"`
}
