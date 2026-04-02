package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/middleware"
	"github.com/user/remna-user-panel/internal/sdk/jellyfin"
	"github.com/user/remna-user-panel/internal/services"
)

const (
	jellyfinQuery1 = "SELECT expires_at FROM jellyfin_accounts WHERE user_id = ?"
	jellyfinQuery2 = "UPDATE jellyfin_accounts SET parental_rating = ? WHERE user_id = ?"
)

func (h *Handler) PurchaseJellyfin(w http.ResponseWriter, r *http.Request) {
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

	amount := cfg.Jellyfin.MonthlyPriceRMB * float64(req.Months)
	baseTime := time.Now()
	var currentExpiry time.Time
	if err := database.DB().QueryRowContext(r.Context(),
		jellyfinQuery1,
		user.ID,
	).Scan(&currentExpiry); err == nil && currentExpiry.After(baseTime) {
		baseTime = currentExpiry
	}
	expiry := baseTime.AddDate(0, req.Months, 0)

	metadata, _ := json.Marshal(map[string]interface{}{
		"months": req.Months,
		"expiry": expiry.Format(time.RFC3339),
	})

	payResp, err := h.Payment.CreatePayment(user.ID, services.CreatePaymentRequest{
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
		middleware.WriteError(w, http.StatusInternalServerError, "Quick Connect unavailable: "+err.Error())
		return
	}
	if !enabled {
		middleware.WriteError(w, http.StatusBadRequest, "Quick Connect is disabled on the Jellyfin server")
		return
	}
	if err := jfClient.AuthorizeQuickConnect(user.JellyfinUserID, req.Code); err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "Quick Connect failed: "+err.Error())
		return
	}

	middleware.WriteSuccess(w, map[string]string{"status": "authorized"})
}

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

	jfClient := jellyfin.NewClient(cfg.Jellyfin.URL, cfg.Jellyfin.Token)
	if err := jfClient.UpdatePassword(user.JellyfinUserID, req.CurrentPassword, req.NewPassword); err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "password change failed: "+err.Error())
		return
	}

	middleware.WriteSuccess(w, map[string]string{"status": "updated"})
}

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
		middleware.WriteError(w, http.StatusInternalServerError, "failed to get devices")
		return
	}

	middleware.WriteSuccess(w, devices)
}

func (h *Handler) JellyfinUpdateParentalRating(w http.ResponseWriter, r *http.Request) {
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

	jfClient := jellyfin.NewClient(cfg.Jellyfin.URL, cfg.Jellyfin.Token)
	if err := jfClient.UpdateParentalRating(user.JellyfinUserID, req.Rating); err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to update rating")
		return
	}

	// Update local record
	database.DB().ExecContext(r.Context(), jellyfinQuery2, req.Rating, user.ID)

	middleware.WriteSuccess(w, map[string]string{"status": "updated"})
}
