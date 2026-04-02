package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/middleware"
	"github.com/user/remna-user-panel/internal/models"
	"github.com/user/remna-user-panel/internal/sdk/bepusdt"
	"github.com/user/remna-user-panel/internal/sdk/ezpay"
	"github.com/user/remna-user-panel/internal/services"
)

const (
	queryComboByUUID = "SELECT uuid, name, squad_uuid, traffic_gb, strategy, cycle, price_rmb FROM combos WHERE uuid = ? AND active = 1"
)

// ─── Credit Handlers ─────────────────────────────────────────────────

// GetCreditBalance returns the authenticated user's current credit balance.
func (h *Handler) GetCreditBalance(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	balance, _ := h.Credit.GetBalance(user.ID)
	middleware.WriteSuccess(w, map[string]interface{}{
		"balance": balance,
		"name":    config.Get().Credit.Name,
	})
}

// CreditSignup awards a random daily credit to the user.
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

// CreditBet lets the user gamble an amount of credits.
func (h *Handler) CreditBet(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)

	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.Amount <= 0 {
		middleware.WriteError(w, http.StatusBadRequest, "amount must be greater than 0")
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

// GetCreditHistory returns paginated credit transaction history.
func (h *Handler) GetCreditHistory(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	logs, err := h.Credit.GetHistory(user.ID, limit, offset)
	if err != nil {
		slog.Error("credit-history: query failed", "user_id", user.ID, "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "failed to get history")
		return
	}
	middleware.WriteSuccess(w, logs)
}

// ─── Payment Handlers ────────────────────────────────────────────────

// CreatePayment creates a generic payment order through the configured
// payment gateway (BEPusdt or EZPay).
func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)

	var req services.CreatePaymentRequest
	if err := middleware.DecodeJSON(r, &req); err != nil {
		middleware.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}
	req.ClientIP = r.RemoteAddr

	resp, err := h.Payment.CreatePayment(r.Context(), user.ID, req)
	if err != nil {
		middleware.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	middleware.WriteSuccess(w, resp)
}

// BEPusdtCallback handles payment-success webhook callbacks from BEPusdt.
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

	if data.Status == 2 { // Status 2 = Payment success
		if err := h.Payment.CompleteOrder(data.OrderID); err != nil {
			slog.Error("bepusdt-callback: complete order failed", "order_id", data.OrderID, "error", err)
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "success")
}

// EZPayCallback handles payment-success webhook callbacks from EZPay.
func (h *Handler) EZPayCallback(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	params := make(map[string]string, len(r.Form))
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
			slog.Error("ezpay-callback: complete order failed", "order_id", orderID, "error", err)
		}
	}

	fmt.Fprint(w, "success")
}

// ─── Combo Purchase ──────────────────────────────────────────────────

// PurchaseCombo creates a payment order for a subscription combo plan.
// It resolves the combo details, calculates the expiry based on the billing
// cycle, and delegates to the payment service.
func (h *Handler) PurchaseCombo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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
	if req.ComboUUID == "" {
		middleware.WriteError(w, http.StatusBadRequest, "combo_uuid is required")
		return
	}

	// Fetch combo details from the database.
	var combo models.Combo
	if err := database.DB().QueryRowContext(ctx, queryComboByUUID, req.ComboUUID).
		Scan(&combo.UUID, &combo.Name, &combo.SquadUUID, &combo.TrafficGB, &combo.Strategy, &combo.Cycle, &combo.PriceRMB); err != nil {
		middleware.WriteError(w, http.StatusNotFound, "combo not found or inactive")
		return
	}

	// Derive a username from the Telegram ID for Remnawave account creation.
	username := fmt.Sprintf("tg_%d", user.TelegramID)

	// Calculate expiry based on the combo's billing cycle.
	now := time.Now()
	var expiry time.Time
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
		expiry = now.AddDate(0, 1, 0) // Default to monthly
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

	payResp, err := h.Payment.CreatePayment(ctx, user.ID, services.CreatePaymentRequest{
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

// ─── Order Handlers ──────────────────────────────────────────────────

// ListOrders returns paginated orders for the authenticated user.
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	orders, err := h.Payment.GetUserOrders(user.ID, limit, offset)
	if err != nil {
		slog.Error("list-orders: query failed", "user_id", user.ID, "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "failed to load orders")
		return
	}
	middleware.WriteSuccess(w, orders)
}

// GetOrder returns a single order by UUID. Non-admin users may only
// access their own orders.
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	orderUUID := chi.URLParam(r, "uuid")

	order, err := h.Payment.GetOrderDetail(orderUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.WriteError(w, http.StatusNotFound, "order not found")
			return
		}
		slog.Error("get-order: query failed", "uuid", orderUUID, "error", err)
		middleware.WriteError(w, http.StatusInternalServerError, "failed to load order")
		return
	}

	if !user.IsAdmin && order.UserID != user.ID {
		middleware.WriteError(w, http.StatusForbidden, "order access denied")
		return
	}

	middleware.WriteSuccess(w, order)
}

// CustomPayment creates a freeform payment order (e.g. custom donation or
// manual renewal) with an optional user-supplied message.
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

	payResp, err := h.Payment.CreatePayment(r.Context(), user.ID, services.CreatePaymentRequest{
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
