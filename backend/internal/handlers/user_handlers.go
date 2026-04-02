package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/middleware"
	"github.com/user/remna-user-panel/internal/sdk/remnawave"
)

const (
	queryActiveSub    = "SELECT id, user_id, combo_uuid, remnawave_uuid, status, expires_at, created_at FROM subscriptions WHERE user_id = ? AND status = 'active' ORDER BY created_at DESC LIMIT 1"
	queryJellyfinAcct = "SELECT id, user_id, jellyfin_user_id, username, parental_rating, expires_at FROM jellyfin_accounts WHERE user_id = ?"
	queryBoundCount   = "SELECT COUNT(*) FROM users WHERE remnawave_uuid = ? AND id != ?"
	queryBindUser     = "UPDATE users SET remnawave_uuid = ?, updated_at = ? WHERE id = ?"
	queryComboBySquad = "SELECT uuid FROM combos WHERE squad_uuid = ? ORDER BY created_at DESC LIMIT 1"
	queryInsertSub    = "INSERT INTO subscriptions (user_id, combo_uuid, remnawave_uuid, status, expires_at, created_at, updated_at) VALUES (?, ?, ?, 'active', ?, ?, ?)"
)

// GetMe returns the authenticated user's profile, active subscription,
// Jellyfin account (if any), and the public application configuration.
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUser(r)
	if user == nil {
		middleware.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	// Fetch active subscription (soft-fail: absence is normal for new users).
	var sub *struct {
		ID            int       `json:"id"`
		UserID        int       `json:"user_id"`
		ComboUUID     string    `json:"combo_uuid"`
		RemnawaveUUID string    `json:"remnawave_uuid"`
		Status        string    `json:"status"`
		ExpiresAt     time.Time `json:"expires_at"`
		CreatedAt     time.Time `json:"created_at"`
	}
	var s struct {
		ID            int
		UserID        int
		ComboUUID     string
		RemnawaveUUID string
		Status        string
		ExpiresAt     time.Time
		CreatedAt     time.Time
	}
	if err := database.DB().QueryRowContext(ctx, queryActiveSub, user.ID).
		Scan(&s.ID, &s.UserID, &s.ComboUUID, &s.RemnawaveUUID, &s.Status, &s.ExpiresAt, &s.CreatedAt); err == nil {
		sub = &struct {
			ID            int       `json:"id"`
			UserID        int       `json:"user_id"`
			ComboUUID     string    `json:"combo_uuid"`
			RemnawaveUUID string    `json:"remnawave_uuid"`
			Status        string    `json:"status"`
			ExpiresAt     time.Time `json:"expires_at"`
			CreatedAt     time.Time `json:"created_at"`
		}{s.ID, s.UserID, s.ComboUUID, s.RemnawaveUUID, s.Status, s.ExpiresAt, s.CreatedAt}
	}

	// Fetch Jellyfin account (soft-fail).
	var jfAccount interface{}
	var jf struct {
		ID             int
		UserID         int
		JellyfinUserID string
		Username       string
		ParentalRating int
		ExpiresAt      time.Time
	}
	if err := database.DB().QueryRowContext(ctx, queryJellyfinAcct, user.ID).
		Scan(&jf.ID, &jf.UserID, &jf.JellyfinUserID, &jf.Username, &jf.ParentalRating, &jf.ExpiresAt); err == nil {
		jfAccount = map[string]interface{}{
			"id":               jf.ID,
			"user_id":          jf.UserID,
			"jellyfin_user_id": jf.JellyfinUserID,
			"username":         jf.Username,
			"parental_rating":  jf.ParentalRating,
			"expires_at":       jf.ExpiresAt,
		}
	}

	cfg := config.Get()

	middleware.WriteSuccess(w, map[string]interface{}{
		"user":         user,
		"subscription": sub,
		"jellyfin":     jfAccount,
		"config": map[string]interface{}{
			"credit_name":     cfg.Credit.Name,
			"rmb_to_txb_rate": cfg.Credit.RMBToTXBRate,
			"txb_to_rmb_rate": cfg.Credit.TXBToRMBRate,
			"credit": map[string]interface{}{
				"name":            cfg.Credit.Name,
				"rmb_to_txb_rate": cfg.Credit.RMBToTXBRate,
				"txb_to_rmb_rate": cfg.Credit.TXBToRMBRate,
			},
			"jellyfin": map[string]interface{}{
				"monthly_price_rmb": cfg.Jellyfin.MonthlyPriceRMB,
			},
			"payments": map[string]interface{}{
				"usdt_networks": []map[string]string{
					{"value": "usdt.aptos", "label": "USDT Aptos"},
					{"value": "usdt.arbitrum", "label": "USDT Arbitrum"},
					{"value": "usdt.polygon", "label": "USDT Polygon"},
				},
			},
		},
	})
}

// BindSubscription links an existing Remnawave subscription to the
// authenticated panel user via short UUID extraction from a subscription URL.
func (h *Handler) BindSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUser(r)
	cfg := config.Get()

	if user.RemnawaveUUID != "" {
		middleware.WriteError(w, http.StatusConflict, "subscription already bound")
		return
	}

	var req struct {
		SubURL string `json:"sub_url"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.SubURL == "" {
		middleware.WriteError(w, http.StatusBadRequest, "please provide subscription URL")
		return
	}

	// Extract short UUID from the subscription link.
	shortUUID := extractShortUUID(req.SubURL)
	if shortUUID == "" {
		middleware.WriteError(w, http.StatusBadRequest, "could not extract short UUID from the provided URL")
		return
	}

	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)

	// Look up the Remnawave user by short UUID.
	rwUser, err := rwClient.GetUserByShortUUID(shortUUID)
	if err != nil {
		slog.Warn("bind-sub: remnawave lookup failed", "short_uuid", shortUUID, "error", err)
		middleware.WriteError(w, http.StatusNotFound, fmt.Sprintf("subscription not found: %v", err))
		return
	}

	// Prevent double-binding: ensure this Remnawave UUID isn't already bound to another panel user.
	var existingCount int
	if err := database.DB().QueryRowContext(ctx, queryBoundCount, rwUser.UUID, user.ID).Scan(&existingCount); err != nil {
		slog.Error("bind-sub: count query failed", "error", err)
	}
	if existingCount > 0 {
		middleware.WriteError(w, http.StatusConflict, "this subscription is already bound to another user")
		return
	}

	// Bind the UUID to this user.
	now := time.Now()
	if _, err := database.DB().ExecContext(ctx, queryBindUser, rwUser.UUID, now, user.ID); err != nil {
		slog.Error("bind-sub: failed to update user", "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "failed to bind subscription")
		return
	}

	// Auto-create a local subscription record if the Remnawave user belongs to a known squad.
	if len(rwUser.ActiveInternalSquads) > 0 {
		var comboUUID string
		_ = database.DB().QueryRowContext(ctx, queryComboBySquad, rwUser.ActiveInternalSquads[0].UUID).Scan(&comboUUID)
		if comboUUID != "" {
			if _, err := database.DB().ExecContext(ctx, queryInsertSub,
				user.ID, comboUUID, rwUser.UUID, rwUser.ExpireAt, now, now,
			); err != nil {
				slog.Error("bind-sub: failed to create subscription record", "error", err)
			}
		}
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"status":  "bound",
		"rw_user": rwUser.Username,
		"rw_uuid": rwUser.UUID,
		"expires": rwUser.ExpireAt,
	})
}
