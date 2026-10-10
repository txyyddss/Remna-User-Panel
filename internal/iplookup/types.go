// Package iplookup owns queued IP reputation checks and normalized reports.
package iplookup

import "github.com/txyyddss/Remna-User-Panel/internal/model"

const SettingKey = "ip_lookup.config"
const OperationKind = "ip_reputation_lookup"
const Version = "residential-v1"

// ProviderConfig is public configuration; credentials live in the secret vault.
type ProviderConfig struct {
	ID        string `json:"id"`
	Enabled   bool   `json:"enabled"`
	AccountID string `json:"accountId"`
}

// Config contains explicitly configured decimal TXB fees and ordered providers.
type Config struct {
	Enabled       bool             `json:"enabled"`
	LookupFeeTXB  string           `json:"lookupFeeTxb"`
	RefreshFeeTXB string           `json:"refreshFeeTxb"`
	Providers     []ProviderConfig `json:"providers"`
}

// CredentialEdit replaces or explicitly clears a write-only credential.
type CredentialEdit struct {
	Value string `json:"value"`
	Clear bool   `json:"clear"`
}

// AdminSettings is the atomic administrator configuration projection.
type AdminSettings struct {
	Config
	Credentials map[string]CredentialEdit `json:"credentials"`
	Configured  map[string]bool           `json:"configured"`
	ComboQuotas map[string]*int           `json:"comboQuotas"`
}

// Allowance belongs to one successfully activated purchase, never a calendar cycle.
type Allowance struct {
	PurchaseID string `json:"purchaseId"`
	Total      int    `json:"total"`
	Remaining  int    `json:"remaining"`
	ValidUntil string `json:"validUntil"`
}

// State is the safe member-facing availability, prices, and current allowance.
type State struct {
	Enabled    bool         `json:"enabled"`
	LookupFee  *model.Money `json:"lookupFee"`
	RefreshFee *model.Money `json:"refreshFee"`
	Allowance  *Allowance   `json:"allowance"`
}

// Quote binds an action to the price and allowance observed before submission.
type Quote struct {
	IP            string      `json:"ip"`
	Refresh       bool        `json:"refresh"`
	Charge        model.Money `json:"charge"`
	UseQuota      bool        `json:"useQuota"`
	PurchaseID    string      `json:"purchaseId"`
	Remaining     int         `json:"remaining"`
	CacheReportID string      `json:"cacheReportId"`
	ConfigHash    string      `json:"configHash"`
	ExpiresAt     int64       `json:"expiresAt"`
	Token         string      `json:"token"`
}

// Signals keeps unknown booleans distinct from negative evidence.
type Signals struct {
	Abuse      *bool `json:"abuse"`
	Datacenter *bool `json:"datacenter"`
	VPN        *bool `json:"vpn"`
	Proxy      *bool `json:"proxy"`
	Tor        *bool `json:"tor"`
}

// Facts is the small, source-attributed geographical and network projection.
type Facts struct {
	Country     string `json:"country"`
	Region      string `json:"region"`
	City        string `json:"city"`
	ISP         string `json:"isp"`
	ASN         string `json:"asn"`
	NetworkType string `json:"networkType"`
}

// ProviderResult retains parsed evidence, including missing subscription capabilities.
type ProviderResult struct {
	ID        string             `json:"id"`
	Status    string             `json:"status"`
	ErrorCode string             `json:"errorCode"`
	CheckedAt string             `json:"checkedAt"`
	Facts     Facts              `json:"facts"`
	Signals   Signals            `json:"signals"`
	Scores    map[string]float64 `json:"scores"`
	Reports   *int               `json:"reports"`
	Complete  bool               `json:"complete"`
}

// Report freezes the verdict and original coverage; toggles never rewrite it.
type Report struct {
	ID            string            `json:"id"`
	IP            string            `json:"ip"`
	Status        string            `json:"status"`
	Verdict       string            `json:"verdict"`
	Reasons       []string          `json:"reasons"`
	Facts         Facts             `json:"facts"`
	Sources       map[string]string `json:"sources"`
	Providers     []ProviderResult  `json:"providers"`
	CheckedAt     string            `json:"checkedAt"`
	PolicyVersion string            `json:"policyVersion"`
	ParserVersion string            `json:"parserVersion"`
}

// Check is an owner-scoped paid receipt referencing one shared report.
type Check struct {
	Operation model.OperationReceipt `json:"operation"`
	Report    *Report                `json:"report"`
	Cached    bool                   `json:"cached"`
	Charge    model.Money            `json:"charge"`
	UsedQuota bool                   `json:"usedQuota"`
	Refunded  bool                   `json:"refunded"`
}
