package httpapi

import (
	"github.com/txyyddss/Remna-User-Panel/internal/iplookup"
	"net/http"
)

func (s *Server) ipLookupDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if !s.ipLookupMember(w, r) {
		return
	}
	if s.deps.IPLookup.Live == nil {
		s.ipLookupError(w, r, &iplookup.CodeError{Code: "IP_DETAILS_UNAVAILABLE"})
		return
	}
	result, err := s.deps.IPLookup.Live.Details(r.Context(), currentUser(r).ID, r.URL.Query().Get("ip"))
	if err != nil {
		s.ipLookupError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
