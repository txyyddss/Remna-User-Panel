package httpapi

import (
	"net/http"
)

func (s *Server) adminSettings(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Admin.Settings(r.Context())
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) adminCreateSetting(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := decodeJSON(w, r, &request); err != nil || request.Key == "" {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_SETTING", "A known setting key and value are required.")
		return
	}
	if err := s.deps.Admin.PutSetting(r.Context(), currentUser(r).ID, request.Key, request.Value); err != nil {
		s.adminFailure(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) adminUpdateSetting(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Value string `json:"value"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_SETTING", "A setting value is required.")
		return
	}
	if err := s.deps.Admin.PutSetting(r.Context(), currentUser(r).ID, chiURLParam(r, "key"), request.Value); err != nil {
		s.adminFailure(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
