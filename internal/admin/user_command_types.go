package admin

import "time"

// EntitlementEdit replaces every exposed mutable entitlement field.
type EntitlementEdit struct {
	Version           time.Time
	Reason            string
	ComboID           string
	ValidFrom         time.Time
	ValidUntil        time.Time
	Status            string
	TrafficLimitBytes int64
	ResetStrategy     string
	SquadUUIDs        []string
}

// ComboReplacement changes active configuration without moving TXB.
type ComboReplacement struct {
	ComboID         string
	AddonSquadUUIDs []string
	Reason          string
}

// BulkExtension selects active users with inclusive OR matching.
type BulkExtension struct {
	ComboIDs        []string
	AddonSquadUUIDs []string
	Days            int
	Reason          string
}
