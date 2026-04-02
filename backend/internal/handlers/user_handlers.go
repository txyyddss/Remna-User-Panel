package handlers

import (
	"net/http"
	"time"

	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/middleware"
	"github.com/user/remna-user-panel/internal/models"
	"github.com/user/remna-user-panel/internal/sdk/remnawave"
)

const (
	userQuery1 = "SELECT id, user_id, combo_uuid, remnawave_uuid, status, expires_at, created_at FROM subscriptions WHERE user_id = ? AND status = 'active' ORDER BY created_at DESC LIMIT 1"
	userQuery2 = "SELECT id, user_id, jellyfin_user_id, username, parental_rating, expires_at FROM jellyfin_accounts WHERE user_id = ?"
	userQuery3 = "SELECT COUNT(*) FROM users WHERE remnawave_uuid = ? AND id != ?"
	userQuery4 = "UPDATE users SET remnawave_uuid = ?, updated_at = ? WHERE id = ?"
	userQuery5 = "SELECT uuid FROM combos WHERE squad_uuid = ? ORDER BY created_at DESC LIMIT 1"
	userQuery6 = "INSERT INTO subscriptions (user_id, combo_uuid, remnawave_uuid, status, expires_at, created_at, updated_at) VALUES (?, ?, ?, 'active', ?, ?, ?)"
)

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		middleware.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	// Get subscription info
	var sub *models.Subscription
	var subData models.Subscription
	err := database.DB().QueryRowContext(r.Context(), 
		userQuery1,
		user.ID,
	).Scan(&subData.ID, &subData.UserID, &subData.ComboUUID, &subData.RemnawaveUUID, &subData.Status, &subData.ExpiresAt, &subData.CreatedAt)
	if err == nil {
		sub = &subData
	}

	// Get Jellyfin account info
	var jfAccount *models.JellyfinAccount
	var jf models.JellyfinAccount
	err = database.DB().QueryRowContext(r.Context(), 
		userQuery2,
		user.ID,
	).Scan(&jf.ID, &jf.UserID, &jf.JellyfinUserID, &jf.Username, &jf.ParentalRating, &jf.ExpiresAt)
	if err == nil {
		jfAccount = &jf
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

func (h *Handler) BindSubscription(w http.ResponseWriter, r *http.Request) {
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

	// Extract short UUID from subscription URL
	// Typical format: https://panel.example.com/api/sub/SHORTUUID or just the short UUID
	shortUUID := req.SubURL
	// Try to extract from URL path
	if idx := len(shortUUID) - 1; idx >= 0 {
		parts := make([]string, 0)
		for _, p := range splitURL(shortUUID) {
			if p != "" {
				parts = append(parts, p)
			}
		}
		if len(parts) > 0 {
			shortUUID = parts[len(parts)-1]
		}
	}

	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)

	// Look up the user by short UUID
	rwUser, err := rwClient.GetUserByShortUUID(shortUUID)
	if err != nil {
		middleware.WriteError(w, http.StatusNotFound, "subscription not found: "+err.Error())
		return
	}

	// Verify the Remnawave user isn't already bound to another panel user
	var existingCount int
	database.DB().QueryRowContext(r.Context(), userQuery3, rwUser.UUID, user.ID).Scan(&existingCount)
	if existingCount > 0 {
		middleware.WriteError(w, http.StatusConflict, "this subscription is already bound to another user")
		return
	}

	// Bind user
	database.DB().ExecContext(r.Context(), userQuery4, rwUser.UUID, time.Now(), user.ID)

	var comboUUID string
	if len(rwUser.ActiveInternalSquads) > 0 {
		_ = database.DB().QueryRowContext(r.Context(), 
			userQuery5,
			rwUser.ActiveInternalSquads[0].UUID,
		).Scan(&comboUUID)
	}
	if comboUUID != "" {
		database.DB().ExecContext(r.Context(), 
			userQuery6,
			user.ID, comboUUID, rwUser.UUID, rwUser.ExpireAt, time.Now(), time.Now(),
		)
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"status":  "bound",
		"rw_user": rwUser.Username,
		"rw_uuid": rwUser.UUID,
		"expires": rwUser.ExpireAt,
	})
}

func splitURL(u string) []string {
	// Simple URL path splitter
	result := make([]string, 0)
	current := ""
	for _, c := range u {
		if c == '/' || c == '?' {
			if current != "" {
				result = append(result, current)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
