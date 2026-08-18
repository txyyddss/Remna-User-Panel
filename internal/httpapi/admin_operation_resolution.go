package httpapi

import "net/http"

func (s *Server) adminResolveOperation(w http.ResponseWriter, r *http.Request) {
	key, ok := s.requireIdempotencyKey(w, r)
	if !ok {
		return
	}
	var request struct {
		Resolution string `json:"resolution"`
		Reason     string `json:"reason"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_OPERATION_RESOLUTION",
			"A supported resolution and audit reason are required.")
		return
	}
	receipt, err := s.deps.AdminUsers.ResolveOperation(r.Context(), currentUser(r).ID,
		chiURLParam(r, "operationId"), key, request.Resolution, request.Reason)
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, receipt)
}
