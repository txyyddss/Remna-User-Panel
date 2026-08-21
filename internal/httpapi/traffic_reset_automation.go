package httpapi

import "net/http"

type trafficResetAutomationRequest struct {
	Enabled *bool `json:"enabled"`
}

func (s *Server) trafficResetAutomation(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if !s.requireOnboarded(w, r, user) {
		return
	}
	setting, err := s.deps.PurchaseOperations.TrafficResetAutomation(r.Context(), user.ID)
	if err != nil {
		s.writeMemberOperationError(w, r, err, false)
		return
	}
	writeJSON(w, http.StatusOK, setting)
}

func (s *Server) updateTrafficResetAutomation(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if !s.requireOnboarded(w, r, user) {
		return
	}
	var request trafficResetAutomationRequest
	if err := decodeJSON(w, r, &request); err != nil || request.Enabled == nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_TRAFFIC_RESET_AUTOMATION", "Provide the automatic traffic reset state.")
		return
	}
	setting, err := s.deps.PurchaseOperations.SetTrafficResetAutomation(r.Context(), user.ID, *request.Enabled)
	if err != nil {
		s.writeMemberOperationError(w, r, err, true)
		return
	}
	writeJSON(w, http.StatusOK, setting)
}
