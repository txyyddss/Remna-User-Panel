package httpapi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/txyyddss/Remna-User-Panel/internal/catalog"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

type renewalRequest struct {
	TermCount int `json:"termCount"`
}

func (s *Server) renewalQuote(w http.ResponseWriter, r *http.Request) {
	var request renewalRequest
	if err := decodeJSON(w, r, &request); err != nil || request.TermCount < 1 || request.TermCount > 6 {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_RENEWAL", "Choose between 1 and 6 renewal terms.")
		return
	}
	quote, err := s.deps.Catalog.RenewalQuote(r.Context(), currentUser(r), chi.URLParam(r, "id"), request.TermCount)
	if err != nil {
		s.writeRenewalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, quote)
}

func (s *Server) renewal(w http.ResponseWriter, r *http.Request) {
	var request renewalRequest
	if err := decodeJSON(w, r, &request); err != nil || request.TermCount < 1 || request.TermCount > 6 {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_RENEWAL", "Choose between 1 and 6 renewal terms.")
		return
	}
	key, err := optionalOrGeneratedIdempotencyKey(w, r)
	if err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_IDEMPOTENCY_KEY", "Idempotency-Key must contain 1 to 128 characters.")
		return
	}
	batch, err := s.deps.Catalog.Renew(r.Context(), currentUser(r), chi.URLParam(r, "id"), request.TermCount, key)
	if err != nil {
		s.writeRenewalError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, batch)
}

func (s *Server) writeRenewalError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := http.StatusConflict, "RENEWAL_CONFLICT", "The renewal could not be completed."
	if errors.Is(err, catalog.ErrNoAccessibleNodes) {
		code, message = "NO_ACCESSIBLE_NODES", "The current ride has no accessible nodes."
	}
	if errors.Is(err, database.ErrInsufficientBalance) {
		code, message = "INSUFFICIENT_BALANCE", "Your TXB balance is too low for this renewal."
	}
	if errors.Is(err, database.ErrNotFound) {
		status, code, message = http.StatusNotFound, "PURCHASE_NOT_FOUND", "The current ride could not be found."
	}
	if errors.Is(err, database.ErrStockUnavailable) {
		code, message = "SQUAD_STOCK_UNAVAILABLE", "A selected squad is currently full."
	}
	s.writeError(w, r, status, code, message)
}
