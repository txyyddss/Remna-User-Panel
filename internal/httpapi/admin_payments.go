package httpapi

import (
	"net/http"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

func (s *Server) adminPayments(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Store.ListPaymentOrders(r.Context(), "", 200)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	type adminPayment struct {
		model.PaymentOrder
		UserID string `json:"userId"`
	}
	response := make([]adminPayment, 0, len(items))
	for _, item := range items {
		response = append(response, adminPayment{PaymentOrder: item, UserID: item.UserID})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": response})
}

func (s *Server) adminRefundPayment(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_REFUND", "A refund reason is required.")
		return
	}
	order, err := s.deps.Admin.Refund(r.Context(), currentUser(r).ID, chiURLParam(r, "id"), request.Reason)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (s *Server) adminRefunds(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Store.ListRefunds(r.Context(), 200)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
