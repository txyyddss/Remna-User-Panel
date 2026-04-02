package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/middleware"
	"github.com/user/remna-user-panel/internal/sdk/remnawave"
)

func (h *Handler) GetSubInfo(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()

	if user.RemnawaveUUID == "" {
		middleware.WriteSuccess(w, map[string]interface{}{"has_subscription": false})
		return
	}

	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
	rwUser, err := rwClient.GetUserByUUID(user.RemnawaveUUID)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to get subscription info")
		return
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"has_subscription": true,
		"user":             rwUser,
	})
}

func (h *Handler) GetSubKeys(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()

	if user.RemnawaveUUID == "" {
		middleware.WriteError(w, http.StatusNotFound, "no active subscription")
		return
	}

	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
	rwUser, err := rwClient.GetUserByUUID(user.RemnawaveUUID)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to get keys")
		return
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"subscription_url": rwUser.SubscriptionURL,
		"short_uuid":       rwUser.ShortUUID,
		"vless_uuid":       rwUser.VlessUUID,
		"trojan_password":  rwUser.TrojanPassword,
		"ss_password":      rwUser.SSPassword,
		"username":         rwUser.Username,
		"instructions": []string{
			"Import the subscription URL into Clash, v2rayN, v2rayNG, Stash, Shadowrocket, or sing-box.",
			"If your client supports manual setup, use the VLESS UUID or Trojan password shown here.",
			"After changing route groups or resetting traffic, refresh the subscription in your client.",
			"Keep the subscription URL private. Anyone with the URL can import your profile.",
		},
	})
}

// --- VPN Info ---

func (h *Handler) GetBandwidthStats(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()

	if user.RemnawaveUUID == "" {
		middleware.WriteError(w, http.StatusNotFound, "no subscription")
		return
	}

	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
	start := time.Now().AddDate(0, -1, 0).Format(time.RFC3339)
	end := time.Now().Format(time.RFC3339)
	stats, err := rwClient.GetUserBandwidthStats(user.RemnawaveUUID, start, end)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to get bandwidth stats")
		return
	}
	middleware.WriteSuccess(w, stats)
}

func (h *Handler) GetHWIDDevices(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()

	if user.RemnawaveUUID == "" {
		middleware.WriteError(w, http.StatusNotFound, "no subscription")
		return
	}

	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
	devices, err := rwClient.GetUserHWIDDevices(user.RemnawaveUUID)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to get devices")
		return
	}
	middleware.WriteSuccess(w, devices)
}

func (h *Handler) GetIPList(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()

	if user.RemnawaveUUID == "" {
		middleware.WriteError(w, http.StatusNotFound, "no subscription")
		return
	}

	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
	jobID, err := rwClient.FetchUserIPs(user.RemnawaveUUID)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to fetch IPs")
		return
	}

	// Poll for result (max 10 seconds)
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		result, err := rwClient.GetFetchIPsResult(jobID)
		if err == nil && result != nil {
			middleware.WriteSuccess(w, json.RawMessage(result))
			return
		}
	}

	middleware.WriteError(w, http.StatusGatewayTimeout, "IP lookup timed out")
}

func (h *Handler) GetSubHistory(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()

	if user.RemnawaveUUID == "" {
		middleware.WriteError(w, http.StatusNotFound, "no subscription")
		return
	}

	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
	history, err := rwClient.GetUserSubHistory(user.RemnawaveUUID)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to get history")
		return
	}
	middleware.WriteSuccess(w, history)
}

// --- External Squads ---

func (h *Handler) GetExternalSquads(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
	squads, err := rwClient.GetExternalSquads()
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to get squads")
		return
	}
	middleware.WriteSuccess(w, squads)
}

func (h *Handler) UpdateExternalSquad(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()

	if user.RemnawaveUUID == "" {
		middleware.WriteError(w, http.StatusNotFound, "no subscription")
		return
	}

	var req struct {
		SquadUUID string `json:"squad_uuid"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
	_, err := rwClient.UpdateUser(remnawave.UpdateUserRequest{
		UUID:              user.RemnawaveUUID,
		ExternalSquadUUID: req.SquadUUID,
	})
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to update squad")
		return
	}

	middleware.WriteSuccess(w, map[string]string{"status": "updated"})
}
