package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/middleware"
	"github.com/user/remna-user-panel/internal/sdk/remnawave"
)

const (
	// ipLookupTimeout is the maximum duration to wait for the async IP lookup job.
	ipLookupTimeout = 10 * time.Second
	// ipLookupPoll is the interval between successive result checks.
	ipLookupPoll = time.Second
)

// ─── Subscription Info ───────────────────────────────────────────────

// GetSubInfo returns the Remnawave user profile for the authenticated user.
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
		slog.Error("sub-info: failed to fetch remnawave user", "uuid", user.RemnawaveUUID, "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "failed to get subscription info")
		return
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"has_subscription": true,
		"user":             rwUser,
	})
}

// GetSubKeys returns the connection credentials (subscription URL, VLESS
// UUID, Trojan password, etc.) for import into proxy clients.
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
		slog.Error("sub-keys: failed to fetch remnawave user", "uuid", user.RemnawaveUUID, "error", err)
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

// ─── VPN Info ────────────────────────────────────────────────────────

// GetBandwidthStats returns per-node bandwidth usage for the past month.
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
		slog.Error("bandwidth-stats: failed", "uuid", user.RemnawaveUUID, "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "failed to get bandwidth stats")
		return
	}
	middleware.WriteSuccess(w, stats)
}

// GetHWIDDevices returns the list of hardware-bound devices.
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
		slog.Error("hwid-devices: failed", "uuid", user.RemnawaveUUID, "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "failed to get devices")
		return
	}
	middleware.WriteSuccess(w, devices)
}

// GetIPList starts an async IP address lookup and polls for the result.
// The request context is respected so client disconnects stop polling.
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
		slog.Error("ip-list: job creation failed", "uuid", user.RemnawaveUUID, "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "failed to fetch IPs")
		return
	}

	// Poll for result with context-aware timeout.
	ctx, cancel := context.WithTimeout(r.Context(), ipLookupTimeout)
	defer cancel()

	ticker := time.NewTicker(ipLookupPoll)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			middleware.WriteError(w, http.StatusGatewayTimeout, "IP lookup timed out")
			return
		case <-ticker.C:
			result, err := rwClient.GetFetchIPsResult(jobID)
			if err == nil && result != nil {
				middleware.WriteSuccess(w, json.RawMessage(result))
				return
			}
		}
	}
}

// GetSubHistory returns the subscription fetch history logs.
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
		slog.Error("sub-history: failed", "uuid", user.RemnawaveUUID, "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "failed to get history")
		return
	}
	middleware.WriteSuccess(w, history)
}

// ─── External Squads ─────────────────────────────────────────────────

// GetExternalSquads returns all available external routing squads.
func (h *Handler) GetExternalSquads(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)

	squads, err := rwClient.GetExternalSquads()
	if err != nil {
		slog.Error("external-squads: failed", "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "failed to get squads")
		return
	}
	middleware.WriteSuccess(w, squads)
}

// UpdateExternalSquad changes the user's active external routing squad.
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
	if req.SquadUUID == "" {
		middleware.WriteError(w, http.StatusBadRequest, "squad_uuid is required")
		return
	}

	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
	if _, err := rwClient.UpdateUser(remnawave.UpdateUserRequest{
		UUID:              user.RemnawaveUUID,
		ExternalSquadUUID: req.SquadUUID,
	}); err != nil {
		slog.Error("update-squad: failed", "uuid", user.RemnawaveUUID, "squad", req.SquadUUID, "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "failed to update squad")
		return
	}

	middleware.WriteSuccess(w, map[string]string{"status": "updated"})
}
