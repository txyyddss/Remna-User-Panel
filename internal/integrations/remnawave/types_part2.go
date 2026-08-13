package remnawave

import "time"

type InternalSquad struct {
	UUID         string    `json:"uuid"`
	ViewPosition int64     `json:"viewPosition"`
	Name         string    `json:"name"`
	Info         SquadInfo `json:"info"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// NodeInbound is one currently active inbound exposed by a Remnawave node.

type NodeInbound struct {
	UUID string `json:"uuid"`
}

// Node is the documented subset required to map node selections to squad
// inbounds. Links and other volatile upstream properties are intentionally not
// represented or persisted.

type Node struct {
	UUID                  string  `json:"uuid"`
	Name                  string  `json:"name"`
	CountryCode           string  `json:"countryCode"`
	ConsumptionMultiplier float64 `json:"consumptionMultiplier"`
	IsDisabled            bool    `json:"isDisabled"`
	ConfigProfile         struct {
		ActiveInbounds []NodeInbound `json:"activeInbounds"`
	} `json:"configProfile"`
}

// AccessibleNode is returned after Remnawave resolves squad inbounds to nodes.

type AccessibleNode struct {
	UUID           string   `json:"uuid"`
	NodeName       string   `json:"nodeName"`
	CountryCode    string   `json:"countryCode"`
	ActiveInbounds []string `json:"activeInbounds"`
}

// ValidationIssue describes one Remnawave request validation failure.

type ValidationIssue struct {
	Validation string   `json:"validation"`
	Code       string   `json:"code"`
	Message    string   `json:"message"`
	Path       []string `json:"path"`
}
