package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/middleware"
	"github.com/user/remna-user-panel/internal/sdk/jellyfin"
	"github.com/user/remna-user-panel/internal/services"
)

const (
	queryJellyfinExpiry  = "SELECT expires_at FROM jellyfin_accounts WHERE user_id = ?"
	queryUpdateParentRtg = "UPDATE jellyfin_accounts SET parental_rating = ? WHERE user_id = ?"

	// maxParentalRating is the upper bound allowed by the Jellyfin API.
	maxParentalRating = 22
)

// PurchaseJellyfin creates a payment order for Jellyfin media access.
// If the user already has an active account, the new months stack onto
// the current expiry rather than starting from now.
func (h *Handler) PurchaseJellyfin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUser(r)
	cfg := config.Get()

	var req struct {
		Months        int     `json:"months"`
		PaymentMethod string  `json:"payment_method"`
		PaymentType   string  `json:"payment_type"`
		UseTXB        bool    `json:"use_txb"`
		DiscountRMB   float64 `json:"discount_rmb"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.Months <= 0 {
		req.Months = 1
	}
	if req.Months > 24 {
		middleware.WriteError(w, http.StatusBadRequest, "maximum 24 months per purchase")
		return
	}

	amount := cfg.Jellyfin.MonthlyPriceRMB * float64(req.Months)

	// Stack onto existing expiry if still valid, otherwise start from now.
	baseTime := time.Now()
	var currentExpiry time.Time
	if err := database.DB().QueryRowContext(ctx, queryJellyfinExpiry, user.ID).
		Scan(&currentExpiry); err == nil && currentExpiry.After(baseTime) {
		baseTime = currentExpiry
	}
	expiry := baseTime.AddDate(0, req.Months, 0)

	metadata, _ := json.Marshal(map[string]interface{}{
		"months": req.Months,
		"expiry": expiry.Format(time.RFC3339),
	})

	payResp, err := h.Payment.CreatePayment(ctx, user.ID, services.CreatePaymentRequest{
		OrderType:     "jellyfin",
		Amount:        amount,
		PaymentMethod: req.PaymentMethod,
		PaymentType:   req.PaymentType,
		UseTXB:        req.UseTXB,
		DiscountRMB:   req.DiscountRMB,
		Metadata:      string(metadata),
		ClientIP:      r.RemoteAddr,
	})
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	middleware.WriteSuccess(w, payResp)
}

// JellyfinQuickConnect authorises a Quick Connect code for the user's
// Jellyfin account, allowing passwordless device pairing.
func (h *Handler) JellyfinQuickConnect(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()

	if user.JellyfinUserID == "" {
		middleware.WriteError(w, http.StatusNotFound, "no Jellyfin account")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.Code == "" {
		middleware.WriteError(w, http.StatusBadRequest, "code is required")
		return
	}

	jfClient := jellyfin.NewClient(cfg.Jellyfin.URL, cfg.Jellyfin.Token)

	enabled, err := jfClient.IsQuickConnectEnabled()
	if err != nil {
		slog.Error("jellyfin-qc: availability check failed", "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "Quick Connect unavailable")
		return
	}
	if !enabled {
		middleware.WriteError(w, http.StatusBadRequest, "Quick Connect is disabled on the Jellyfin server")
		return
	}

	if err := jfClient.AuthorizeQuickConnect(user.JellyfinUserID, req.Code); err != nil {
		slog.Error("jellyfin-qc: authorization failed", "user_id", user.JellyfinUserID, "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "Quick Connect authorization failed")
		return
	}

	middleware.WriteSuccess(w, map[string]string{"status": "authorized"})
}

// JellyfinUpdatePassword changes the Jellyfin user's password.
func (h *Handler) JellyfinUpdatePassword(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()

	if user.JellyfinUserID == "" {
		middleware.WriteError(w, http.StatusNotFound, "no Jellyfin account")
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.NewPassword == "" {
		middleware.WriteError(w, http.StatusBadRequest, "new password is required")
		return
	}

	jfClient := jellyfin.NewClient(cfg.Jellyfin.URL, cfg.Jellyfin.Token)
	if err := jfClient.UpdatePassword(user.JellyfinUserID, req.CurrentPassword, req.NewPassword); err != nil {
		slog.Error("jellyfin-pw: update failed", "user_id", user.JellyfinUserID, "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "password change failed")
		return
	}

	middleware.WriteSuccess(w, map[string]string{"status": "updated"})
}

// JellyfinGetDevices returns the list of devices for the user's Jellyfin account.
func (h *Handler) JellyfinGetDevices(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()

	if user.JellyfinUserID == "" {
		middleware.WriteError(w, http.StatusNotFound, "no Jellyfin account")
		return
	}

	jfClient := jellyfin.NewClient(cfg.Jellyfin.URL, cfg.Jellyfin.Token)
	devices, err := jfClient.GetDevices(user.JellyfinUserID)
	if err != nil {
		slog.Error("jellyfin-devices: failed", "user_id", user.JellyfinUserID, "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "failed to get devices")
		return
	}

	middleware.WriteSuccess(w, devices)
}

// JellyfinUpdateParentalRating updates the content rating filter for
// the user's Jellyfin account. Valid range is 0 (unrestricted) to 22.
func (h *Handler) JellyfinUpdateParentalRating(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUser(r)
	cfg := config.Get()

	if user.JellyfinUserID == "" {
		middleware.WriteError(w, http.StatusNotFound, "no Jellyfin account")
		return
	}

	var req struct {
		Rating int `json:"rating"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.Rating < 0 || req.Rating > maxParentalRating {
		middleware.WriteError(w, http.StatusBadRequest, "rating must be between 0 and 22")
		return
	}

	jfClient := jellyfin.NewClient(cfg.Jellyfin.URL, cfg.Jellyfin.Token)
	if err := jfClient.UpdateParentalRating(user.JellyfinUserID, req.Rating); err != nil {
		slog.Error("jellyfin-rating: update failed", "user_id", user.JellyfinUserID, "rating", req.Rating, "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "failed to update rating")
		return
	}

	// Sync local record.
	if _, err := database.DB().ExecContext(ctx, queryUpdateParentRtg, req.Rating, user.ID); err != nil {
		slog.Error("jellyfin-rating: local db update failed", "error", err)
	}

	middleware.WriteSuccess(w, map[string]string{"status": "updated"})
}
