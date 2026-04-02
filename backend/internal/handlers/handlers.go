package handlers

import (
	"bytes"
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
	cfg := config.Get()

	var req struct {
		ComboUUID     string `json:"combo_uuid"`
		PaymentMethod string `json:"payment_method"`
		PaymentType   string `json:"payment_type"`
		UseTXB        bool   `json:"use_txb"`
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
		Metadata:      string(metadata),
		ClientIP:      r.RemoteAddr,
	})
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// If payment amount is 0 (fully discounted), create remnawave user immediately
	if payResp.FinalAmount <= 0 {
		rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)
		rwUser, err := rwClient.CreateUser(remnawave.CreateUserRequest{
			Username:             username,
			Status:               "ACTIVE",
			TrafficLimitBytes:    combo.TrafficGB * 1024 * 1024 * 1024,
			TrafficLimitStrategy: combo.Strategy,
			ExpireAt:             expiry.Format(time.RFC3339),
			TelegramID:           user.TelegramID,
			ActiveInternalSquads: []string{combo.SquadUUID},
		})
		if err != nil {
			middleware.WriteError(w, http.StatusInternalServerError, "failed to create VPN account: "+err.Error())
			return
		}

		// Save subscription
		database.DB().Exec(
			"INSERT INTO subscriptions (user_id, combo_uuid, remnawave_uuid, status, expires_at) VALUES (?, ?, ?, 'active', ?)",
			user.ID, combo.UUID, rwUser.UUID, expiry,
		)
		database.DB().Exec("UPDATE users SET remnawave_uuid = ? WHERE id = ?", rwUser.UUID, user.ID)
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

	// Check if already has account
	var existing int
	database.DB().QueryRow("SELECT COUNT(*) FROM jellyfin_accounts WHERE user_id = ?", user.ID).Scan(&existing)
	if existing > 0 {
		middleware.WriteError(w, http.StatusConflict, "already have a Jellyfin account")
		return
	}

	var req struct {
		Months        int    `json:"months"`
		PaymentMethod string `json:"payment_method"`
		PaymentType   string `json:"payment_type"`
		UseTXB        bool   `json:"use_txb"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.Months <= 0 {
		req.Months = 1
	}

	amount := cfg.Jellyfin.MonthlyPriceRMB * float64(req.Months)
	expiry := time.Now().AddDate(0, req.Months, 0)

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
		Metadata:      string(metadata),
		ClientIP:      r.RemoteAddr,
	})
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// If fully discounted, create immediately
	if payResp.FinalAmount <= 0 {
		h.createJellyfinAccount(user, expiry)
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

	jfClient := jellyfin.NewClient(cfg.Jellyfin.URL, cfg.Jellyfin.Token)
	if err := jfClient.AuthorizeQuickConnect(req.Code); err != nil {
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

	// Get total count
	var total int
	database.DB().QueryRow("SELECT COUNT(*) FROM users").Scan(&total)

	middleware.WriteSuccess(w, map[string]interface{}{
		"users": users,
		"total": total,
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
			h.Credit.AddCredit(userID, diff, "管理员调整")
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

	middleware.WriteSuccess(w, map[string]string{"status": "updated"})
}

// --- Subscription Binding ---

func (h *Handler) BindSubscription(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()

	if user.RemnawaveUUID != "" {
		middleware.WriteError(w, http.StatusConflict, "已有绑定的订阅")
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
		middleware.WriteError(w, http.StatusBadRequest, "请提供订阅链接")
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
		middleware.WriteError(w, http.StatusNotFound, "未找到该订阅: "+err.Error())
		return
	}

	// Verify the Remnawave user isn't already bound to another panel user
	var existingCount int
	database.DB().QueryRow("SELECT COUNT(*) FROM users WHERE remnawave_uuid = ? AND id != ?", rwUser.UUID, user.ID).Scan(&existingCount)
	if existingCount > 0 {
		middleware.WriteError(w, http.StatusConflict, "该订阅已被其他用户绑定")
		return
	}

	// Bind user
	database.DB().Exec("UPDATE users SET remnawave_uuid = ?, updated_at = ? WHERE id = ?", rwUser.UUID, time.Now(), user.ID)

	// Create subscription record
	database.DB().Exec(
		"INSERT INTO subscriptions (user_id, combo_uuid, remnawave_uuid, status, expires_at) VALUES (?, 'bound', ?, 'active', ?)",
		user.ID, rwUser.UUID, rwUser.ExpireAt,
	)

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

// --- Custom Payment ---

func (h *Handler) CustomPayment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	cfg := config.Get()

	var req struct {
		Amount  float64 `json:"amount"`
		Message string  `json:"message"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Amount <= 0 {
		middleware.WriteError(w, http.StatusBadRequest, "金额必须大于0")
		return
	}

	// Create a pending order of type "custom"
	orderUUID := uuid.New().String()
	metadata, _ := json.Marshal(map[string]interface{}{
		"message": req.Message,
	})

	_, err := database.DB().Exec(
		`INSERT INTO orders (uuid, user_id, order_type, amount, txb_discount, final_amount, status, payment_method, payment_type, metadata, created_at, updated_at)
		 VALUES (?, ?, 'custom', ?, 0, ?, 'pending', 'manual', 'manual', ?, ?, ?)`,
		orderUUID, user.ID, req.Amount, req.Amount, string(metadata), time.Now(), time.Now(),
	)
	if err != nil {
		middleware.WriteError(w, http.StatusInternalServerError, "创建订单失败")
		return
	}

	// Send Telegram notification to all admins
	notifyText := fmt.Sprintf("💰 新的自定义充值请求\n\n用户: %s (ID: %d)\n金额: ¥%.2f\n订单: %s\n备注: %s\n\n请管理员手动处理",
		user.TelegramName, user.TelegramID, req.Amount, orderUUID[:8], req.Message)

	for _, adminID := range cfg.Telegram.AdminIDs {
		go sendTelegramMessage(cfg.Telegram.BotToken, adminID, notifyText)
	}

	middleware.WriteSuccess(w, map[string]interface{}{
		"order_uuid": orderUUID,
		"amount":     req.Amount,
		"status":     "pending",
		"message":    "已提交，等待管理员处理",
	})
}

// sendTelegramMessage sends a message via Telegram Bot API directly
func sendTelegramMessage(botToken string, chatID int64, text string) {
	if botToken == "" {
		return
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	body, _ := json.Marshal(map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	})
	http.Post(url, "application/json", bytes.NewReader(body))
}
