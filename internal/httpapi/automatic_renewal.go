package httpapi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/txyyddss/Remna-User-Panel/internal/catalog"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

type automaticRenewalRequest struct {
	Enabled *bool `json:"enabled"`
}

func (s *Server) automaticRenewal(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if !s.requireOnboarded(w, r, user) {
		return
	}
	status, err := s.deps.Catalog.AutomaticRenewal(r.Context(), user, chi.URLParam(r, "id"))
	if err != nil {
		s.writeAutomaticRenewalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) updateAutomaticRenewal(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if !s.requireOnboarded(w, r, user) {
		return
	}
	var request automaticRenewalRequest
	if err := decodeJSON(w, r, &request); err != nil || request.Enabled == nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_AUTO_RENEWAL", "Provide the automatic renewal state.")
		return
	}
	status, err := s.deps.Catalog.SetAutomaticRenewal(r.Context(), user, chi.URLParam(r, "id"), *request.Enabled)
	if err != nil {
		if errors.Is(err, catalog.ErrAutoRenewalIneligible) {
			s.writeError(w, r, http.StatusConflict, "AUTO_RENEWAL_INELIGIBLE", "Automatic renewal cannot be enabled for this term.")
			return
		}
		s.writeAutomaticRenewalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) writeAutomaticRenewalError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, database.ErrNotFound):
		s.writeError(w, r, http.StatusNotFound, "PURCHASE_NOT_FOUND", "The current ride could not be found.")
	case errors.Is(err, database.ErrConflict):
		s.writeError(w, r, http.StatusConflict, "AUTO_RENEWAL_CONFLICT", "Automatic renewal is no longer available for this term.")
	default:
		s.writeError(w, r, http.StatusBadGateway, "AUTO_RENEWAL_UNAVAILABLE", "Automatic renewal is temporarily unavailable.")
	}
}
