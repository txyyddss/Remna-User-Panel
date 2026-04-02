package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/middleware"
	"github.com/user/remna-user-panel/internal/models"
	"github.com/user/remna-user-panel/internal/sdk/remnawave"
)

const (
	defaultCooldownHours = 6

	ipChangeQueryLastChange = "SELECT created_at FROM ip_change_logs WHERE user_id = ? ORDER BY created_at DESC LIMIT 1"
	ipChangeQueryInsertLog  = "INSERT INTO ip_change_logs (user_id, created_at) VALUES (?, ?)"
)

// IPChange handles the IP change request with cooldown enforcement.
// It drops a user's active Remnawave connections so the next reconnect
// picks up a fresh exit IP, then logs the event for cooldown tracking.
func (h *Handler) IPChange(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUser(r)
	cfg := config.Get()

	var req struct {
		Subscription string `json:"subscription"`
	}
	if r.ContentLength > 0 {
		if err := middleware.DecodeJSON(r, &req); err != nil {
			middleware.WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}

	// Resolve cooldown window
	cooldownHours := cfg.IPChange.CooldownHours
	if cooldownHours <= 0 {
		cooldownHours = defaultCooldownHours
	}
	cooldownDuration := time.Duration(cooldownHours) * time.Hour

	// Check cooldown against last change timestamp
	var lastChange time.Time
	err := database.DB().QueryRowContext(ctx, ipChangeQueryLastChange, user.ID).Scan(&lastChange)
	if err == nil {
		remaining := cooldownDuration - time.Since(lastChange)
		if remaining > 0 {
			hours := int(remaining.Hours())
			minutes := int(remaining.Minutes()) % 60
			middleware.WriteError(w, http.StatusTooManyRequests,
				fmt.Sprintf("IP change on cooldown, please wait %dh %dm", hours, minutes))
			return
		}
	}

	// Resolve the target Remnawave UUID
	targetUUID, err := h.resolveRemnawaveUUID(req.Subscription, user)
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Drop active connections to force IP rotation
	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
	if err := rwClient.DropConnections([]string{targetUUID}); err != nil {
		slog.Error("ip-change: drop connections failed", "user_id", user.ID, "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "failed to disconnect active sessions")
		return
	}

	// Record the change for cooldown tracking
	if _, err := database.DB().ExecContext(ctx, ipChangeQueryInsertLog, user.ID, time.Now()); err != nil {
		slog.Error("ip-change: failed to log change", "user_id", user.ID, "error", err)
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"status":            "success",
		"message":           "connection dropped, please reconnect to get a new IP",
		"subscription":      req.Subscription,
		"next_change_after": time.Now().Add(cooldownDuration).Format(time.RFC3339),
	})
}

// GetIPChangeStatus returns the current cooldown status for the requesting user.
func (h *Handler) GetIPChangeStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUser(r)
	cfg := config.Get()

	cooldownHours := cfg.IPChange.CooldownHours
	if cooldownHours <= 0 {
		cooldownHours = defaultCooldownHours
	}

	var lastChange time.Time
	err := database.DB().QueryRowContext(ctx, ipChangeQueryLastChange, user.ID).Scan(&lastChange)

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

// resolveRemnawaveUUID determines the Remnawave user UUID from either
// the authenticated user's stored UUID or a subscription link/short UUID.
func (h *Handler) resolveRemnawaveUUID(subscription string, user *models.User) (string, error) {
	// Fast path: user already has a bound UUID and no override was provided.
	if user != nil && user.RemnawaveUUID != "" && subscription == "" {
		return user.RemnawaveUUID, nil
	}

	raw := strings.TrimSpace(subscription)
	if raw == "" {
		return "", fmt.Errorf("subscription link or short UUID is required")
	}

	// Extract the last path segment as the short UUID.
	shortUUID := extractShortUUID(raw)
	if shortUUID == "" {
		return "", fmt.Errorf("invalid subscription link format")
	}

	cfg := config.Get()
	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
	rwUser, err := rwClient.GetUserByShortUUID(shortUUID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve subscription: %w", err)
	}
	if rwUser.Status != "ACTIVE" && rwUser.Status != "LIMITED" {
		return "", fmt.Errorf("subscription is not active (status: %s)", rwUser.Status)
	}
	return rwUser.UUID, nil
}

// extractShortUUID pulls the last meaningful path segment from a URL or
// returns the raw string if it contains no separators.
func extractShortUUID(raw string) string {
	if !strings.ContainsAny(raw, "/?#") {
		return raw
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '/' || r == '?' || r == '#'
	})
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
