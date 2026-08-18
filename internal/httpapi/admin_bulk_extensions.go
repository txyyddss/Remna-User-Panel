package httpapi

import (
	"net/http"

	"github.com/txyyddss/Remna-User-Panel/internal/admin"
)

type bulkExtensionRequest struct {
	ComboIDs        []string `json:"comboIds"`
	AddonSquadUUIDs []string `json:"addonSquadUuids"`
	Days            int      `json:"days"`
	Reason          string   `json:"reason"`
}

func (request bulkExtensionRequest) domain() admin.BulkExtension {
	return admin.BulkExtension{ComboIDs: request.ComboIDs, AddonSquadUUIDs: request.AddonSquadUUIDs,
		Days: request.Days, Reason: request.Reason}
}

func (s *Server) adminPreviewBulkExtension(w http.ResponseWriter, r *http.Request) {
	var request bulkExtensionRequest
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_BULK_EXTENSION", "Filters and extension days are required.")
		return
	}
	preview, err := s.deps.AdminUsers.PreviewBulkExtension(r.Context(), request.domain())
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, preview)
}

func (s *Server) adminCreateBulkExtension(w http.ResponseWriter, r *http.Request) {
	key, ok := s.requireIdempotencyKey(w, r)
	if !ok {
		return
	}
	var request bulkExtensionRequest
	if err := decodeJSON(w, r, &request); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_BULK_EXTENSION", "Filters, days, and reason are required.")
		return
	}
	receipt, err := s.deps.AdminUsers.CreateBulkExtension(r.Context(), currentUser(r).ID, key, request.domain())
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusAccepted, receipt)
}
