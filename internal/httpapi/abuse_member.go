package httpapi

import (
	"net/http"
	"strconv"
)

func (s *Server) memberAbuseRecords(w http.ResponseWriter, r *http.Request) {
	limit := 25
	if raw := r.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 100 {
			s.writeError(w, r, http.StatusBadRequest, "INVALID_PAGE_SIZE", "Page size must be between 1 and 100.")
			return
		}
		limit = value
	}
	page, err := s.deps.Abuse.MemberRecords(r.Context(), currentUser(r).ID, r.URL.Query().Get("cursor"), limit)
	if err != nil {
		s.writeError(w, r, http.StatusUnprocessableEntity, "ABUSE_RECORDS_UNAVAILABLE", "Abuse records are unavailable.")
		return
	}
	writeJSON(w, http.StatusOK, page)
}
