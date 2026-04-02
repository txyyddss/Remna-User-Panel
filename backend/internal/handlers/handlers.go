package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/middleware"
	"github.com/user/remna-user-panel/internal/models"
	"github.com/user/remna-user-panel/internal/sdk/bepusdt"
	"github.com/user/remna-user-panel/internal/sdk/ezpay"
	"github.com/user/remna-user-panel/internal/sdk/jellyfin"
	"github.com/user/remna-user-panel/internal/sdk/remnawave"
	"github.com/user/remna-user-panel/internal/services"
)

// Handler holds all HTTP handler dependencies
type Handler struct {
	Credit  *services.CreditService
	Payment *services.PaymentService
}

// NewHandler creates a new Handler
func NewHandler() *Handler {
	credit := services.NewCreditService()
	payment := services.NewPaymentService(credit)
	return &Handler{
		Credit:  credit,
		Payment: payment,
	}
}

// --- Auth ---

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		middleware.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	// Get subscription info
	var sub *models.Subscription
	var subData models.Subscription
	err := database.DB().QueryRow(
		"SELECT id, user_id, combo_uuid, remnawave_uuid, status, expires_at, created_at FROM subscriptions WHERE user_id = ? AND status = 'active' ORDER BY created_at DESC LIMIT 1",
		user.ID,
	).Scan(&subData.ID, &subData.UserID, &subData.ComboUUID, &subData.RemnawaveUUID, &subData.Status, &subData.ExpiresAt, &subData.CreatedAt)
	if err == nil {
		sub = &subData
	}

	// Get Jellyfin account info
	var jfAccount *models.JellyfinAccount
	var jf models.JellyfinAccount
	err = database.DB().QueryRow(
		"SELECT id, user_id, jellyfin_user_id, username, parental_rating, expires_at FROM jellyfin_accounts WHERE user_id = ?",
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
		},
	})
}

// --- Credit ---

func (h *Handler) GetCreditBalance(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	balance, _ := h.Credit.GetBalance(user.ID)
	middleware.WriteSuccess(w, map[string]interface{}{
		"balance": balance,
		"name":    config.Get().Credit.Name,
	})
}

func (h *Handler) CreditSignup(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	value, newBalance, err := h.Credit.Signup(user.ID)
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"value":       value,
		"new_balance": newBalance,
		"auto_delete": value < 1,
	})
}

func (h *Handler) CreditBet(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)

	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	result, newBalance, err := h.Credit.Bet(user.ID, req.Amount)
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"result":      result,
		"new_balance": newBalance,
	})
}

func (h *Handler) GetCreditHistory(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	logs, err := h.Credit.GetHistory(user.ID, limit, offset)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to get history")
		return
	}
	middleware.WriteSuccess(w, logs)
}

// --- Payment ---

func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)

	var req services.CreatePaymentRequest
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	req.ClientIP = r.RemoteAddr

	resp, err := h.Payment.CreatePayment(user.ID, req)
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	middleware.WriteSuccess(w, resp)
}

func (h *Handler) BEPusdtCallback(w http.ResponseWriter, r *http.Request) {
	var data bepusdt.CallbackData
	if err := middleware.DecodeJSON(r, &data); err != nil {
		http.Error(w, "invalid", http.StatusBadRequest)
		return
	}

	cfg := config.Get()
	client := bepusdt.NewClient(cfg.BEPusdt.URL, cfg.BEPusdt.Token, "", "")
	if !client.VerifyCallback(&data) {
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}

	if data.Status == 2 { // Payment success
		if err := h.Payment.CompleteOrder(data.OrderID); err != nil {
			log.Printf("[bepusdt-callback] complete order error: %v", err)
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}

func (h *Handler) EZPayCallback(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	params := make(map[string]string)
	for k, v := range r.Form {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	cfg := config.Get()
	client := ezpay.NewClient(cfg.EZPay.URL, cfg.EZPay.PID, cfg.EZPay.Key, "", "")
	if !client.VerifyCallback(params) {
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}

	if params["trade_status"] == "TRADE_SUCCESS" {
		orderID := params["out_trade_no"]
		if err := h.Payment.CompleteOrder(orderID); err != nil {
			log.Printf("[ezpay-callback] complete order error: %v", err)
		}
	}

	fmt.Fprint(w, "success")
}

// --- Combos ---

func (h *Handler) ListCombos(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB().Query("SELECT uuid, name, description, squad_uuid, traffic_gb, strategy, cycle, price_rmb, reset_price, active FROM combos WHERE active = 1")
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to list combos")
		return
	}
	defer rows.Close()

	var combos []models.Combo
	for rows.Next() {
		var c models.Combo
		rows.Scan(&c.UUID, &c.Name, &c.Description, &c.SquadUUID, &c.TrafficGB, &c.Strategy, &c.Cycle, &c.PriceRMB, &c.ResetPrice, &c.Active)
		combos = append(combos, c)
	}
	middleware.WriteSuccess(w, combos)
}

func (h *Handler) CreateCombo(w http.ResponseWriter, r *http.Request) {
	var combo models.Combo
	if err := middleware.DecodeJSON(r, &combo); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	combo.UUID = uuid.New().String()
	combo.Active = true
	combo.CreatedAt = time.Now()

	_, err := database.DB().Exec(
		`INSERT INTO combos (uuid, name, description, squad_uuid, traffic_gb, strategy, cycle, price_rmb, reset_price, active, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		combo.UUID, combo.Name, combo.Description, combo.SquadUUID, combo.TrafficGB,
		combo.Strategy, combo.Cycle, combo.PriceRMB, combo.ResetPrice, combo.Active, combo.CreatedAt,
	)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to create combo")
		return
	}

	middleware.WriteJSON(w, http.StatusCreated, combo)
}

func (h *Handler) UpdateCombo(w http.ResponseWriter, r *http.Request) {
	comboUUID := chi.URLParam(r, "uuid")

	var updates map[string]interface{}
	if err := middleware.DecodeJSON(r, &updates); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// Build dynamic update
	allowed := map[string]bool{"name": true, "description": true, "squad_uuid": true, "traffic_gb": true, "strategy": true, "cycle": true, "price_rmb": true, "reset_price": true, "active": true}
	for key := range updates {
		if !allowed[key] {
			delete(updates, key)
		}
	}

	for key, val := range updates {
		database.DB().Exec(fmt.Sprintf("UPDATE combos SET %s = ? WHERE uuid = ?", key), val, comboUUID)
	}

	middleware.WriteSuccess(w, map[string]string{"uuid": comboUUID})
}

// --- Subscription / VPN ---

func (h *Handler) PurchaseCombo(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)

	var req struct {
		ComboUUID     string  `json:"combo_uuid"`
		PaymentMethod string  `json:"payment_method"`
		PaymentType   string  `json:"payment_type"`
		UseTXB        bool    `json:"use_txb"`
		DiscountRMB   float64 `json:"discount_rmb"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// Get combo details
	var combo models.Combo
	err := database.DB().QueryRow("SELECT uuid, name, squad_uuid, traffic_gb, strategy, cycle, price_rmb FROM combos WHERE uuid = ? AND active = 1", req.ComboUUID).
		Scan(&combo.UUID, &combo.Name, &combo.SquadUUID, &combo.TrafficGB, &combo.Strategy, &combo.Cycle, &combo.PriceRMB)
	if err != nil {
		middleware.WriteError(w, http.StatusNotFound, "combo not found")
		return
	}

	// Create username from telegram ID
	username := fmt.Sprintf("tg_%d", user.TelegramID)

	// Calculate expiry based on cycle
	var expiry time.Time
	now := time.Now()
	switch combo.Cycle {
	case "monthly":
		expiry = now.AddDate(0, 1, 0)
	case "quarterly":
		expiry = now.AddDate(0, 3, 0)
	case "semiannual":
		expiry = now.AddDate(0, 6, 0)
	case "annual":
		expiry = now.AddDate(1, 0, 0)
	default:
		expiry = now.AddDate(0, 1, 0)
	}

	metadata, _ := json.Marshal(map[string]interface{}{
		"combo_uuid": combo.UUID,
		"combo_name": combo.Name,
		"squad_uuid": combo.SquadUUID,
		"traffic_gb": combo.TrafficGB,
		"strategy":   combo.Strategy,
		"username":   username,
		"expiry":     expiry.Format(time.RFC3339),
	})

	// Create payment
	payResp, err := h.Payment.CreatePayment(user.ID, services.CreatePaymentRequest{
		OrderType:     "combo",
		Amount:        combo.PriceRMB,
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

// --- Jellyfin ---

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
	if err := database.DB().QueryRow(
		"SELECT expires_at FROM jellyfin_accounts WHERE user_id = ?",
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

func (h *Handler) createJellyfinAccount(user *models.User, expiry time.Time) error {
	cfg := config.Get()
	jfClient := jellyfin.NewClient(cfg.Jellyfin.URL, cfg.Jellyfin.Token)

	username := fmt.Sprintf("tg_%d", user.TelegramID)
	password := uuid.New().String()[:12]

	jfUser, err := jfClient.CreateUser(username, password)
	if err != nil {
		return err
	}

	database.DB().Exec(
		"INSERT INTO jellyfin_accounts (user_id, jellyfin_user_id, username, parental_rating, expires_at) VALUES (?, ?, ?, 0, ?)",
		user.ID, jfUser.ID, username, expiry,
	)
	database.DB().Exec("UPDATE users SET jellyfin_user_id = ? WHERE id = ?", jfUser.ID, user.ID)
	return nil
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
	database.DB().Exec("UPDATE jellyfin_accounts SET parental_rating = ? WHERE user_id = ?", req.Rating, user.ID)

	middleware.WriteSuccess(w, map[string]string{"status": "updated"})
}

// --- Admin ---

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	// Return safe copy (hide sensitive tokens)
	safeCfg := map[string]interface{}{
		"credit":    cfg.Credit,
		"ai":        map[string]interface{}{"enabled": cfg.AI.Enabled, "model": cfg.AI.Model, "message_batch_size": cfg.AI.MessageBatchSize, "credit_min": cfg.AI.CreditMin, "credit_max": cfg.AI.CreditMax, "leaderboard_interval": cfg.AI.LeaderboardInterval},
		"backup":    cfg.Backup,
		"ip_change": cfg.IPChange,
		"jellyfin":  map[string]interface{}{"monthly_price_rmb": cfg.Jellyfin.MonthlyPriceRMB},
	}
	middleware.WriteSuccess(w, safeCfg)
}

func (h *Handler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var updates map[string]interface{}
	if err := middleware.DecodeJSON(r, &updates); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	err := config.Update(func(cfg *config.Config) {
		data, _ := json.Marshal(updates)
		json.Unmarshal(data, cfg)
	})
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to update config")
		return
	}

	middleware.WriteSuccess(w, map[string]string{"status": "updated"})
}

// GetInternalSquads returns internal squads from Remnawave for admin combo creation
func (h *Handler) GetInternalSquads(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
	squads, err := rwClient.GetInternalSquads()
	if err != nil {
		log.Printf("[admin] GetInternalSquads error: %v", err)
		middleware.WriteError(w, http.StatusInternalServerError, "failed to get internal squads: "+err.Error())
		return
	}
	if squads == nil {
		squads = []remnawave.Squad{}
	}
	middleware.WriteSuccess(w, squads)
}

// DeleteCombo soft-deletes a combo by setting active=0
func (h *Handler) DeleteCombo(w http.ResponseWriter, r *http.Request) {
	comboUUID := chi.URLParam(r, "uuid")
	_, err := database.DB().Exec("UPDATE combos SET active = 0 WHERE uuid = ?", comboUUID)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to delete combo")
		return
	}
	middleware.WriteSuccess(w, map[string]string{"status": "deleted"})
}

// AdminListCombos lists all combos including inactive ones for admin
func (h *Handler) AdminListCombos(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB().Query("SELECT uuid, name, description, squad_uuid, traffic_gb, strategy, cycle, price_rmb, reset_price, active FROM combos ORDER BY created_at DESC")
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to list combos")
		return
	}
	defer rows.Close()

	var combos []models.Combo
	for rows.Next() {
		var c models.Combo
		rows.Scan(&c.UUID, &c.Name, &c.Description, &c.SquadUUID, &c.TrafficGB, &c.Strategy, &c.Cycle, &c.PriceRMB, &c.ResetPrice, &c.Active)
		combos = append(combos, c)
	}
	if combos == nil {
		combos = []models.Combo{}
	}
	middleware.WriteSuccess(w, combos)
}

// --- Admin User Management ---

func (h *Handler) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var rows *sql.Rows
	var err error

	if search != "" {
		searchPattern := "%" + search + "%"
		rows, err = database.DB().Query(
			"SELECT id, telegram_id, telegram_name, remnawave_uuid, jellyfin_user_id, credit, is_admin, created_at, updated_at FROM users WHERE telegram_name LIKE ? OR CAST(telegram_id AS TEXT) LIKE ? ORDER BY id DESC LIMIT ? OFFSET ?",
			searchPattern, searchPattern, limit, offset,
		)
	} else {
		rows, err = database.DB().Query(
			"SELECT id, telegram_id, telegram_name, remnawave_uuid, jellyfin_user_id, credit, is_admin, created_at, updated_at FROM users ORDER BY id DESC LIMIT ? OFFSET ?",
			limit, offset,
		)
	}
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.TelegramID, &u.TelegramName, &u.RemnawaveUUID, &u.JellyfinUserID, &u.Credit, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
		users = append(users, u)
	}
	if users == nil {
		users = []models.User{}
	}

	// Get total count with the same filter that powers the current page.
	var total int
	if search != "" {
		searchPattern := "%" + search + "%"
		database.DB().QueryRow(
			"SELECT COUNT(*) FROM users WHERE telegram_name LIKE ? OR CAST(telegram_id AS TEXT) LIKE ?",
			searchPattern, searchPattern,
		).Scan(&total)
	} else {
		database.DB().QueryRow("SELECT COUNT(*) FROM users").Scan(&total)
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"users": users,
		"total": total,
	})
}

func (h *Handler) AdminGetUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var user models.User
	err = database.DB().QueryRow(
		"SELECT id, telegram_id, telegram_name, remnawave_uuid, jellyfin_user_id, credit, is_admin, created_at, updated_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.TelegramID, &user.TelegramName, &user.RemnawaveUUID, &user.JellyfinUserID, &user.Credit, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		middleware.WriteError(w, http.StatusInternalServerError, "failed to load user")
		return
	}

	var subscription map[string]interface{}
	var comboUUID, status string
	var expiresAt time.Time
	if err := database.DB().QueryRow(
		"SELECT combo_uuid, status, expires_at FROM subscriptions WHERE user_id = ? ORDER BY created_at DESC LIMIT 1",
		userID,
	).Scan(&comboUUID, &status, &expiresAt); err == nil {
		subscription = map[string]interface{}{
			"combo_uuid": comboUUID,
			"status":     status,
			"expires_at": expiresAt,
		}
	}

	var jellyfinAccount map[string]interface{}
	var jfUsername string
	var jfRating int
	var jfExpires time.Time
	if err := database.DB().QueryRow(
		"SELECT jellyfin_user_id, username, parental_rating, expires_at FROM jellyfin_accounts WHERE user_id = ?",
		userID,
	).Scan(&user.JellyfinUserID, &jfUsername, &jfRating, &jfExpires); err == nil {
		jellyfinAccount = map[string]interface{}{
			"jellyfin_user_id": user.JellyfinUserID,
			"username":         jfUsername,
			"parental_rating":  jfRating,
			"expires_at":       jfExpires,
		}
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"user":         user,
		"subscription": subscription,
		"jellyfin":     jellyfinAccount,
	})
}

func (h *Handler) AdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req struct {
		Credit         *float64 `json:"credit"`
		RemnawaveUUID  *string  `json:"remnawave_uuid"`
		JellyfinUserID *string  `json:"jellyfin_user_id"`
		IsAdmin        *bool    `json:"is_admin"`
		Subscription   *struct {
			RemnawaveUUID *string `json:"remnawave_uuid"`
			ComboUUID     *string `json:"combo_uuid"`
			Status        *string `json:"status"`
			ExpiresAt     *string `json:"expires_at"`
		} `json:"subscription"`
		Jellyfin *struct {
			JellyfinUserID *string `json:"jellyfin_user_id"`
			Username       *string `json:"username"`
			ParentalRating *int    `json:"parental_rating"`
			ExpiresAt      *string `json:"expires_at"`
		} `json:"jellyfin"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Credit != nil {
		// Set absolute credit value
		var currentCredit float64
		database.DB().QueryRow("SELECT credit FROM users WHERE id = ?", userID).Scan(&currentCredit)
		diff := *req.Credit - currentCredit
		if diff != 0 {
			h.Credit.AddCredit(userID, diff, "admin adjustment")
		}
	}
	if req.RemnawaveUUID != nil {
		database.DB().Exec("UPDATE users SET remnawave_uuid = ?, updated_at = ? WHERE id = ?", *req.RemnawaveUUID, time.Now(), userID)
	}
	if req.JellyfinUserID != nil {
		database.DB().Exec("UPDATE users SET jellyfin_user_id = ?, updated_at = ? WHERE id = ?", *req.JellyfinUserID, time.Now(), userID)
	}
	if req.IsAdmin != nil {
		adminVal := 0
		if *req.IsAdmin {
			adminVal = 1
		}
		database.DB().Exec("UPDATE users SET is_admin = ?, updated_at = ? WHERE id = ?", adminVal, time.Now(), userID)
	}

	currentRemnawaveUUID := ""
	_ = database.DB().QueryRow("SELECT remnawave_uuid FROM users WHERE id = ?", userID).Scan(&currentRemnawaveUUID)
	if req.Subscription != nil {
		if req.Subscription.RemnawaveUUID != nil {
			currentRemnawaveUUID = *req.Subscription.RemnawaveUUID
			database.DB().Exec("UPDATE users SET remnawave_uuid = ?, updated_at = ? WHERE id = ?", currentRemnawaveUUID, time.Now(), userID)
		}

		var expiresAt time.Time
		if req.Subscription.ExpiresAt != nil && *req.Subscription.ExpiresAt != "" {
			if parsed, err := time.Parse(time.RFC3339, *req.Subscription.ExpiresAt); err == nil {
				expiresAt = parsed
			}
		}

		var existingSubID int64
		err := database.DB().QueryRow(
			"SELECT id FROM subscriptions WHERE user_id = ? ORDER BY created_at DESC LIMIT 1",
			userID,
		).Scan(&existingSubID)
		if err == nil {
			if req.Subscription.ComboUUID != nil {
				database.DB().Exec("UPDATE subscriptions SET combo_uuid = ?, updated_at = ? WHERE id = ?", *req.Subscription.ComboUUID, time.Now(), existingSubID)
			}
			if req.Subscription.Status != nil {
				database.DB().Exec("UPDATE subscriptions SET status = ?, updated_at = ? WHERE id = ?", *req.Subscription.Status, time.Now(), existingSubID)
			}
			if !expiresAt.IsZero() {
				database.DB().Exec("UPDATE subscriptions SET expires_at = ?, updated_at = ? WHERE id = ?", expiresAt, time.Now(), existingSubID)
			}
		} else if req.Subscription.ComboUUID != nil && currentRemnawaveUUID != "" && !expiresAt.IsZero() {
			status := "active"
			if req.Subscription.Status != nil && *req.Subscription.Status != "" {
				status = *req.Subscription.Status
			}
			database.DB().Exec(
				"INSERT INTO subscriptions (user_id, combo_uuid, remnawave_uuid, status, expires_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
				userID, *req.Subscription.ComboUUID, currentRemnawaveUUID, status, expiresAt, time.Now(), time.Now(),
			)
		}

		if currentRemnawaveUUID != "" {
			cfg := config.Get()
			rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
			updateReq := remnawave.UpdateUserRequest{UUID: currentRemnawaveUUID}
			if req.Subscription.Status != nil {
				switch *req.Subscription.Status {
				case "active":
					updateReq.Status = "ACTIVE"
				case "disabled":
					updateReq.Status = "DISABLED"
				case "expired":
					updateReq.Status = "EXPIRED"
				}
			}
			if !expiresAt.IsZero() {
				updateReq.ExpireAt = expiresAt.Format(time.RFC3339)
			}
			if req.Subscription.ComboUUID != nil && *req.Subscription.ComboUUID != "" {
				var squadUUID string
				if err := database.DB().QueryRow("SELECT squad_uuid FROM combos WHERE uuid = ?", *req.Subscription.ComboUUID).Scan(&squadUUID); err == nil && squadUUID != "" {
					updateReq.ActiveInternalSquads = []string{squadUUID}
				}
			}
			if updateReq.Status != "" || updateReq.ExpireAt != "" || len(updateReq.ActiveInternalSquads) > 0 {
				if _, err := rwClient.UpdateUser(updateReq); err != nil {
					log.Printf("[admin] failed to update Remnawave user %d: %v", userID, err)
				}
			}
		}
	}

	currentJellyfinUserID := ""
	_ = database.DB().QueryRow("SELECT jellyfin_user_id FROM users WHERE id = ?", userID).Scan(&currentJellyfinUserID)
	if req.Jellyfin != nil {
		if req.Jellyfin.JellyfinUserID != nil {
			currentJellyfinUserID = *req.Jellyfin.JellyfinUserID
			database.DB().Exec("UPDATE users SET jellyfin_user_id = ?, updated_at = ? WHERE id = ?", currentJellyfinUserID, time.Now(), userID)
		}

		var expiresAt time.Time
		if req.Jellyfin.ExpiresAt != nil && *req.Jellyfin.ExpiresAt != "" {
			if parsed, err := time.Parse(time.RFC3339, *req.Jellyfin.ExpiresAt); err == nil {
				expiresAt = parsed
			}
		}

		var existingAccountID int64
		err := database.DB().QueryRow("SELECT id FROM jellyfin_accounts WHERE user_id = ?", userID).Scan(&existingAccountID)
		if err == nil {
			if req.Jellyfin.JellyfinUserID != nil {
				database.DB().Exec("UPDATE jellyfin_accounts SET jellyfin_user_id = ? WHERE id = ?", *req.Jellyfin.JellyfinUserID, existingAccountID)
			}
			if req.Jellyfin.Username != nil {
				database.DB().Exec("UPDATE jellyfin_accounts SET username = ? WHERE id = ?", *req.Jellyfin.Username, existingAccountID)
			}
			if req.Jellyfin.ParentalRating != nil {
				database.DB().Exec("UPDATE jellyfin_accounts SET parental_rating = ? WHERE id = ?", *req.Jellyfin.ParentalRating, existingAccountID)
			}
			if !expiresAt.IsZero() {
				database.DB().Exec("UPDATE jellyfin_accounts SET expires_at = ? WHERE id = ?", expiresAt, existingAccountID)
			}
		} else if currentJellyfinUserID != "" {
			username := ""
			if req.Jellyfin.Username != nil {
				username = *req.Jellyfin.Username
			}
			rating := 0
			if req.Jellyfin.ParentalRating != nil {
				rating = *req.Jellyfin.ParentalRating
			}
			if !expiresAt.IsZero() {
				database.DB().Exec(
					"INSERT INTO jellyfin_accounts (user_id, jellyfin_user_id, username, parental_rating, expires_at, created_at) VALUES (?, ?, ?, ?, ?, ?)",
					userID, currentJellyfinUserID, username, rating, expiresAt, time.Now(),
				)
			}
		}

		if currentJellyfinUserID != "" && req.Jellyfin.ParentalRating != nil {
			cfg := config.Get()
			jfClient := jellyfin.NewClient(cfg.Jellyfin.URL, cfg.Jellyfin.Token)
			if err := jfClient.UpdateParentalRating(currentJellyfinUserID, *req.Jellyfin.ParentalRating); err != nil {
				log.Printf("[admin] failed to update Jellyfin rating for user %d: %v", userID, err)
			}
		}
	}

	middleware.WriteSuccess(w, map[string]string{"status": "updated"})
}

// --- Subscription Binding ---

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
	database.DB().QueryRow("SELECT COUNT(*) FROM users WHERE remnawave_uuid = ? AND id != ?", rwUser.UUID, user.ID).Scan(&existingCount)
	if existingCount > 0 {
		middleware.WriteError(w, http.StatusConflict, "this subscription is already bound to another user")
		return
	}

	// Bind user
	database.DB().Exec("UPDATE users SET remnawave_uuid = ?, updated_at = ? WHERE id = ?", rwUser.UUID, time.Now(), user.ID)

	var comboUUID string
	if len(rwUser.ActiveInternalSquads) > 0 {
		_ = database.DB().QueryRow(
			"SELECT uuid FROM combos WHERE squad_uuid = ? ORDER BY created_at DESC LIMIT 1",
			rwUser.ActiveInternalSquads[0].UUID,
		).Scan(&comboUUID)
	}
	if comboUUID != "" {
		database.DB().Exec(
			"INSERT INTO subscriptions (user_id, combo_uuid, remnawave_uuid, status, expires_at, created_at, updated_at) VALUES (?, ?, ?, 'active', ?, ?, ?)",
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

// --- Orders / Billing ---

func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	orders, err := h.Payment.GetUserOrders(user.ID, limit, offset)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to load orders")
		return
	}
	middleware.WriteSuccess(w, orders)
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	orderUUID := chi.URLParam(r, "uuid")

	order, err := h.Payment.GetOrderDetail(orderUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.WriteError(w, http.StatusNotFound, "order not found")
			return
		}
		middleware.WriteError(w, http.StatusInternalServerError, "failed to load order")
		return
	}

	if !user.IsAdmin && order.UserID != user.ID {
		middleware.WriteError(w, http.StatusForbidden, "order access denied")
		return
	}

	middleware.WriteSuccess(w, order)
}

func (h *Handler) AdminListOrders(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	orders, total, err := h.Payment.GetAdminOrders(services.OrderFilters{
		Search:        r.URL.Query().Get("search"),
		Status:        r.URL.Query().Get("status"),
		ServiceStatus: r.URL.Query().Get("service_status"),
		OrderType:     r.URL.Query().Get("order_type"),
		DateFrom:      r.URL.Query().Get("date_from"),
		DateTo:        r.URL.Query().Get("date_to"),
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "failed to load orders")
		return
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"orders": orders,
		"total":  total,
	})
}

func (h *Handler) AdminUpdateOrder(w http.ResponseWriter, r *http.Request) {
	admin := middleware.GetUser(r)
	orderUUID := chi.URLParam(r, "uuid")

	var updates services.AdminOrderUpdate
	if err := middleware.DecodeJSON(r, &updates); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	order, err := h.Payment.UpdateOrderByAdmin(orderUUID, admin.ID, updates)
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	middleware.WriteSuccess(w, order)
}

func (h *Handler) AdminOrderAction(w http.ResponseWriter, r *http.Request) {
	admin := middleware.GetUser(r)
	orderUUID := chi.URLParam(r, "uuid")
	action := chi.URLParam(r, "action")

	var (
		order *models.OrderDetail
		err   error
	)

	switch action {
	case "apply-credit":
		order, err = h.Payment.ApplyCustomOrderCredit(orderUUID, admin.ID)
	case "resend-notice":
		order, err = h.Payment.ResendCustomOrderNotification(orderUUID, admin.ID)
	case "refund":
		order, err = h.Payment.RefundOrder(orderUUID, admin.ID)
	case "cancel":
		updateStatus := "cancelled"
		updateService := "cancelled"
		order, err = h.Payment.UpdateOrderByAdmin(orderUUID, admin.ID, services.AdminOrderUpdate{
			Status:        &updateStatus,
			ServiceStatus: &updateService,
		})
	default:
		middleware.WriteError(w, http.StatusBadRequest, "unknown action")
		return
	}

	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	middleware.WriteSuccess(w, order)
}

// --- Custom Payment ---

func (h *Handler) CustomPayment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)

	var req struct {
		Amount        float64 `json:"amount"`
		Message       string  `json:"message"`
		PaymentMethod string  `json:"payment_method"`
		PaymentType   string  `json:"payment_type"`
		UseTXB        bool    `json:"use_txb"`
		DiscountRMB   float64 `json:"discount_rmb"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Amount <= 0 {
		middleware.WriteError(w, http.StatusBadRequest, "amount must be greater than 0")
		return
	}

	metadata, _ := json.Marshal(map[string]interface{}{
		"message": req.Message,
	})

	payResp, err := h.Payment.CreatePayment(user.ID, services.CreatePaymentRequest{
		OrderType:     "custom",
		Amount:        req.Amount,
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
