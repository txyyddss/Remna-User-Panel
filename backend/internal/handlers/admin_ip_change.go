package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/user/remna-user-panel/internal/middleware"
)

func (h *Handler) AdminListIPChangeRequests(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	requests, total, err := h.IPChanges.ListAdminRequests(r.Context(), limit, offset)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"requests": requests,
		"total":    total,
	})
}

func (h *Handler) AdminIPChangeAction(w http.ResponseWriter, r *http.Request) {
	requestID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request id")
		return
	}

	action := chi.URLParam(r, "action")

	switch action {
	case "approve":
		req, err := h.IPChanges.AdminApproveRequest(r.Context(), requestID)
		if err != nil {
			writeIPChangeError(w, err)
			return
		}
		middleware.WriteSuccess(w, req)
	case "decline":
		req, err := h.IPChanges.AdminDeclineRequest(r.Context(), requestID)
		if err != nil {
			writeIPChangeError(w, err)
			return
		}
		middleware.WriteSuccess(w, req)
	case "complete":
		req, err := h.IPChanges.AdminCompleteRequest(r.Context(), requestID)
		if err != nil {
			writeIPChangeError(w, err)
			return
		}
		middleware.WriteSuccess(w, req)
	default:
		middleware.WriteError(w, http.StatusBadRequest, "unknown action")
	}
}

func (h *Handler) AdminDeleteIPChangeRequest(w http.ResponseWriter, r *http.Request) {
	requestID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request id")
		return
	}

	if err := h.IPChanges.AdminDeleteRequest(r.Context(), requestID); err != nil {
		writeIPChangeError(w, err)
		return
	}

	middleware.WriteSuccess(w, map[string]string{"status": "deleted"})
}
