package questionnaires

import (
	"time"
)

type ImportPreview struct {
	ID                string          `json:"id"`
	QuestionnaireID   string          `json:"questionnaireId"`
	Status            ImportStatus    `json:"status"`
	Delimiter         string          `json:"delimiter"`
	Headers           []string        `json:"headers"`
	SampleRows        [][]string      `json:"sampleRows"`
	DataRowCount      int             `json:"dataRowCount"`
	MalformedRowCount int             `json:"malformedRowCount"`
	CodeColumn        *string         `json:"codeColumn,omitempty"`
	Analysis          *ImportAnalysis `json:"analysis,omitempty"`
	IdempotencyKey    string          `json:"idempotencyKey,omitempty"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

// ImportAnalysis is the non-mutating match preview shown before settlement.

type ImportAnalysis struct {
	ImportID            string `json:"importId"`
	QuestionnaireID     string `json:"questionnaireId"`
	CodeColumn          string `json:"codeColumn"`
	MatchedCount        int    `json:"matchedCount"`
	DuplicateCount      int    `json:"duplicateCount"`
	UnknownCount        int    `json:"unknownCount"`
	MalformedCount      int    `json:"malformedCount"`
	AlreadyAwardedCount int    `json:"alreadyAwardedCount"`
}

// SettlementReport records the exact result of a confirmed CSV import.

type SettlementReport struct {
	ImportID            string    `json:"importId"`
	QuestionnaireID     string    `json:"questionnaireId"`
	MatchedCount        int       `json:"matchedCount"`
	DuplicateCount      int       `json:"duplicateCount"`
	UnknownCount        int       `json:"unknownCount"`
	MalformedCount      int       `json:"malformedCount"`
	AlreadyAwardedCount int       `json:"alreadyAwardedCount"`
	RewardedCount       int       `json:"rewardedCount"`
	RewardTXBMinor      int64     `json:"rewardTxbMinor"`
	SettledAt           time.Time `json:"settledAt"`
	Replayed            bool      `json:"replayed"`
}

// ImportState combines preview/progress data with an optional final report.

type ImportState struct {
	Preview ImportPreview     `json:"preview"`
	Report  *SettlementReport `json:"report,omitempty"`
}
