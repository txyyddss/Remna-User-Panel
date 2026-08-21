package model

import "time"

// TrafficResetAutomation is the account-wide automatic traffic reset preference.
type TrafficResetAutomation struct {
	Enabled   bool      `json:"enabled"`
	UpdatedAt time.Time `json:"updatedAt"`
}
