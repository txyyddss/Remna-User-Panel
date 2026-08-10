package httpapi

import (
	"net/http"
	"time"
)

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := contextWithTimeout(r, 2*time.Second)
	defer cancel()
	if err := s.deps.Store.DB().PingContext(ctx); err != nil {
		s.writeError(w, r, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database is unavailable.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := contextWithTimeout(r, 3*time.Second)
	defer cancel()
	if err := s.deps.Store.DB().PingContext(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "degraded", "missing": []string{"database"}})
		return
	}
	combos, err := s.deps.Store.ListCombos(ctx, true)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "degraded", "missing": []string{"catalog"}})
		return
	}
	issues := s.deps.Settings.Readiness(ctx, len(combos))
	if len(issues) > 0 {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "setup_required", "missing": issues})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
