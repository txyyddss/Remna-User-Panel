package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

func (s *Server) adminPaymentProfiles(w http.ResponseWriter, r *http.Request) {
	items, err := s.deps.Settings.PaymentProfiles(r.Context())
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) adminSavePaymentProfile(w http.ResponseWriter, r *http.Request) {
	var profile struct {
		ID              string   `json:"id"`
		ProviderName    string   `json:"providerName"`
		EnabledChannels []string `json:"enabledChannels"`
		Endpoint        string   `json:"endpoint"`
		MerchantID      string   `json:"merchantId"`
		Credential      string   `json:"credential"`
		Acknowledgement string   `json:"acknowledgement"`
		Enabled         bool     `json:"enabled"`
	}
	if err := decodeJSON(w, r, &profile); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_PAYMENT_PROFILE", "Payment profile fields are invalid.")
		return
	}
	provider := chi.URLParam(r, "provider")
	saved, err := s.deps.Settings.SavePaymentProfile(r.Context(), currentUser(r).ID, model.PaymentProfile{
		ID: profile.ID, Provider: provider, ProviderName: profile.ProviderName, EnabledChannels: profile.EnabledChannels,
		Endpoint: profile.Endpoint, MerchantID: profile.MerchantID, Credential: profile.Credential,
		Acknowledgement: profile.Acknowledgement, Enabled: profile.Enabled,
	})
	if err != nil {
		s.adminFailure(w, r, err)
		return
	}
	actorID := currentUser(r).ID
	_ = s.deps.Store.AppendAudit(r.Context(), &actorID, "payment_profile.update", "payment_profile", saved.ID, "{}", time.Now().UTC())
	writeJSON(w, http.StatusOK, saved)
}
