// Package questionnaires implements form participation and CSV reward settlement.
package questionnaires

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

var ErrInvalidInput = errors.New("invalid questionnaire input")

// Status identifies the questionnaire lifecycle.

type Status string

const (
	// StatusDraft is editable and unavailable to members.
	StatusDraft Status = "draft"
	// StatusActive is the single questionnaire currently available to members.
	StatusActive Status = "active"
	// StatusClosed no longer accepts new participants.
	StatusClosed Status = "closed"
	// StatusSettling has a confirmed import queued or processing.
	StatusSettling Status = "settling"
	// StatusSettled has completed its one reward settlement.
	StatusSettled Status = "settled"
)

// QuestionnaireInput is an administrator-authored external form definition.

type QuestionnaireInput struct {
	ID             string     `json:"id,omitempty"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	FormURL        string     `json:"formUrl"`
	RewardTXBMinor int64      `json:"rewardTxbMinor"`
	Status         Status     `json:"status"`
	ClosesAt       *time.Time `json:"closesAt,omitempty"`
}

// Validate rejects invalid form hosts and lifecycle values.

func (input QuestionnaireInput) Validate() error {
	if title := strings.TrimSpace(input.Title); title == "" || len(title) > 120 {
		return fmt.Errorf("%w: title must be 1 to 120 bytes", ErrInvalidInput)
	}
	if len(strings.TrimSpace(input.Description)) > 4_000 {
		return fmt.Errorf("%w: description is too long", ErrInvalidInput)
	}
	if input.RewardTXBMinor < 0 {
		return fmt.Errorf("%w: reward cannot be negative", ErrInvalidInput)
	}
	if err := ValidateFormURL(input.FormURL); err != nil {
		return err
	}
	if input.Status != StatusDraft && input.Status != StatusActive && input.Status != StatusClosed {
		return fmt.Errorf("%w: status must be draft, active, or closed", ErrInvalidInput)
	}
	if input.ClosesAt != nil && input.ClosesAt.IsZero() {
		return fmt.Errorf("%w: close time cannot be zero", ErrInvalidInput)
	}
	return nil
}

// ValidateFormURL allows only HTTPS Google Forms and Microsoft Forms URLs.

func ValidateFormURL(rawURL string) error {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Host == "" {
		return fmt.Errorf("%w: form URL must be HTTPS without credentials", ErrInvalidInput)
	}
	host := strings.ToLower(parsed.Hostname())
	allowed := false
	switch host {
	case "forms.gle", "forms.office.com", "forms.microsoft.com", "forms.cloud.microsoft":
		allowed = true
	case "docs.google.com":
		allowed = strings.HasPrefix(parsed.EscapedPath(), "/forms/")
	}
	if !allowed {
		return fmt.Errorf("%w: form URL host is not an approved Google or Microsoft Forms host", ErrInvalidInput)
	}
	return nil
}

// Questionnaire is a persisted form definition.

type Questionnaire struct {
	QuestionnaireInput
	ParticipantCount int       `json:"participantCount"`
	RewardedCount    int       `json:"rewardedCount"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// Participant holds the durable validation code shown to a member.

type Participant struct {
	ID              string     `json:"id"`
	QuestionnaireID string     `json:"questionnaireId"`
	UserID          string     `json:"userId,omitempty"`
	ValidationCode  string     `json:"validationCode"`
	AwardedAt       *time.Time `json:"awardedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}

// ParticipationHistory combines a member code/result with its questionnaire.

type ParticipationHistory struct {
	Questionnaire Questionnaire `json:"questionnaire"`
	Participation Participant   `json:"participation"`
}

// CodeGenerator creates high-entropy validation codes.

type CodeGenerator interface {
	NewCode() (string, error)
}

// CryptoCodeGenerator creates 128-bit base32 validation codes.

type CryptoCodeGenerator struct{}

// NewCode returns a cryptographically random, human-copyable code.

func (CryptoCodeGenerator) NewCode() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("read questionnaire code entropy: %w", err)
	}
	return "TXQ-" + base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(value), nil
}

// NormalizeValidationCode trims and case-normalizes an imported validation code.

func NormalizeValidationCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

// ImportStatus identifies the CSV import lifecycle.

type ImportStatus string

const (
	// ImportStatusPreview has been parsed but not confirmed.
	ImportStatusPreview ImportStatus = "preview"
	// ImportStatusQueued is awaiting its durable settlement worker.
	ImportStatusQueued ImportStatus = "queued"
	// ImportStatusProcessing is currently leased by the settlement worker.
	ImportStatusProcessing ImportStatus = "processing"
	// ImportStatusSettled completed and will not credit again.
	ImportStatusSettled ImportStatus = "settled"
	// ImportStatusFailed needs an explicit administrator retry.
	ImportStatusFailed ImportStatus = "failed"
)

// CSVDocument is a bounded parsed upload. Raw is intentionally excluded from JSON.

type CSVDocument struct {
	Raw               []byte     `json:"-"`
	Delimiter         rune       `json:"-"`
	DelimiterName     string     `json:"delimiter"`
	Headers           []string   `json:"headers"`
	SampleRows        [][]string `json:"sampleRows"`
	DataRowCount      int        `json:"dataRowCount"`
	MalformedRowCount int        `json:"malformedRowCount"`
}

// ImportPreview is the persisted administrator confirmation surface.
