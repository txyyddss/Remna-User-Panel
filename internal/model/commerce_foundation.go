package model

import "time"

// AddTXBBounds is the global inclusive amount range for balance top-ups.
type AddTXBBounds struct {
	MinimumTXBMinor int64     `json:"-"`
	MaximumTXBMinor int64     `json:"-"`
	Minimum         Money     `json:"minimum"`
	Maximum         Money     `json:"maximum"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

