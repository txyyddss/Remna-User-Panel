// Package abuse implements privacy-safe QPS abuse detection.
package abuse

import "time"

const (
	GracePeriod    = 5 * time.Minute
	StreakEvery    = 30
	MaxReportBytes = 16 << 20
)

type Action string

const (
	ActionWarning      Action = "warning"
	ActionIPBan        Action = "ip_ban"
	ActionRevoke       Action = "subscription_revoke"
	ActionTemporaryBan Action = "temporary_ban"
)

type DomainRule struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Expression string `json:"expression"`
	QPSLimit   int    `json:"qpsLimit"`
	Enabled    bool   `json:"enabled"`
	Revision   int    `json:"revision"`
}
type Policy struct {
	GlobalEnabled       bool `json:"globalEnabled"`
	GlobalLimit         int  `json:"globalLimit"`
	WarningValidityDays int  `json:"warningValidityDays"`
	Revision            int  `json:"revision"`
}
type PunishmentRule struct {
	Action            Action `json:"action"`
	Enabled           bool   `json:"enabled"`
	IncidentThreshold int    `json:"incidentThreshold"`
	DurationMinutes   int    `json:"durationMinutes"`
	Revision          int    `json:"revision"`
	AllNodes          bool   `json:"allNodes"`
}
type Node struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}
type NodeCredential struct {
	Node
	LastReportAt *time.Time `json:"lastReportAt"`
	RotatedAt    time.Time  `json:"rotatedAt"`
}
type Sample struct {
	UserID, NodeUUID, ReasonName, Fingerprint string
	BucketAt                                  time.Time
	QPSLimit                                  int
	Count                                     int
}
type ReportCounts struct {
	Accepted  int `json:"accepted"`
	Duplicate int `json:"duplicate"`
	Discarded int `json:"discarded"`
}
type Record struct {
	ID          string     `json:"id"`
	Reason      string     `json:"reason"`
	OccurredAt  time.Time  `json:"occurredAt"`
	MeasuredQPS int        `json:"measuredQps"`
	QPSLimit    int        `json:"qpsLimit"`
	Action      Action     `json:"action"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}
type RecordPage struct {
	Items      []Record `json:"items"`
	NextCursor string   `json:"nextCursor"`
}
