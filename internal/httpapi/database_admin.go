package httpapi

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/backup"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/databaseadmin"
)

// DatabaseAdministrationHTTP exposes the schema-aware editor and verified
// snapshot lifecycle. Mount it only inside the existing authenticated admin
// router so currentUser contains the trusted administrator.
type DatabaseAdministrationHTTP struct {
	editor     *databaseadmin.Service
	backups    *backup.Service
	migrations []string
}

// NewDatabaseAdministrationHTTP constructs independently mountable handlers.
func NewDatabaseAdministrationHTTP(editor *databaseadmin.Service, backups *backup.Service, migrations []string) *DatabaseAdministrationHTTP {
	return &DatabaseAdministrationHTTP{editor: editor, backups: backups, migrations: append([]string(nil), migrations...)}
}

// Mount registers database, backup download, and staged restore routes on an
// already-authenticated administrator router.
func (h *DatabaseAdministrationHTTP) Mount(router chi.Router) {
	router.Get("/database/tables", h.tables)
	router.Get("/database/tables/{table}/rows", h.rows)
	router.Post("/database/tables/{table}/query", h.queryRows)
	router.Put("/database/tables/{table}/rows", h.updateCompatibility)
	router.Post("/database/mutations/review", h.review)
	router.Post("/database/mutations", h.apply)
	router.Get("/backups/{id}/download", h.download)
	router.Post("/backups/{id}/restore", h.stageRestore)
	router.Get("/restores/{id}", h.restoreStatus)
}

func (h *DatabaseAdministrationHTTP) tables(w http.ResponseWriter, r *http.Request) {
	items, err := h.editor.Tables(r.Context())
	if err != nil {
		h.failure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *DatabaseAdministrationHTTP) rows(w http.ResponseWriter, r *http.Request) {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil && r.URL.Query().Get("limit") != "" {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_CURSOR", "The page limit must be an integer.")
		return
	}
	page, err := h.editor.Records(r.Context(), chi.URLParam(r, "table"), r.URL.Query().Get("cursor"), limit)
	if err != nil {
		h.failure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *DatabaseAdministrationHTTP) queryRows(w http.ResponseWriter, r *http.Request) {
	var request databaseadmin.QueryRequest
	if err := decodeJSON(w, r, &request); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_DATABASE_QUERY", "The typed database query is invalid.")
		return
	}
	page, err := h.editor.QueryRecords(r.Context(), chi.URLParam(r, "table"), request)
	if err != nil {
		h.failure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h *DatabaseAdministrationHTTP) review(w http.ResponseWriter, r *http.Request) {
	var request databaseadmin.MutationRequest
	if err := decodeJSON(w, r, &request); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_DATABASE_MUTATION", "The typed database mutation is invalid.")
		return
	}
	review, err := h.editor.ReviewMutation(r.Context(), currentUser(r).ID, request)
	if err != nil {
		h.failure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, review)
}

func (h *DatabaseAdministrationHTTP) apply(w http.ResponseWriter, r *http.Request) {
	var request databaseadmin.MutationRequest
	if err := decodeJSON(w, r, &request); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_DATABASE_MUTATION", "The reviewed database mutation is invalid.")
		return
	}
	result, err := h.editor.ApplyMutation(r.Context(), currentUser(r).ID, request)
	if err != nil {
		h.failure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// updateCompatibility preserves the initially published path but intentionally
// refuses old one-step writes because ApplyMutation requires a live reviewHash.
func (h *DatabaseAdministrationHTTP) updateCompatibility(w http.ResponseWriter, r *http.Request) {
	var request databaseadmin.MutationRequest
	if err := decodeJSON(w, r, &request); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_DATABASE_MUTATION", "The reviewed database mutation is invalid.")
		return
	}
	request.Action = "update"
	request.Table = chi.URLParam(r, "table")
	result, err := h.editor.ApplyMutation(r.Context(), currentUser(r).ID, request)
	if err != nil {
		h.failure(w, r, err)
		return
	}
	if result.Row == nil {
		h.writeError(w, r, http.StatusInternalServerError, "DATABASE_EDIT_FAILED", "The database record was not returned.")
		return
	}
	writeJSON(w, http.StatusOK, result.Row)
}

func (h *DatabaseAdministrationHTTP) download(w http.ResponseWriter, r *http.Request) {
	download, err := h.backups.OpenDownload(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.failure(w, r, err)
		return
	}
	defer func() { _ = download.File.Close() }()
	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": download.Name})
	w.Header().Set("Content-Type", "application/vnd.sqlite3")
	w.Header().Set("Content-Disposition", disposition)
	w.Header().Set("Content-Length", strconv.FormatInt(download.Size, 10))
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, download.File)
}

func (h *DatabaseAdministrationHTTP) stageRestore(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Reason       string `json:"reason"`
		Confirmation string `json:"confirmation"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_RESTORE", "A reason and typed confirmation are required.")
		return
	}
	job, err := h.backups.StageRestore(r.Context(), chi.URLParam(r, "id"), currentUser(r).ID, request.Reason, request.Confirmation, h.migrations)
	if err != nil {
		h.failure(w, r, err)
		return
	}
	writeJSON(w, http.StatusAccepted, mapRestoreOperation(job))
	h.backups.RequestRestart()
}

func (h *DatabaseAdministrationHTTP) restoreStatus(w http.ResponseWriter, r *http.Request) {
	job, err := h.backups.Restore(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.failure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, mapRestoreOperation(job))
}

type restoreOperationResponse struct {
	ID        string    `json:"id"`
	BackupID  string    `json:"backupId"`
	Status    string    `json:"status"`
	Error     *string   `json:"error,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func mapRestoreOperation(job backup.RestoreJob) restoreOperationResponse {
	status := "staging"
	switch job.Status {
	case "ready", "applying":
		status = "restarting"
	case "complete":
		status = "complete"
	case "failed", "rolled_back":
		status = "failed"
	}
	var message *string
	if status == "failed" {
		value := "Restore did not complete. The original database remains available."
		message = &value
	}
	return restoreOperationResponse{ID: job.ID, BackupID: job.BackupID, Status: status, Error: message, CreatedAt: job.CreatedAt, UpdatedAt: job.UpdatedAt}
}

func (h *DatabaseAdministrationHTTP) failure(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := http.StatusUnprocessableEntity, "DATABASE_ADMIN_FAILED", "The database administration operation could not be completed."
	switch {
	case errors.Is(err, databaseadmin.ErrTableNotFound), errors.Is(err, databaseadmin.ErrRecordNotFound), errors.Is(err, backup.ErrBackupNotFound):
		status, code, message = http.StatusNotFound, "NOT_FOUND", "The requested table, record, backup, or restore was not found."
	case errors.Is(err, databaseadmin.ErrOptimisticConflict), errors.Is(err, databaseadmin.ErrReviewConflict), errors.Is(err, backup.ErrRestoreConflict):
		status, code, message = http.StatusConflict, "DATABASE_CONFLICT", "The record or review changed. Refresh, review the diff again, and retry."
	case errors.Is(err, databaseadmin.ErrInvalidValue), errors.Is(err, databaseadmin.ErrConfirmation):
		status, code, message = http.StatusUnprocessableEntity, "INVALID_DATABASE_MUTATION", err.Error()
	case strings.Contains(strings.ToLower(err.Error()), "backup") || strings.Contains(strings.ToLower(err.Error()), "restore"):
		status, code, message = http.StatusInternalServerError, "BACKUP_OPERATION_FAILED", "The verified backup operation could not be completed."
	}
	h.writeError(w, r, status, code, message)
}

func (h *DatabaseAdministrationHTTP) writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, status, apiError{Code: code, Message: message, RequestID: middleware.GetReqID(r.Context())})
}
