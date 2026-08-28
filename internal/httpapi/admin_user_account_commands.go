package httpapi

import "net/http"

func (s *Server) adminGrantUserCoupon(w http.ResponseWriter, r *http.Request) {
	key, ok := s.requireIdempotencyKey(w, r)
	if !ok {
		return
	}
	var request struct {
		CouponID string `json:"couponId"`
		Reason   string `json:"reason"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_ADMIN_COUPON_GRANT", "Coupon and reason are required.")
		return
	}
	grant, err := s.deps.AdminUsers.GrantCoupon(r.Context(), currentUser(r).ID, chiURLParam(r, "userId"), request.CouponID, key, request.Reason)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, grant)
}

func (s *Server) adminDiscardUserCoupon(w http.ResponseWriter, r *http.Request) {
	key, ok := s.requireIdempotencyKey(w, r)
	if !ok {
		return
	}
	if err := s.deps.AdminUsers.DiscardCoupon(r.Context(), currentUser(r).ID, chiURLParam(r, "userId"), chiURLParam(r, "grantId"), key); err != nil {
		s.adminFailure(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
