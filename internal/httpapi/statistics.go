package httpapi

import (
	"net/http"
	"time"
)

func (s *Server) statisticsSnapshot(w http.ResponseWriter, r *http.Request) {
	snapshot, err := s.deps.Statistics.Snapshot(r.Context())
	if err != nil {
		s.writeError(w, r, http.StatusServiceUnavailable, "STATISTICS_UNAVAILABLE", "Statistics are not available yet.")
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (s *Server) statisticsNodes(w http.ResponseWriter, r *http.Request) {
	snapshot, err := s.deps.Statistics.NodeSnapshot(r.Context(), time.Now().UTC())
	if err != nil {
		s.deps.Logger.Warn("load live node statistics", "request_id", middlewareRequestID(r), "error", err)
		s.writeError(w, r, http.StatusBadGateway, "NODE_STATISTICS_UNAVAILABLE", "Node statistics are temporarily unavailable.")
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (s *Server) statisticsNodeGeocheck(w http.ResponseWriter, r *http.Request) {
	if squadUUID := r.URL.Query().Get("squadUuid"); squadUUID != "" {
		enabled, err := s.deps.Store.SquadGeocheckEnabled(r.Context(), squadUUID)
		if err != nil {
			s.writeError(w, r, http.StatusServiceUnavailable, "NODE_GEOCHECK_UNAVAILABLE", "Squad geocheck settings are unavailable.")
			return
		}
		if !enabled {
			s.writeError(w, r, http.StatusForbidden, "SQUAD_GEOCHECK_DISABLED", "Geocheck disabled.")
			return
		}
	}
	result, ok := s.deps.Statistics.NodeGeocheck(chiURLParam(r, "nodeUuid"))
	if !ok {
		s.writeError(w, r, http.StatusNotFound, "NODE_GEOCHECK_UNAVAILABLE", "Node geocheck is not available yet.")
		return
	}
	writeJSON(w, http.StatusOK, result)
}
