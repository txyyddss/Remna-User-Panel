package questionnaires

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

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
