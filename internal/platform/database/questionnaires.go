package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
	"github.com/txyyddss/Remna-User-Panel/internal/questionnaires"
)

const questionnaireSelect = `SELECT id,title,description,form_url,reward_txb_minor,status,closes_at,created_at,updated_at FROM questionnaires`

// SaveQuestionnaire creates or updates a draft, active, or closed questionnaire.
func (s *Store) SaveQuestionnaire(ctx context.Context, input questionnaires.QuestionnaireInput, now time.Time) (questionnaires.Questionnaire, error) {
	if err := input.Validate(); err != nil {
		return questionnaires.Questionnaire{}, err
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	input.FormURL = strings.TrimSpace(input.FormURL)
	now = now.UTC()
	if input.ClosesAt != nil {
		value := input.ClosesAt.UTC()
		input.ClosesAt = &value
		if input.Status == questionnaires.StatusActive && !value.After(now) {
			return questionnaires.Questionnaire{}, questionnaires.ErrInvalidInput
		}
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return questionnaires.Questionnaire{}, fmt.Errorf("begin questionnaire save: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if input.Status == questionnaires.StatusActive {
		var activeID string
		var activeClosesAt sql.NullString
		loadErr := tx.QueryRowContext(ctx, `SELECT id,closes_at FROM questionnaires WHERE status='active'`).Scan(&activeID, &activeClosesAt)
		if loadErr != nil && !errors.Is(loadErr, sql.ErrNoRows) {
			return questionnaires.Questionnaire{}, fmt.Errorf("load active questionnaire: %w", loadErr)
		}
		if loadErr == nil && activeClosesAt.Valid {
			closes, parseErr := parseStamp(activeClosesAt.String)
			if parseErr != nil {
				return questionnaires.Questionnaire{}, parseErr
			}
			if !closes.After(now) {
				if _, err := tx.ExecContext(ctx, `UPDATE questionnaires SET status='closed',updated_at=? WHERE id=? AND status='active'`, stamp(now), activeID); err != nil {
					return questionnaires.Questionnaire{}, fmt.Errorf("close expired questionnaire: %w", err)
				}
			}
		}
	}
	if input.ID == "" {
		input.ID, err = ids.New()
		if err != nil {
			return questionnaires.Questionnaire{}, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO questionnaires(id,title,description,form_url,reward_txb_minor,status,closes_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`,
			input.ID, input.Title, input.Description, input.FormURL, input.RewardTXBMinor, input.Status, nullableStamp(input.ClosesAt), stamp(now), stamp(now))
	} else {
		var result sql.Result
		result, err = tx.ExecContext(ctx, `UPDATE questionnaires SET title=?,description=?,form_url=?,reward_txb_minor=?,status=?,closes_at=?,updated_at=?
			WHERE id=? AND status IN ('draft','active','closed')`, input.Title, input.Description, input.FormURL, input.RewardTXBMinor, input.Status, nullableStamp(input.ClosesAt), stamp(now), input.ID)
		if err == nil {
			if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
				return questionnaires.Questionnaire{}, rowsErr
			} else if affected == 0 {
				return questionnaires.Questionnaire{}, ErrConflict
			}
		}
	}
	if err != nil {
		if isUniqueConstraint(err) {
			return questionnaires.Questionnaire{}, ErrConflict
		}
		return questionnaires.Questionnaire{}, fmt.Errorf("save questionnaire: %w", err)
	}
	questionnaire, err := questionnaireByID(ctx, tx, input.ID)
	if err != nil {
		return questionnaires.Questionnaire{}, err
	}
	if err := tx.Commit(); err != nil {
		return questionnaires.Questionnaire{}, fmt.Errorf("commit questionnaire save: %w", err)
	}
	return questionnaire, nil
}

// ListQuestionnaires returns newest questionnaire history first.
func (s *Store) ListQuestionnaires(ctx context.Context) ([]questionnaires.Questionnaire, error) {
	rows, err := s.db.QueryContext(ctx, questionnaireSelect+` ORDER BY created_at DESC,id DESC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make([]questionnaires.Questionnaire, 0)
	for rows.Next() {
		questionnaire, scanErr := scanQuestionnaire(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, questionnaire)
	}
	return result, rows.Err()
}

// ActiveQuestionnaire returns the single active questionnaire, if present.
func (s *Store) ActiveQuestionnaire(ctx context.Context, now time.Time) (*questionnaires.Questionnaire, error) {
	questionnaire, err := scanQuestionnaire(s.db.QueryRowContext(ctx, questionnaireSelect+` WHERE status='active'`))
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if questionnaire.ClosesAt != nil && !questionnaire.ClosesAt.After(now.UTC()) {
		return nil, nil
	}
	return &questionnaire, nil
}

// EnsureQuestionnaireParticipant returns a durable code, generating it once.
func (s *Store) EnsureQuestionnaireParticipant(ctx context.Context, questionnaireID, userID string, generator questionnaires.CodeGenerator, now time.Time) (questionnaires.Participant, error) {
	if strings.TrimSpace(questionnaireID) == "" || strings.TrimSpace(userID) == "" || generator == nil {
		return questionnaires.Participant{}, questionnaires.ErrInvalidInput
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return questionnaires.Participant{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if existing, loadErr := participantByUserTx(ctx, tx, questionnaireID, userID); loadErr == nil {
		return existing, nil
	} else if !errors.Is(loadErr, ErrNotFound) {
		return questionnaires.Participant{}, loadErr
	}
	var status questionnaires.Status
	var closesAt sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT status,closes_at FROM questionnaires WHERE id=?`, questionnaireID).Scan(&status, &closesAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return questionnaires.Participant{}, ErrNotFound
		}
		return questionnaires.Participant{}, err
	}
	if status != questionnaires.StatusActive {
		return questionnaires.Participant{}, ErrConflict
	}
	now = now.UTC()
	if closesAt.Valid {
		closes, parseErr := parseStamp(closesAt.String)
		if parseErr != nil {
			return questionnaires.Participant{}, parseErr
		}
		if !closes.After(now) {
			return questionnaires.Participant{}, ErrConflict
		}
	}
	participantID, err := ids.New()
	if err != nil {
		return questionnaires.Participant{}, err
	}
	for attempt := 0; attempt < 5; attempt++ {
		code, codeErr := generator.NewCode()
		if codeErr != nil {
			return questionnaires.Participant{}, codeErr
		}
		code = questionnaires.NormalizeValidationCode(code)
		if code == "" || len(code) > 128 {
			return questionnaires.Participant{}, questionnaires.ErrInvalidInput
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO questionnaire_participants(id,questionnaire_id,user_id,validation_code,created_at) VALUES(?,?,?,?,?)`,
			participantID, questionnaireID, userID, code, stamp(now))
		if err == nil {
			if err := tx.Commit(); err != nil {
				return questionnaires.Participant{}, err
			}
			return s.questionnaireParticipantByID(ctx, participantID)
		}
		if !isUniqueConstraint(err) {
			return questionnaires.Participant{}, fmt.Errorf("create questionnaire participant: %w", err)
		}
	}
	return questionnaires.Participant{}, fmt.Errorf("generate unique questionnaire code: %w", ErrConflict)
}

// ListQuestionnaireParticipations returns a member's durable questionnaire history.
func (s *Store) ListQuestionnaireParticipations(ctx context.Context, userID string, limit int) ([]questionnaires.ParticipationHistory, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `SELECT
		q.id,q.title,q.description,q.form_url,q.reward_txb_minor,q.status,q.closes_at,q.created_at,q.updated_at,
		p.id,p.questionnaire_id,p.user_id,p.validation_code,p.awarded_at,p.created_at
		FROM questionnaire_participants p JOIN questionnaires q ON q.id=p.questionnaire_id
		WHERE p.user_id=? ORDER BY p.created_at DESC,p.id DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list questionnaire participation history: %w", err)
	}
	defer func() { _ = rows.Close() }()
	result := make([]questionnaires.ParticipationHistory, 0)
	for rows.Next() {
		var item questionnaires.ParticipationHistory
		var questionnaireCreated, questionnaireUpdated, participantCreated string
		var closesAt, awarded sql.NullString
		if err := rows.Scan(&item.Questionnaire.ID, &item.Questionnaire.Title, &item.Questionnaire.Description, &item.Questionnaire.FormURL,
			&item.Questionnaire.RewardTXBMinor, &item.Questionnaire.Status, &closesAt, &questionnaireCreated, &questionnaireUpdated,
			&item.Participation.ID, &item.Participation.QuestionnaireID, &item.Participation.UserID, &item.Participation.ValidationCode,
			&awarded, &participantCreated); err != nil {
			return nil, err
		}
		if item.Questionnaire.CreatedAt, err = parseStamp(questionnaireCreated); err != nil {
			return nil, err
		}
		if item.Questionnaire.UpdatedAt, err = parseStamp(questionnaireUpdated); err != nil {
			return nil, err
		}
		if closesAt.Valid {
			value, parseErr := parseStamp(closesAt.String)
			if parseErr != nil {
				return nil, parseErr
			}
			item.Questionnaire.ClosesAt = &value
		}
		if item.Participation.CreatedAt, err = parseStamp(participantCreated); err != nil {
			return nil, err
		}
		if awarded.Valid {
			value, parseErr := parseStamp(awarded.String)
			if parseErr != nil {
				return nil, parseErr
			}
			item.Participation.AwardedAt = &value
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// CreateQuestionnaireImport persists a bounded preview without applying rewards.
func (s *Store) CreateQuestionnaireImport(ctx context.Context, questionnaireID string, document questionnaires.CSVDocument, idempotencyKey string, now time.Time) (questionnaires.ImportPreview, error) {
	if strings.TrimSpace(questionnaireID) == "" || strings.TrimSpace(idempotencyKey) == "" || len(idempotencyKey) > 128 || len(document.Raw) == 0 || len(document.Raw) > 5<<20 || len(document.Headers) == 0 || len(document.Headers) > 100 || document.DataRowCount > 50_000 {
		return questionnaires.ImportPreview{}, questionnaires.ErrInvalidInput
	}
	if document.Delimiter != ',' && document.Delimiter != ';' && document.Delimiter != '\t' {
		return questionnaires.ImportPreview{}, questionnaires.ErrInvalidInput
	}
	headersJSON, err := json.Marshal(document.Headers)
	if err != nil {
		return questionnaires.ImportPreview{}, err
	}
	sampleJSON, err := json.Marshal(document.SampleRows)
	if err != nil {
		return questionnaires.ImportPreview{}, err
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return questionnaires.ImportPreview{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if existing, loadErr := importByKeyTx(ctx, tx, questionnaireID, idempotencyKey); loadErr == nil {
		return existing, nil
	} else if !errors.Is(loadErr, ErrNotFound) {
		return questionnaires.ImportPreview{}, loadErr
	}
	var status questionnaires.Status
	if err := tx.QueryRowContext(ctx, `SELECT status FROM questionnaires WHERE id=?`, questionnaireID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return questionnaires.ImportPreview{}, ErrNotFound
		}
		return questionnaires.ImportPreview{}, err
	}
	if status != questionnaires.StatusActive && status != questionnaires.StatusClosed {
		return questionnaires.ImportPreview{}, ErrConflict
	}
	importID, err := ids.New()
	if err != nil {
		return questionnaires.ImportPreview{}, err
	}
	now = now.UTC()
	_, err = tx.ExecContext(ctx, `INSERT INTO questionnaire_imports(id,questionnaire_id,status,raw_csv,delimiter,headers_json,sample_rows_json,data_row_count,malformed_row_count,idempotency_key,created_at,updated_at)
		VALUES(?,?,'preview',?,?,?,?,?,?,?,?,?)`, importID, questionnaireID, document.Raw, string(document.Delimiter), string(headersJSON), string(sampleJSON),
		document.DataRowCount, document.MalformedRowCount, idempotencyKey, stamp(now), stamp(now))
	if err != nil {
		return questionnaires.ImportPreview{}, fmt.Errorf("create questionnaire import: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return questionnaires.ImportPreview{}, err
	}
	return s.questionnaireImportByID(ctx, importID)
}

// AnalyzeQuestionnaireImport matches a selected code column without applying rewards.
func (s *Store) AnalyzeQuestionnaireImport(ctx context.Context, importID, codeColumn string, now time.Time) (questionnaires.ImportAnalysis, error) {
	if strings.TrimSpace(importID) == "" || strings.TrimSpace(codeColumn) == "" {
		return questionnaires.ImportAnalysis{}, questionnaires.ErrInvalidInput
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return questionnaires.ImportAnalysis{}, err
	}
	defer func() { _ = tx.Rollback() }()
	preview, raw, delimiter, _, err := importSettlementDataTx(ctx, tx, importID)
	if err != nil {
		return questionnaires.ImportAnalysis{}, err
	}
	if preview.Status != questionnaires.ImportStatusPreview {
		return questionnaires.ImportAnalysis{}, ErrConflict
	}
	matches := 0
	for _, header := range preview.Headers {
		if header == codeColumn {
			matches++
		}
	}
	if matches != 1 {
		return questionnaires.ImportAnalysis{}, questionnaires.ErrInvalidInput
	}
	parsed, err := questionnaires.ExtractCodes(raw, delimiter, preview.Headers, codeColumn, 50_000)
	if err != nil {
		return questionnaires.ImportAnalysis{}, err
	}
	analysis := questionnaires.ImportAnalysis{ImportID: importID, QuestionnaireID: preview.QuestionnaireID, CodeColumn: codeColumn,
		DuplicateCount: parsed.DuplicateCount, MalformedCount: parsed.MalformedCount}
	for _, code := range parsed.Codes {
		participant, loadErr := participantByCodeTx(ctx, tx, preview.QuestionnaireID, code)
		if errors.Is(loadErr, ErrNotFound) {
			analysis.UnknownCount++
			continue
		}
		if loadErr != nil {
			return questionnaires.ImportAnalysis{}, loadErr
		}
		analysis.MatchedCount++
		if participant.AwardedAt != nil {
			analysis.AlreadyAwardedCount++
		}
	}
	analysisJSON, err := json.Marshal(analysis)
	if err != nil {
		return questionnaires.ImportAnalysis{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE questionnaire_imports SET code_column=?,analysis_json=?,updated_at=? WHERE id=? AND status='preview'`, codeColumn, string(analysisJSON), stamp(now.UTC()), importID); err != nil {
		return questionnaires.ImportAnalysis{}, err
	}
	if err := tx.Commit(); err != nil {
		return questionnaires.ImportAnalysis{}, err
	}
	return analysis, nil
}

// QueueQuestionnaireSettlement confirms an analyzed import and appends durable work.
func (s *Store) QueueQuestionnaireSettlement(ctx context.Context, importID string, now time.Time) (questionnaires.ImportPreview, error) {
	if strings.TrimSpace(importID) == "" {
		return questionnaires.ImportPreview{}, questionnaires.ErrInvalidInput
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return questionnaires.ImportPreview{}, err
	}
	defer func() { _ = tx.Rollback() }()
	preview, err := importByIDTx(ctx, tx, importID)
	if err != nil {
		return questionnaires.ImportPreview{}, err
	}
	if preview.Status == questionnaires.ImportStatusQueued || preview.Status == questionnaires.ImportStatusProcessing || preview.Status == questionnaires.ImportStatusSettled {
		return preview, nil
	}
	if preview.Status != questionnaires.ImportStatusPreview {
		return questionnaires.ImportPreview{}, ErrConflict
	}
	if preview.CodeColumn == nil || preview.Analysis == nil || preview.Analysis.CodeColumn != *preview.CodeColumn {
		return questionnaires.ImportPreview{}, ErrConflict
	}
	var competing int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM questionnaire_imports WHERE questionnaire_id=? AND id<>? AND status IN ('queued','processing','settled')`, preview.QuestionnaireID, importID).Scan(&competing); err != nil {
		return questionnaires.ImportPreview{}, err
	}
	if competing != 0 {
		return questionnaires.ImportPreview{}, ErrConflict
	}
	now = now.UTC()
	if _, err := tx.ExecContext(ctx, `UPDATE questionnaire_imports SET status='queued',updated_at=? WHERE id=? AND status='preview'`, stamp(now), importID); err != nil {
		return questionnaires.ImportPreview{}, err
	}
	result, err := tx.ExecContext(ctx, `UPDATE questionnaires SET status='settling',updated_at=? WHERE id=? AND status IN ('active','closed')`, stamp(now), preview.QuestionnaireID)
	if err != nil {
		return questionnaires.ImportPreview{}, err
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return questionnaires.ImportPreview{}, rowsErr
	} else if affected != 1 {
		return questionnaires.ImportPreview{}, ErrConflict
	}
	payload, err := json.Marshal(struct {
		ImportID string `json:"importId"`
	}{ImportID: importID})
	if err != nil {
		return questionnaires.ImportPreview{}, err
	}
	if err := insertOutboxTx(ctx, tx, "questionnaire_settlement", string(payload), now, now); err != nil {
		return questionnaires.ImportPreview{}, err
	}
	if err := tx.Commit(); err != nil {
		return questionnaires.ImportPreview{}, err
	}
	return s.questionnaireImportByID(ctx, importID)
}

// SettleQuestionnaireImport atomically matches codes and credits each participant once.
func (s *Store) SettleQuestionnaireImport(ctx context.Context, importID string, now time.Time) (questionnaires.SettlementReport, error) {
	if strings.TrimSpace(importID) == "" {
		return questionnaires.SettlementReport{}, questionnaires.ErrInvalidInput
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	now = now.UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return questionnaires.SettlementReport{}, err
	}
	defer func() { _ = tx.Rollback() }()
	preview, raw, delimiter, reportJSON, err := importSettlementDataTx(ctx, tx, importID)
	if err != nil {
		return questionnaires.SettlementReport{}, err
	}
	if preview.Status == questionnaires.ImportStatusSettled {
		var report questionnaires.SettlementReport
		if err := json.Unmarshal([]byte(reportJSON), &report); err != nil {
			return questionnaires.SettlementReport{}, fmt.Errorf("decode questionnaire settlement report: %w", err)
		}
		report.Replayed = true
		return report, nil
	}
	if preview.Status != questionnaires.ImportStatusQueued && preview.Status != questionnaires.ImportStatusProcessing {
		return questionnaires.SettlementReport{}, ErrConflict
	}
	if preview.CodeColumn == nil {
		return questionnaires.SettlementReport{}, ErrConflict
	}
	if _, err := tx.ExecContext(ctx, `UPDATE questionnaire_imports SET status='processing',updated_at=? WHERE id=?`, stamp(now), importID); err != nil {
		return questionnaires.SettlementReport{}, err
	}
	parsed, err := questionnaires.ExtractCodes(raw, delimiter, preview.Headers, *preview.CodeColumn, 50_000)
	if err != nil {
		return questionnaires.SettlementReport{}, err
	}
	var rewardMinor int64
	if err := tx.QueryRowContext(ctx, `SELECT reward_txb_minor FROM questionnaires WHERE id=? AND status='settling'`, preview.QuestionnaireID).Scan(&rewardMinor); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return questionnaires.SettlementReport{}, ErrConflict
		}
		return questionnaires.SettlementReport{}, err
	}
	report := questionnaires.SettlementReport{ImportID: importID, QuestionnaireID: preview.QuestionnaireID,
		DuplicateCount: parsed.DuplicateCount, MalformedCount: parsed.MalformedCount,
		RewardTXBMinor: rewardMinor, SettledAt: now}
	for _, code := range parsed.Codes {
		participant, loadErr := participantByCodeTx(ctx, tx, preview.QuestionnaireID, code)
		if errors.Is(loadErr, ErrNotFound) {
			report.UnknownCount++
			continue
		}
		if loadErr != nil {
			return questionnaires.SettlementReport{}, loadErr
		}
		report.MatchedCount++
		if participant.AwardedAt != nil {
			report.AlreadyAwardedCount++
			continue
		}
		result, updateErr := tx.ExecContext(ctx, `UPDATE questionnaire_participants SET awarded_at=? WHERE id=? AND awarded_at IS NULL`, stamp(now), participant.ID)
		if updateErr != nil {
			return questionnaires.SettlementReport{}, updateErr
		}
		if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
			return questionnaires.SettlementReport{}, rowsErr
		} else if affected != 1 {
			report.AlreadyAwardedCount++
			continue
		}
		balance, balanceErr := changeBalanceTx(ctx, tx, participant.UserID, rewardMinor, now)
		if balanceErr != nil {
			return questionnaires.SettlementReport{}, balanceErr
		}
		if _, ledgerErr := insertLedgerTx(ctx, tx, participant.UserID, rewardMinor, balance, "questionnaire_reward", participant.ID, preview.QuestionnaireID, now); ledgerErr != nil {
			return questionnaires.SettlementReport{}, ledgerErr
		}
		report.RewardedCount++
	}
	reportBytes, err := json.Marshal(report)
	if err != nil {
		return questionnaires.SettlementReport{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE questionnaire_imports SET status='settled',raw_csv=x'',report_json=?,last_error='',updated_at=? WHERE id=?`, string(reportBytes), stamp(now), importID); err != nil {
		return questionnaires.SettlementReport{}, err
	}
	result, err := tx.ExecContext(ctx, `UPDATE questionnaires SET status='settled',updated_at=? WHERE id=? AND status='settling'`, stamp(now), preview.QuestionnaireID)
	if err != nil {
		return questionnaires.SettlementReport{}, err
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return questionnaires.SettlementReport{}, rowsErr
	} else if affected != 1 {
		return questionnaires.SettlementReport{}, ErrConflict
	}
	if err := tx.Commit(); err != nil {
		return questionnaires.SettlementReport{}, err
	}
	return report, nil
}

// QuestionnaireImportState returns pollable preview/progress/final state.
func (s *Store) QuestionnaireImportState(ctx context.Context, importID string) (questionnaires.ImportState, error) {
	preview, err := s.questionnaireImportByID(ctx, importID)
	if err != nil {
		return questionnaires.ImportState{}, err
	}
	state := questionnaires.ImportState{Preview: preview}
	if preview.Status != questionnaires.ImportStatusSettled {
		return state, nil
	}
	var reportJSON string
	if err := s.db.QueryRowContext(ctx, `SELECT report_json FROM questionnaire_imports WHERE id=?`, importID).Scan(&reportJSON); err != nil {
		return questionnaires.ImportState{}, err
	}
	var report questionnaires.SettlementReport
	if err := json.Unmarshal([]byte(reportJSON), &report); err != nil {
		return questionnaires.ImportState{}, fmt.Errorf("decode questionnaire settlement report: %w", err)
	}
	state.Report = &report
	return state, nil
}

func questionnaireByID(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, questionnaireID string) (questionnaires.Questionnaire, error) {
	return scanQuestionnaire(queryer.QueryRowContext(ctx, questionnaireSelect+` WHERE id=?`, questionnaireID))
}

func scanQuestionnaire(row rowScanner) (questionnaires.Questionnaire, error) {
	var questionnaire questionnaires.Questionnaire
	var closesAt sql.NullString
	var created, updated string
	if err := row.Scan(&questionnaire.ID, &questionnaire.Title, &questionnaire.Description, &questionnaire.FormURL, &questionnaire.RewardTXBMinor, &questionnaire.Status, &closesAt, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return questionnaires.Questionnaire{}, ErrNotFound
		}
		return questionnaires.Questionnaire{}, err
	}
	var err error
	if questionnaire.CreatedAt, err = parseStamp(created); err != nil {
		return questionnaires.Questionnaire{}, err
	}
	if closesAt.Valid {
		value, parseErr := parseStamp(closesAt.String)
		if parseErr != nil {
			return questionnaires.Questionnaire{}, parseErr
		}
		questionnaire.ClosesAt = &value
	}
	questionnaire.UpdatedAt, err = parseStamp(updated)
	return questionnaire, err
}

const participantSelect = `SELECT id,questionnaire_id,user_id,validation_code,awarded_at,created_at FROM questionnaire_participants`

func (s *Store) questionnaireParticipantByID(ctx context.Context, participantID string) (questionnaires.Participant, error) {
	return scanParticipant(s.db.QueryRowContext(ctx, participantSelect+` WHERE id=?`, participantID))
}

func participantByUserTx(ctx context.Context, tx *sql.Tx, questionnaireID, userID string) (questionnaires.Participant, error) {
	return scanParticipant(tx.QueryRowContext(ctx, participantSelect+` WHERE questionnaire_id=? AND user_id=?`, questionnaireID, userID))
}

func participantByCodeTx(ctx context.Context, tx *sql.Tx, questionnaireID, code string) (questionnaires.Participant, error) {
	return scanParticipant(tx.QueryRowContext(ctx, participantSelect+` WHERE questionnaire_id=? AND validation_code=?`, questionnaireID, questionnaires.NormalizeValidationCode(code)))
}

func scanParticipant(row rowScanner) (questionnaires.Participant, error) {
	var participant questionnaires.Participant
	var awarded sql.NullString
	var created string
	if err := row.Scan(&participant.ID, &participant.QuestionnaireID, &participant.UserID, &participant.ValidationCode, &awarded, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return questionnaires.Participant{}, ErrNotFound
		}
		return questionnaires.Participant{}, err
	}
	var err error
	if participant.CreatedAt, err = parseStamp(created); err != nil {
		return questionnaires.Participant{}, err
	}
	if awarded.Valid {
		value, parseErr := parseStamp(awarded.String)
		if parseErr != nil {
			return questionnaires.Participant{}, parseErr
		}
		participant.AwardedAt = &value
	}
	return participant, nil
}

const questionnaireImportSelect = `SELECT id,questionnaire_id,status,delimiter,headers_json,sample_rows_json,data_row_count,malformed_row_count,code_column,analysis_json,idempotency_key,created_at,updated_at FROM questionnaire_imports`

func (s *Store) questionnaireImportByID(ctx context.Context, importID string) (questionnaires.ImportPreview, error) {
	return scanQuestionnaireImport(s.db.QueryRowContext(ctx, questionnaireImportSelect+` WHERE id=?`, importID))
}

func importByIDTx(ctx context.Context, tx *sql.Tx, importID string) (questionnaires.ImportPreview, error) {
	return scanQuestionnaireImport(tx.QueryRowContext(ctx, questionnaireImportSelect+` WHERE id=?`, importID))
}

func importByKeyTx(ctx context.Context, tx *sql.Tx, questionnaireID, key string) (questionnaires.ImportPreview, error) {
	return scanQuestionnaireImport(tx.QueryRowContext(ctx, questionnaireImportSelect+` WHERE questionnaire_id=? AND idempotency_key=?`, questionnaireID, key))
}

func scanQuestionnaireImport(row rowScanner) (questionnaires.ImportPreview, error) {
	var preview questionnaires.ImportPreview
	var delimiter, headersJSON, samplesJSON string
	var codeColumn, analysisJSON sql.NullString
	var created, updated string
	if err := row.Scan(&preview.ID, &preview.QuestionnaireID, &preview.Status, &delimiter, &headersJSON, &samplesJSON, &preview.DataRowCount,
		&preview.MalformedRowCount, &codeColumn, &analysisJSON, &preview.IdempotencyKey, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return questionnaires.ImportPreview{}, ErrNotFound
		}
		return questionnaires.ImportPreview{}, err
	}
	switch delimiter {
	case ";":
		preview.Delimiter = "semicolon"
	case "\t":
		preview.Delimiter = "tab"
	default:
		preview.Delimiter = "comma"
	}
	if err := json.Unmarshal([]byte(headersJSON), &preview.Headers); err != nil {
		return questionnaires.ImportPreview{}, err
	}
	if err := json.Unmarshal([]byte(samplesJSON), &preview.SampleRows); err != nil {
		return questionnaires.ImportPreview{}, err
	}
	if codeColumn.Valid {
		preview.CodeColumn = &codeColumn.String
	}
	if analysisJSON.Valid {
		var analysis questionnaires.ImportAnalysis
		if err := json.Unmarshal([]byte(analysisJSON.String), &analysis); err != nil {
			return questionnaires.ImportPreview{}, err
		}
		preview.Analysis = &analysis
	}
	var err error
	if preview.CreatedAt, err = parseStamp(created); err != nil {
		return questionnaires.ImportPreview{}, err
	}
	preview.UpdatedAt, err = parseStamp(updated)
	return preview, err
}

func importSettlementDataTx(ctx context.Context, tx *sql.Tx, importID string) (questionnaires.ImportPreview, []byte, rune, string, error) {
	var preview questionnaires.ImportPreview
	var raw []byte
	var delimiter, headersJSON, samplesJSON string
	var codeColumn, analysisJSON, reportJSON sql.NullString
	var created, updated string
	err := tx.QueryRowContext(ctx, `SELECT id,questionnaire_id,status,raw_csv,delimiter,headers_json,sample_rows_json,data_row_count,malformed_row_count,code_column,analysis_json,idempotency_key,report_json,created_at,updated_at
		FROM questionnaire_imports WHERE id=?`, importID).Scan(&preview.ID, &preview.QuestionnaireID, &preview.Status, &raw, &delimiter, &headersJSON, &samplesJSON,
		&preview.DataRowCount, &preview.MalformedRowCount, &codeColumn, &analysisJSON, &preview.IdempotencyKey, &reportJSON, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return questionnaires.ImportPreview{}, nil, 0, "", ErrNotFound
	}
	if err != nil {
		return questionnaires.ImportPreview{}, nil, 0, "", err
	}
	if err := json.Unmarshal([]byte(headersJSON), &preview.Headers); err != nil {
		return questionnaires.ImportPreview{}, nil, 0, "", err
	}
	if err := json.Unmarshal([]byte(samplesJSON), &preview.SampleRows); err != nil {
		return questionnaires.ImportPreview{}, nil, 0, "", err
	}
	if codeColumn.Valid {
		preview.CodeColumn = &codeColumn.String
	}
	if analysisJSON.Valid {
		var analysis questionnaires.ImportAnalysis
		if err := json.Unmarshal([]byte(analysisJSON.String), &analysis); err != nil {
			return questionnaires.ImportPreview{}, nil, 0, "", err
		}
		preview.Analysis = &analysis
	}
	preview.CreatedAt, err = parseStamp(created)
	if err != nil {
		return questionnaires.ImportPreview{}, nil, 0, "", err
	}
	preview.UpdatedAt, err = parseStamp(updated)
	if err != nil {
		return questionnaires.ImportPreview{}, nil, 0, "", err
	}
	runes := []rune(delimiter)
	if len(runes) != 1 {
		return questionnaires.ImportPreview{}, nil, 0, "", questionnaires.ErrInvalidInput
	}
	return preview, raw, runes[0], reportJSON.String, nil
}
