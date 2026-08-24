// Package abuse implements privacy-safe QPS abuse detection.
package abuse

import "time"

const (
	GracePeriod               = 5 * time.Minute
	StreakEvery               = 30
	MaxReportBytes            = 16 << 20
	MaxReportEvents           = 10000
	MaxSamplesPerReport       = 20000
	MaxGlobalQPS              = 100000
	MaxWarningValidityDays    = 365
	MaxWarningCooldownMinutes = 525600
	MaxRuleTextLength         = 1024
	MaxRuleNameLength         = 120
	MaxRemoteIDLength         = 256
)

type Action string

const (
	ActionWarning      Action = "warning"
	ActionIPBan        Action = "ip_ban"
	ActionRevoke       Action = "subscription_revoke"
	ActionTemporaryBan Action = "temporary_ban"
)

func (action Action) Valid() bool {
	switch action {
	case ActionWarning, ActionIPBan, ActionRevoke, ActionTemporaryBan:
		return true
	default:
		return false
	}
}

type DomainRule struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Expression string `json:"expression"`
	QPSLimit   int    `json:"qpsLimit"`
	Enabled    bool   `json:"enabled"`
	Revision   int    `json:"revision"`
}
type Policy struct {
	GlobalEnabled          bool `json:"globalEnabled"`
	GlobalLimit            int  `json:"globalLimit"`
	WarningValidityDays    int  `json:"warningValidityDays"`
	WarningCooldownMinutes int  `json:"warningCooldownMinutes"`
	Revision               int  `json:"revision"`
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
	MeasuredQPS int        `json:"measuredQPS"`
	QPSLimit    int        `json:"qpsLimit"`
	Action      Action     `json:"action"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}
type RecordPage struct {
	Items      []Record `json:"items"`
	NextCursor string   `json:"nextCursor"`
}
