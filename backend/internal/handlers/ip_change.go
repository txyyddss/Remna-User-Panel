package handlers

import (
	"errors"
	"net/http"

	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/middleware"
	"github.com/user/remna-user-panel/internal/services"
)

// IPChange submits a new IP-change request and follows the reference
// queue-and-vote flow instead of disconnecting sessions immediately.
func (h *Handler) IPChange(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)

	var req struct {
		Subscription string `json:"subscription"`
		Reason       string `json:"reason"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.IPChanges.SubmitUserRequest(r.Context(), user, req.Subscription, req.Reason)
	if err != nil {
		writeIPChangeError(w, err)
		return
	}

	middleware.WriteSuccess(w, resp)
}

// GetIPChangeStatus keeps the legacy route name but now returns the same
// lookup data shape as the reference implementation.
func (h *Handler) GetIPChangeStatus(w http.ResponseWriter, r *http.Request) {
	h.GetIPChangeLookup(w, r)
}

// GetIPChangeLookup returns the current global request status.
func (h *Handler) GetIPChangeLookup(w http.ResponseWriter, r *http.Request) {
	resp, err := h.IPChanges.Lookup(r.Context())
	if err != nil {
		writeIPChangeError(w, err)
		return
	}

	middleware.WriteSuccess(w, resp)
}

// MarkIPSwapCompleted is the privileged callback used by the upstream swap
// automation once the actual IP replacement is done.
func (h *Handler) MarkIPSwapCompleted(w http.ResponseWriter, r *http.Request) {
	if !authorizedIPChangeToken(r.Header.Get("Authorization")) {
		middleware.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.IPChanges.MarkSwapCompleted(r.Context()); err != nil {
		writeIPChangeError(w, err)
		return
	}

	middleware.WriteSuccess(w, map[string]string{"status": "ok"})
}

// AddIPChangeRequestAPI submits a new vote-gated IP-change request through
// the protected automation endpoint.
func (h *Handler) AddIPChangeRequestAPI(w http.ResponseWriter, r *http.Request) {
	if !authorizedIPChangeToken(r.Header.Get("Authorization")) {
		middleware.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if r.ContentLength > 0 {
		if err := middleware.DecodeJSON(r, &req); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}

	requestID, err := h.IPChanges.AddAPIRequest(r.Context(), req.Reason)
	if err != nil {
		writeIPChangeError(w, err)
		return
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"success":    true,
		"message":    "ip change request added and waiting for votes",
		"request_id": requestID,
	})
}

func authorizedIPChangeToken(authHeader string) bool {
	token := config.Get().IPChange.SwapToken
	if token == "" {
		token = config.Get().Server.APISecret
	}
	if token == "" {
		return false
	}
	return authHeader == token || authHeader == "Bearer "+token
}

func writeIPChangeError(w http.ResponseWriter, err error) {
	var typedErr *services.IPChangeAPIError
	if errors.As(err, &typedErr) {
		if len(typedErr.Data) > 0 {
			middleware.WriteErrorData(w, typedErr.Status, typedErr.Message, typedErr.Data)
			return
		}
		middleware.WriteError(w, typedErr.Status, typedErr.Message)
		return
	}

	middleware.WriteError(w, http.StatusInternalServerError, err.Error())
}
