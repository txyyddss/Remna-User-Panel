package abuse

import (
	"context"
	"errors"
	"time"
)

var ErrInvalid = errors.New("invalid abuse detector input")

type Repository interface {
	NodeByDigest(context.Context, string) (NodeCredential, error)
	TouchNodeReport(context.Context, string, time.Time) error
	KnownUsers(context.Context) (map[string]string, map[string]bool, error)
	Policy(context.Context) (Policy, error)
	DomainRules(context.Context) ([]DomainRule, error)
	StoreSamples(context.Context, string, []string, []Sample, time.Time) (ReportCounts, error)
	ReadyBuckets(context.Context, time.Time) ([]Sample, error)
	DetectorState(context.Context, string, string) (time.Time, int, error)
	SaveDetectorState(context.Context, string, string, time.Time, int) error
	CreateIncident(context.Context, string, time.Time, int, int, []string, []string, Policy, time.Time) (bool, error)
	DueTemporaryBans(context.Context, time.Time) ([]string, error)
	QueueRestore(context.Context, string, time.Time) error
	MemberRecords(context.Context, string, string, int) (RecordPage, error)
	NodeCredentials(context.Context) ([]NodeCredential, error)
	SaveNodeCredential(context.Context, Node, string, string, time.Time) error
	CopyNodeCredential(context.Context, string) (string, error)
	UpdatePolicy(context.Context, string, Policy, time.Time) (Policy, error)
	SaveDomainRule(context.Context, string, DomainRule, time.Time) (DomainRule, error)
	DeleteDomainRule(context.Context, string, string, int, time.Time) error
	Whitelist(context.Context) ([]string, error)
	SetWhitelist(context.Context, string, bool, time.Time) error
	PunishmentRules(context.Context) ([]PunishmentRule, error)
	SavePunishmentRule(context.Context, string, PunishmentRule, time.Time) (PunishmentRule, error)
	Statistics(context.Context, time.Time) (map[string]float64, error)
	DeleteRecord(context.Context, string, string, time.Time) error
}

type NodeProvider interface {
	AbuseNodes(context.Context) ([]Node, error)
}
