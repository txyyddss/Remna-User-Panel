// Package questionnaires implements form participation and CSV reward settlement.
package questionnaires

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"
)

// ErrInvalidInput indicates an unsafe questionnaire request or import.
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
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
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

// Store is the narrow persistence contract required by Service.
type Store interface {
	SaveQuestionnaire(context.Context, QuestionnaireInput, time.Time) (Questionnaire, error)
	ListQuestionnaires(context.Context) ([]Questionnaire, error)
	ActiveQuestionnaire(context.Context, time.Time) (*Questionnaire, error)
	EnsureQuestionnaireParticipant(context.Context, string, string, CodeGenerator, time.Time) (Participant, error)
	ListQuestionnaireParticipations(context.Context, string, int) ([]ParticipationHistory, error)
	CreateQuestionnaireImport(context.Context, string, CSVDocument, string, time.Time) (ImportPreview, error)
	AnalyzeQuestionnaireImport(context.Context, string, string, time.Time) (ImportAnalysis, error)
	QueueQuestionnaireSettlement(context.Context, string, time.Time) (ImportPreview, error)
	QuestionnaireImportState(context.Context, string) (ImportState, error)
}

// Service validates form and CSV operations before persistence.
type Service struct {
	store      Store
	codes      CodeGenerator
	now        func() time.Time
	maxBytes   int64
	maxRows    int
	maxColumns int
}

// NewService constructs the questionnaire application service with plan limits.
func NewService(store Store, codes CodeGenerator, now func() time.Time) *Service {
	if codes == nil {
		codes = CryptoCodeGenerator{}
	}
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, codes: codes, now: now, maxBytes: 5 << 20, maxRows: 50_000, maxColumns: 100}
}

// Save validates and persists a draft, active, or closed questionnaire.
func (service *Service) Save(ctx context.Context, input QuestionnaireInput) (Questionnaire, error) {
	if err := input.Validate(); err != nil {
		return Questionnaire{}, err
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.FormURL = strings.TrimSpace(input.FormURL)
	now := service.now().UTC()
	if input.ClosesAt != nil {
		value := input.ClosesAt.UTC()
		input.ClosesAt = &value
		if input.Status == StatusActive && !value.After(now) {
			return Questionnaire{}, fmt.Errorf("%w: active questionnaire close time must be in the future", ErrInvalidInput)
		}
	}
	return service.store.SaveQuestionnaire(ctx, input, now)
}

// List returns questionnaire history.
func (service *Service) List(ctx context.Context) ([]Questionnaire, error) {
	return service.store.ListQuestionnaires(ctx)
}

// Active returns the single active questionnaire, if one exists.
func (service *Service) Active(ctx context.Context) (*Questionnaire, error) {
	return service.store.ActiveQuestionnaire(ctx, service.now().UTC())
}

// Participate returns the member's durable code, creating it exactly once.
func (service *Service) Participate(ctx context.Context, questionnaireID, userID string) (Participant, error) {
	if strings.TrimSpace(questionnaireID) == "" || strings.TrimSpace(userID) == "" {
		return Participant{}, fmt.Errorf("%w: questionnaire and user are required", ErrInvalidInput)
	}
	return service.store.EnsureQuestionnaireParticipant(ctx, questionnaireID, userID, service.codes, service.now().UTC())
}

// History returns the member's completed and in-progress questionnaire records.
func (service *Service) History(ctx context.Context, userID string, limit int) ([]ParticipationHistory, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("%w: user is required", ErrInvalidInput)
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return service.store.ListQuestionnaireParticipations(ctx, userID, limit)
}

// UploadCSV parses and stores a bounded preview without applying rewards.
func (service *Service) UploadCSV(ctx context.Context, questionnaireID string, reader io.Reader, idempotencyKey string) (ImportPreview, error) {
	if strings.TrimSpace(questionnaireID) == "" || !validIdempotencyKey(idempotencyKey) {
		return ImportPreview{}, fmt.Errorf("%w: questionnaire and idempotency key are required", ErrInvalidInput)
	}
	document, err := ParseCSV(reader, service.maxBytes, service.maxRows, service.maxColumns)
	if err != nil {
		return ImportPreview{}, err
	}
	return service.store.CreateQuestionnaireImport(ctx, questionnaireID, document, idempotencyKey, service.now().UTC())
}

// AnalyzeImport selects the validation-code column and reports matches without mutation.
func (service *Service) AnalyzeImport(ctx context.Context, importID, codeColumn string) (ImportAnalysis, error) {
	if strings.TrimSpace(importID) == "" || strings.TrimSpace(codeColumn) == "" {
		return ImportAnalysis{}, fmt.Errorf("%w: import and code column are required", ErrInvalidInput)
	}
	return service.store.AnalyzeQuestionnaireImport(ctx, importID, strings.TrimSpace(codeColumn), service.now().UTC())
}

// ConfirmImport queues a previously analyzed import for durable settlement.
func (service *Service) ConfirmImport(ctx context.Context, importID string) (ImportPreview, error) {
	if strings.TrimSpace(importID) == "" {
		return ImportPreview{}, fmt.Errorf("%w: import is required", ErrInvalidInput)
	}
	return service.store.QueueQuestionnaireSettlement(ctx, importID, service.now().UTC())
}

// Import returns current preview, processing, or settlement state.
func (service *Service) Import(ctx context.Context, importID string) (ImportState, error) {
	if strings.TrimSpace(importID) == "" {
		return ImportState{}, fmt.Errorf("%w: import is required", ErrInvalidInput)
	}
	return service.store.QuestionnaireImportState(ctx, importID)
}

func validIdempotencyKey(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && len(value) <= 128
}
