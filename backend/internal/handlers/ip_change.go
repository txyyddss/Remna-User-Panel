package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/middleware"
	"github.com/user/remna-user-panel/internal/sdk/remnawave"
)

// IPChange handles the IP change request with cooldown
func (h *Handler) IPChange(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()

	if user.RemnawaveUUID == "" {
		middleware.WriteError(w, http.StatusNotFound, "no active subscription")
		return
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

	// Get current IPs
	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)

	// Drop current connections to force IP change
	err = rwClient.DropConnections([]string{user.RemnawaveUUID})
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
