package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/middleware"
	"github.com/user/remna-user-panel/internal/models"
	"github.com/user/remna-user-panel/internal/sdk/remnawave"
)

// IPChange handles the IP change request with cooldown
func (h *Handler) IPChange(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()

	var req struct {
		Subscription string `json:"subscription"`
	}
	if r.ContentLength > 0 {
		if err := middleware.DecodeJSON(r, &req); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid request")
			return
		}
	}

	// Check cooldown
	cooldownHours := cfg.IPChange.CooldownHours
	if cooldownHours <= 0 {
		cooldownHours = 6
	}

	var lastChange time.Time
	err := database.DB().QueryRow(
		"SELECT created_at FROM ip_change_logs WHERE user_id = ? ORDER BY created_at DESC LIMIT 1",
		user.ID,
	).Scan(&lastChange)

	if err == nil {
		sinceLastChange := time.Since(lastChange)
		cooldownDuration := time.Duration(cooldownHours) * time.Hour
		if sinceLastChange < cooldownDuration {
			remaining := cooldownDuration - sinceLastChange
			hours := int(remaining.Hours())
			minutes := int(remaining.Minutes()) % 60
			middleware.WriteError(w, http.StatusTooManyRequests,
				fmt.Sprintf("IP change on cooldown, please wait %d hour(s) %d minute(s)", hours, minutes))
			return
		}
	}

	targetUUID := user.RemnawaveUUID
	if targetUUID == "" || req.Subscription != "" {
		resolvedUUID, err := h.resolveRemnawaveUUID(req.Subscription, user)
		if err != nil {
			middleware.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		targetUUID = resolvedUUID
	}

	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)

	// Drop current connections to force IP change
	err = rwClient.DropConnections([]string{targetUUID})
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to disconnect: "+err.Error())
		return
	}

	// Log the change
	database.DB().Exec(
		"INSERT INTO ip_change_logs (user_id, created_at) VALUES (?, ?)",
		user.ID, time.Now(),
	)

	middleware.WriteSuccess(w, map[string]interface{}{
		"status":            "success",
		"message":           "connection dropped, please reconnect to get a new IP",
		"subscription":      req.Subscription,
		"next_change_after": time.Now().Add(time.Duration(cooldownHours) * time.Hour).Format(time.RFC3339),
	})
}

// GetIPChangeStatus returns the cooldown status
func (h *Handler) GetIPChangeStatus(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()

	cooldownHours := cfg.IPChange.CooldownHours
	if cooldownHours <= 0 {
		cooldownHours = 6
	}

	var lastChange time.Time
	err := database.DB().QueryRow(
		"SELECT created_at FROM ip_change_logs WHERE user_id = ? ORDER BY created_at DESC LIMIT 1",
		user.ID,
	).Scan(&lastChange)

	canChange := true
	var nextAvailable time.Time
	if err == nil {
		cooldownDuration := time.Duration(cooldownHours) * time.Hour
		nextAvailable = lastChange.Add(cooldownDuration)
		if time.Now().Before(nextAvailable) {
			canChange = false
		}
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"can_change":     canChange,
		"last_change":    lastChange,
		"next_available": nextAvailable,
		"cooldown_hours": cooldownHours,
	})
}

func (h *Handler) resolveRemnawaveUUID(subscription string, user *models.User) (string, error) {
	if user != nil && user.RemnawaveUUID != "" && subscription == "" {
		return user.RemnawaveUUID, nil
	}

	raw := strings.TrimSpace(subscription)
	if raw == "" {
		return "", fmt.Errorf("subscription link or short UUID is required")
	}

	shortUUID := raw
	if strings.Contains(raw, "/") {
		parts := strings.FieldsFunc(raw, func(r rune) bool {
			return r == '/' || r == '?' || r == '#'
		})
		if len(parts) == 0 {
			return "", fmt.Errorf("invalid subscription link")
		}
		shortUUID = parts[len(parts)-1]
	}

	cfg := config.Get()
	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
	rwUser, err := rwClient.GetUserByShortUUID(shortUUID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve subscription")
	}
	if rwUser.Status != "ACTIVE" && rwUser.Status != "LIMITED" {
		return "", fmt.Errorf("subscription is not active")
	}
	return rwUser.UUID, nil
}
