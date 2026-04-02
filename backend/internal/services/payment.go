package services

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/models"
	"github.com/user/remna-user-panel/internal/sdk/bepusdt"
	"github.com/user/remna-user-panel/internal/sdk/ezpay"
	"github.com/user/remna-user-panel/internal/sdk/jellyfin"
	"github.com/user/remna-user-panel/internal/sdk/remnawave"
)

// PaymentService handles payment creation and processing
type PaymentService struct {
	credit *CreditService
}

// NewPaymentService creates a new PaymentService
func NewPaymentService(credit *CreditService) *PaymentService {
	return &PaymentService{credit: credit}
}

// CreatePaymentRequest is the request to create a payment
type CreatePaymentRequest struct {
	OrderType     string  `json:"order_type"`     // combo, jellyfin, credit, renewal, traffic_reset
	Amount        float64 `json:"amount"`         // Amount in RMB
	PaymentMethod string  `json:"payment_method"` // bepusdt, ezpay
	PaymentType   string  `json:"payment_type"`   // alipay, wxpay, usdt (for ezpay)
	UseTXB        bool    `json:"use_txb"`        // Whether to use TXB discount
	DiscountRMB   float64 `json:"discount_rmb"`   // Requested credit discount in RMB
	Metadata      string  `json:"metadata"`       // JSON metadata (combo_uuid, etc.)
	ClientIP      string  `json:"-"`              // Set by handler
}

// CreatePaymentResponse is the response after creating payment
type CreatePaymentResponse struct {
	OrderUUID   string  `json:"order_uuid"`
	FinalAmount float64 `json:"final_amount"`
	TXBDiscount float64 `json:"txb_discount"`
	TXBUsed     float64 `json:"txb_used"`
	PaymentURL  string  `json:"payment_url"`
	TradeID     string  `json:"trade_id,omitempty"`
}

type OrderFilters struct {
	UserID        *int64
	Search        string
	Status        string
	ServiceStatus string
	OrderType     string
	DateFrom      string
	DateTo        string
	Limit         int
	Offset        int
}

type AdminOrderUpdate struct {
	Status        *string  `json:"status"`
	ServiceStatus *string  `json:"service_status"`
	Amount        *float64 `json:"amount"`
	FinalAmount   *float64 `json:"final_amount"`
	PaymentMethod *string  `json:"payment_method"`
	PaymentType   *string  `json:"payment_type"`
	UpstreamID    *string  `json:"upstream_id"`
	AdminNote     *string  `json:"admin_note"`
}

type comboOrderMetadata struct {
	ComboUUID string `json:"combo_uuid"`
	ComboName string `json:"combo_name"`
	SquadUUID string `json:"squad_uuid"`
	TrafficGB int64  `json:"traffic_gb"`
	Strategy  string `json:"strategy"`
	Username  string `json:"username"`
	Expiry    string `json:"expiry"`
}

type jellyfinOrderMetadata struct {
	Months int    `json:"months"`
	Expiry string `json:"expiry"`
}

type customOrderMetadata struct {
	Message string `json:"message"`
}

type paymentUser struct {
	ID             int64
	TelegramID     int64
	TelegramName   string
	RemnawaveUUID  string
	JellyfinUserID string
}

// CreatePayment creates a new payment order
func (s *PaymentService) CreatePayment(userID int64, req CreatePaymentRequest) (*CreatePaymentResponse, error) {
	if req.Amount < 0 {
		return nil, fmt.Errorf("amount cannot be negative")
	}

	discountRMB, consumedTXB, finalAmount, err := s.credit.ApplyDiscount(userID, req.Amount, req.UseTXB, req.DiscountRMB)
	if err != nil {
		return nil, fmt.Errorf("calculate discount: %w", err)
	}

	paymentMethod, paymentType, err := normalizePaymentRequest(finalAmount, req.PaymentMethod, req.PaymentType)
	if err != nil {
		return nil, err
	}

	orderUUID := uuid.New().String()
	order := &models.Order{
		UUID:          orderUUID,
		UserID:        userID,
		OrderType:     req.OrderType,
		Amount:        req.Amount,
		TXBDiscount:   discountRMB,
		FinalAmount:   finalAmount,
		Status:        "pending",
		ServiceStatus: "pending",
		PaymentMethod: paymentMethod,
		PaymentType:   paymentType,
		Metadata:      req.Metadata,
	}

	_, err = database.DB().Exec(
		`INSERT INTO orders (uuid, user_id, order_type, amount, txb_discount, final_amount, status, service_status, payment_method, payment_type, metadata, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, 'pending', ?, ?, ?, ?, ?, ?)`,
		order.UUID, order.UserID, order.OrderType, order.Amount, order.TXBDiscount, order.FinalAmount,
		order.ServiceStatus, order.PaymentMethod, order.PaymentType, order.Metadata, time.Now(), time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	s.recordOrderEvent(order.UUID, nil, "created", "Order created", map[string]interface{}{
		"order_type":     order.OrderType,
		"amount":         order.Amount,
		"final_amount":   order.FinalAmount,
		"txb_discount":   order.TXBDiscount,
		"payment_method": order.PaymentMethod,
		"payment_type":   order.PaymentType,
	})

	if consumedTXB > 0 {
		_, err = s.credit.ConsumeCredit(userID, consumedTXB, fmt.Sprintf("order discount -%.2f (order %s)", consumedTXB, orderUUID[:8]))
		if err != nil {
			s.cancelOrder(orderUUID)
			return nil, fmt.Errorf("deduct TXB: %w", err)
		}
		s.recordOrderEvent(order.UUID, &userID, "discount_applied", "Credit discount reserved", map[string]interface{}{
			"discount_rmb": discountRMB,
			"txb_used":     consumedTXB,
		})
	}

	if finalAmount <= 0 {
		if err := s.CompleteOrder(orderUUID); err != nil {
			return nil, s.handleCreationFailure(orderUUID, userID, consumedTXB, err)
		}
		return &CreatePaymentResponse{
			OrderUUID:   orderUUID,
			FinalAmount: 0,
			TXBDiscount: discountRMB,
			TXBUsed:     consumedTXB,
			PaymentURL:  "",
		}, nil
	}

	paymentURL, tradeID, err := s.createUpstreamPayment(orderUUID, finalAmount, paymentMethod, paymentType, req.ClientIP)
	if err != nil {
		return nil, s.handleCreationFailure(orderUUID, userID, consumedTXB, err)
	}

	if _, err := database.DB().Exec(
		"UPDATE orders SET upstream_id = ?, updated_at = ? WHERE uuid = ?",
		tradeID, time.Now(), orderUUID,
	); err != nil {
		return nil, s.handleCreationFailure(orderUUID, userID, consumedTXB, fmt.Errorf("save upstream trade id: %w", err))
	}

	return &CreatePaymentResponse{
		OrderUUID:   orderUUID,
		FinalAmount: finalAmount,
		TXBDiscount: discountRMB,
		TXBUsed:     consumedTXB,
		PaymentURL:  paymentURL,
		TradeID:     tradeID,
	}, nil
}

// CompleteOrder marks an order as paid and triggers fulfillment
func (s *PaymentService) CompleteOrder(orderUUID string) error {
	order, err := s.GetOrder(orderUUID)
	if err != nil {
		return fmt.Errorf("find order: %w", err)
	}

	if order.Status == "paid" {
		return nil
	}

	result, err := database.DB().Exec(
		"UPDATE orders SET status = 'processing', updated_at = ? WHERE uuid = ? AND status = 'pending'",
		time.Now(), orderUUID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		refreshed, refreshErr := s.GetOrder(orderUUID)
		if refreshErr != nil {
			return refreshErr
		}
		if refreshed.Status == "paid" {
			return nil
		}
		return fmt.Errorf("order cannot be completed from status %s", refreshed.Status)
	}

	s.recordOrderEvent(orderUUID, nil, "payment_received", "Payment callback accepted", nil)

	serviceStatus, err := s.fulfillOrder(order)
	if err != nil {
		if _, rollbackErr := database.DB().Exec(
			"UPDATE orders SET status = 'pending', service_status = 'failed', updated_at = ? WHERE uuid = ? AND status = 'processing'",
			time.Now(), orderUUID,
		); rollbackErr != nil {
			log.Printf("[payment] failed to roll back order %s after fulfillment error: %v", orderUUID, rollbackErr)
		}
		s.recordOrderEvent(orderUUID, nil, "fulfillment_failed", err.Error(), nil)
		return err
	}

	if order.OrderType != "custom" && order.TXBDiscount == 0 && order.FinalAmount > 0 {
		if _, err := s.credit.ConvertPaymentToCredit(order.UserID, order.FinalAmount); err != nil {
			log.Printf("[payment] failed to grant credit bonus for order %s: %v", order.UUID, err)
		}
	}

	_, err = database.DB().Exec(
		"UPDATE orders SET status = 'paid', service_status = ?, paid_at = ?, updated_at = ? WHERE uuid = ? AND status = 'processing'",
		serviceStatus, time.Now(), time.Now(), orderUUID,
	)
	if err == nil {
		s.recordOrderEvent(orderUUID, nil, "fulfilled", "Order fulfillment completed", map[string]interface{}{
			"service_status": serviceStatus,
		})
	}
	return err
}

func normalizePaymentRequest(finalAmount float64, paymentMethod, paymentType string) (string, string, error) {
	if finalAmount <= 0 {
		return "", "", nil
	}

	switch paymentMethod {
	case "bepusdt":
		return paymentMethod, paymentType, nil
	case "ezpay":
		if paymentType == "" {
			paymentType = "alipay"
		}
		return paymentMethod, paymentType, nil
	default:
		return "", "", fmt.Errorf("unsupported payment method: %s", paymentMethod)
	}
}

func (s *PaymentService) createUpstreamPayment(orderUUID string, finalAmount float64, paymentMethod, paymentType, clientIP string) (string, string, error) {
	cfg := config.Get()

	switch paymentMethod {
	case "bepusdt":
		client := bepusdt.NewClient(cfg.BEPusdt.URL, cfg.BEPusdt.Token, cfg.BEPusdt.NotifyURL, cfg.BEPusdt.RedirectURL)
		resp, err := client.CreateOrder(orderUUID, finalAmount, "order payment", "USDT")
		if err == nil && resp.Data.PaymentURL != "" {
			return resp.Data.PaymentURL, resp.Data.TradeID, nil
		}

		resp, fallbackErr := client.CreateTransaction(orderUUID, finalAmount, "order payment")
		if fallbackErr != nil {
			if err != nil {
				return "", "", fmt.Errorf("create USDT payment: %w", err)
			}
			return "", "", fmt.Errorf("create USDT payment: %w", fallbackErr)
		}
		return resp.Data.PaymentURL, resp.Data.TradeID, nil

	case "ezpay":
		client := ezpay.NewClient(cfg.EZPay.URL, cfg.EZPay.PID, cfg.EZPay.Key, cfg.EZPay.NotifyURL, cfg.EZPay.ReturnURL)
		resp, err := client.CreatePayment(orderUUID, paymentType, "order payment", fmt.Sprintf("%.2f", finalAmount), clientIP)
		if err != nil {
			// Fallback to page redirect when the API endpoint is unavailable.
			return client.GetPaymentURL(orderUUID, paymentType, "order payment", fmt.Sprintf("%.2f", finalAmount)), "", nil
		}

		paymentURL := resp.PayURL
		if paymentURL == "" {
			paymentURL = resp.QRCode
		}
		return paymentURL, resp.TradeNo, nil

	default:
		return "", "", fmt.Errorf("unsupported payment method: %s", paymentMethod)
	}
}

func (s *PaymentService) handleCreationFailure(orderUUID string, userID int64, consumedTXB float64, cause error) error {
	if consumedTXB > 0 {
		if _, refundErr := s.credit.AddCredit(userID, consumedTXB, fmt.Sprintf("order failed refund +%.2f (order %s)", consumedTXB, orderUUID[:8])); refundErr != nil {
			log.Printf("[payment] failed to refund TXB for order %s: %v", orderUUID, refundErr)
			s.cancelOrder(orderUUID)
			return fmt.Errorf("%w; refund failed: %v", cause, refundErr)
		}
	}

	s.cancelOrder(orderUUID)
	return cause
}

func (s *PaymentService) cancelOrder(orderUUID string) {
	if _, err := database.DB().Exec(
		"UPDATE orders SET status = 'cancelled', service_status = 'cancelled', updated_at = ? WHERE uuid = ? AND status != 'paid'",
		time.Now(), orderUUID,
	); err != nil {
		log.Printf("[payment] failed to cancel order %s: %v", orderUUID, err)
		return
	}
	s.recordOrderEvent(orderUUID, nil, "cancelled", "Order cancelled", nil)
}

// fulfillOrder handles post-payment fulfillment
func (s *PaymentService) fulfillOrder(order *models.Order) (string, error) {
	switch order.OrderType {
	case "combo":
		return s.fulfillComboOrder(order)
	case "jellyfin":
		return s.fulfillJellyfinOrder(order)
	case "custom":
		return s.fulfillCustomOrder(order)
	case "renewal", "traffic_reset":
		return "fulfilled", nil
	default:
		return "", fmt.Errorf("unknown order type: %s", order.OrderType)
	}
}

func (s *PaymentService) fulfillComboOrder(order *models.Order) (string, error) {
	var metadata comboOrderMetadata
	if err := json.Unmarshal([]byte(order.Metadata), &metadata); err != nil {
		return "", fmt.Errorf("parse combo metadata: %w", err)
	}
	if metadata.ComboUUID == "" || metadata.SquadUUID == "" {
		return "", fmt.Errorf("invalid combo metadata")
	}

	expiry, err := time.Parse(time.RFC3339, metadata.Expiry)
	if err != nil {
		return "", fmt.Errorf("parse combo expiry: %w", err)
	}

	user, err := s.loadUser(order.UserID)
	if err != nil {
		return "", err
	}

	cfg := config.Get()
	rwClient := remnawave.NewClient(cfg.Remnawave.URL, cfg.Remnawave.Token)

	username := metadata.Username
	if username == "" {
		username = fmt.Sprintf("tg_%d", user.TelegramID)
	}

	trafficLimitBytes := metadata.TrafficGB * 1024 * 1024 * 1024
	remnawaveUUID := user.RemnawaveUUID

	if remnawaveUUID != "" {
		_, err = rwClient.UpdateUser(remnawave.UpdateUserRequest{
			UUID:                 remnawaveUUID,
			Username:             username,
			Status:               "ACTIVE",
			TrafficLimitBytes:    trafficLimitBytes,
			TrafficLimitStrategy: metadata.Strategy,
			ExpireAt:             expiry.Format(time.RFC3339),
			TelegramID:           user.TelegramID,
			ActiveInternalSquads: []string{metadata.SquadUUID},
		})
		if err != nil {
			return "", fmt.Errorf("update Remnawave user: %w", err)
		}
	} else {
		rwUser, err := rwClient.CreateUser(remnawave.CreateUserRequest{
			Username:             username,
			Status:               "ACTIVE",
			TrafficLimitBytes:    trafficLimitBytes,
			TrafficLimitStrategy: metadata.Strategy,
			ExpireAt:             expiry.Format(time.RFC3339),
			TelegramID:           user.TelegramID,
			ActiveInternalSquads: []string{metadata.SquadUUID},
		})
		if err != nil {
			return "", fmt.Errorf("create Remnawave user: %w", err)
		}
		remnawaveUUID = rwUser.UUID
	}

	tx, err := database.DB().Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		"UPDATE users SET remnawave_uuid = ?, updated_at = ? WHERE id = ?",
		remnawaveUUID, time.Now(), user.ID,
	); err != nil {
		return "", err
	}

	if _, err := tx.Exec(
		"UPDATE subscriptions SET status = 'expired', updated_at = ? WHERE user_id = ? AND status = 'active'",
		time.Now(), user.ID,
	); err != nil {
		return "", err
	}

	if _, err := tx.Exec(
		"INSERT INTO subscriptions (user_id, combo_uuid, remnawave_uuid, status, expires_at, created_at, updated_at) VALUES (?, ?, ?, 'active', ?, ?, ?)",
		user.ID, metadata.ComboUUID, remnawaveUUID, expiry, time.Now(), time.Now(),
	); err != nil {
		return "", err
	}

	return "fulfilled", tx.Commit()
}

func (s *PaymentService) fulfillJellyfinOrder(order *models.Order) (string, error) {
	var metadata jellyfinOrderMetadata
	if err := json.Unmarshal([]byte(order.Metadata), &metadata); err != nil {
		return "", fmt.Errorf("parse jellyfin metadata: %w", err)
	}

	expiry, err := time.Parse(time.RFC3339, metadata.Expiry)
	if err != nil {
		return "", fmt.Errorf("parse jellyfin expiry: %w", err)
	}

	user, err := s.loadUser(order.UserID)
	if err != nil {
		return "", err
	}

	var existingAccountID int64
	var existingJellyfinUserID string
	err = database.DB().QueryRow(
		"SELECT id, jellyfin_user_id FROM jellyfin_accounts WHERE user_id = ?",
		user.ID,
	).Scan(&existingAccountID, &existingJellyfinUserID)
	if err == nil {
		_, err = database.DB().Exec(
			"UPDATE jellyfin_accounts SET expires_at = ? WHERE id = ?",
			expiry, existingAccountID,
		)
		if err != nil {
			return "", err
		}

		if _, err := database.DB().Exec(
			"UPDATE users SET jellyfin_user_id = ?, updated_at = ? WHERE id = ?",
			existingJellyfinUserID, time.Now(), user.ID,
		); err != nil {
			return "", err
		}
		return "fulfilled", nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("load jellyfin account: %w", err)
	}

	cfg := config.Get()
	jfClient := jellyfin.NewClient(cfg.Jellyfin.URL, cfg.Jellyfin.Token)

	username := fmt.Sprintf("tg_%d", user.TelegramID)
	password := uuid.New().String()[:12]

	jfUser, err := jfClient.CreateUser(username, password)
	if err != nil {
		return "", fmt.Errorf("create Jellyfin user: %w", err)
	}

	tx, err := database.DB().Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		"INSERT INTO jellyfin_accounts (user_id, jellyfin_user_id, username, parental_rating, expires_at, created_at) VALUES (?, ?, ?, 0, ?, ?)",
		user.ID, jfUser.ID, username, expiry, time.Now(),
	); err != nil {
		return "", err
	}

	if _, err := tx.Exec(
		"UPDATE users SET jellyfin_user_id = ?, updated_at = ? WHERE id = ?",
		jfUser.ID, time.Now(), user.ID,
	); err != nil {
		return "", err
	}

	return "fulfilled", tx.Commit()
}

func (s *PaymentService) fulfillCustomOrder(order *models.Order) (string, error) {
	var metadata customOrderMetadata
	if err := json.Unmarshal([]byte(order.Metadata), &metadata); err != nil {
		return "", fmt.Errorf("parse custom metadata: %w", err)
	}

	user, err := s.loadUser(order.UserID)
	if err != nil {
		return "", err
	}

	cfg := config.Get()
	notifyText := fmt.Sprintf(
		"💰 New custom top-up request\n\nUser: %s (ID: %d)\nAmount: ¥%.2f\nOrder: %s\nNote: %s\n\nPlease process manually",
		user.TelegramName,
		user.TelegramID,
		order.FinalAmount,
		order.UUID[:8],
		metadata.Message,
	)

	for _, adminID := range cfg.Telegram.AdminIDs {
		go sendTelegramPaymentMessage(cfg.Telegram.BotToken, adminID, notifyText)
	}

	return "waiting_admin", nil
}

func (s *PaymentService) loadUser(userID int64) (*paymentUser, error) {
	var user paymentUser
	err := database.DB().QueryRow(
		"SELECT id, telegram_id, telegram_name, remnawave_uuid, jellyfin_user_id FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.TelegramID, &user.TelegramName, &user.RemnawaveUUID, &user.JellyfinUserID)
	if err != nil {
		return nil, fmt.Errorf("load user: %w", err)
	}
	return &user, nil
}

func (s *PaymentService) recordOrderEvent(orderUUID string, actorUserID *int64, eventType, message string, payload interface{}) {
	rawPayload := "{}"
	if payload != nil {
		if data, err := json.Marshal(payload); err == nil {
			rawPayload = string(data)
		}
	}

	if _, err := database.DB().Exec(
		"INSERT INTO order_events (order_uuid, actor_user_id, event_type, message, payload, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		orderUUID, actorUserID, eventType, message, rawPayload, time.Now(),
	); err != nil {
		log.Printf("[payment] failed to record order event for %s: %v", orderUUID, err)
	}
}

func (s *PaymentService) GetOrderEvents(orderUUID string) ([]models.OrderEvent, error) {
	rows, err := database.DB().Query(
		`SELECT id, order_uuid, actor_user_id, event_type, message, payload, created_at
		 FROM order_events WHERE order_uuid = ? ORDER BY created_at DESC, id DESC`,
		orderUUID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.OrderEvent
	for rows.Next() {
		var event models.OrderEvent
		if err := rows.Scan(&event.ID, &event.OrderUUID, &event.ActorUserID, &event.EventType, &event.Message, &event.Payload, &event.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if events == nil {
		events = []models.OrderEvent{}
	}
	return events, nil
}

// GetOrder gets an order by UUID
func (s *PaymentService) GetOrder(orderUUID string) (*models.Order, error) {
	var order models.Order
	err := database.DB().QueryRow(
		`SELECT uuid, user_id, order_type, amount, txb_discount, final_amount, status, service_status, payment_method, payment_type, upstream_id, metadata, admin_note, paid_at, created_at, updated_at
		 FROM orders WHERE uuid = ?`, orderUUID,
	).Scan(&order.UUID, &order.UserID, &order.OrderType, &order.Amount, &order.TXBDiscount, &order.FinalAmount,
		&order.Status, &order.ServiceStatus, &order.PaymentMethod, &order.PaymentType, &order.UpstreamID, &order.Metadata, &order.AdminNote, &order.PaidAt, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *PaymentService) GetOrderDetail(orderUUID string) (*models.OrderDetail, error) {
	var detail models.OrderDetail
	err := database.DB().QueryRow(
		`SELECT o.uuid, o.user_id, o.order_type, o.amount, o.txb_discount, o.final_amount, o.status, o.service_status,
		        o.payment_method, o.payment_type, o.upstream_id, o.metadata, o.admin_note, o.paid_at, o.created_at, o.updated_at,
		        u.telegram_id, u.telegram_name
		 FROM orders o
		 JOIN users u ON u.id = o.user_id
		 WHERE o.uuid = ?`,
		orderUUID,
	).Scan(&detail.UUID, &detail.UserID, &detail.OrderType, &detail.Amount, &detail.TXBDiscount, &detail.FinalAmount, &detail.Status, &detail.ServiceStatus,
		&detail.PaymentMethod, &detail.PaymentType, &detail.UpstreamID, &detail.Metadata, &detail.AdminNote, &detail.PaidAt, &detail.CreatedAt, &detail.UpdatedAt,
		&detail.UserTelegramID, &detail.UserTelegramName)
	if err != nil {
		return nil, err
	}
	events, err := s.GetOrderEvents(orderUUID)
	if err != nil {
		return nil, err
	}
	detail.Events = events
	return &detail, nil
}

// GetUserOrders gets orders for a user
func (s *PaymentService) GetUserOrders(userID int64, limit, offset int) ([]models.Order, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := database.DB().Query(
		`SELECT uuid, user_id, order_type, amount, txb_discount, final_amount, status, service_status, payment_method, payment_type, upstream_id, metadata, admin_note, paid_at, created_at, updated_at
		 FROM orders WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.UUID, &o.UserID, &o.OrderType, &o.Amount, &o.TXBDiscount, &o.FinalAmount,
			&o.Status, &o.ServiceStatus, &o.PaymentMethod, &o.PaymentType, &o.UpstreamID, &o.Metadata, &o.AdminNote, &o.PaidAt, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []models.Order{}
	}
	return orders, nil
}

func (s *PaymentService) GetAdminOrders(filters OrderFilters) ([]models.OrderDetail, int, error) {
	if filters.Limit <= 0 || filters.Limit > 100 {
		filters.Limit = 20
	}

	baseWhere := []string{"1=1"}
	args := []interface{}{}
	if filters.Search != "" {
		baseWhere = append(baseWhere, "(o.uuid LIKE ? OR o.upstream_id LIKE ? OR u.telegram_name LIKE ? OR CAST(u.telegram_id AS TEXT) LIKE ?)")
		search := "%" + filters.Search + "%"
		args = append(args, search, search, search, search)
	}
	if filters.Status != "" {
		baseWhere = append(baseWhere, "o.status = ?")
		args = append(args, filters.Status)
	}
	if filters.ServiceStatus != "" {
		baseWhere = append(baseWhere, "o.service_status = ?")
		args = append(args, filters.ServiceStatus)
	}
	if filters.OrderType != "" {
		baseWhere = append(baseWhere, "o.order_type = ?")
		args = append(args, filters.OrderType)
	}
	if filters.DateFrom != "" {
		baseWhere = append(baseWhere, "o.created_at >= ?")
		args = append(args, filters.DateFrom)
	}
	if filters.DateTo != "" {
		baseWhere = append(baseWhere, "o.created_at <= ?")
		args = append(args, filters.DateTo)
	}

	whereSQL := strings.Join(baseWhere, " AND ")

	countArgs := append([]interface{}{}, args...)
	var total int
	if err := database.DB().QueryRow(
		`SELECT COUNT(*)
		 FROM orders o
		 JOIN users u ON u.id = o.user_id
		 WHERE `+whereSQL,
		countArgs...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	listArgs := append(args, filters.Limit, filters.Offset)
	rows, err := database.DB().Query(
		`SELECT o.uuid, o.user_id, o.order_type, o.amount, o.txb_discount, o.final_amount, o.status, o.service_status,
		        o.payment_method, o.payment_type, o.upstream_id, o.metadata, o.admin_note, o.paid_at, o.created_at, o.updated_at,
		        u.telegram_id, u.telegram_name
		 FROM orders o
		 JOIN users u ON u.id = o.user_id
		 WHERE `+whereSQL+`
		 ORDER BY o.created_at DESC, o.uuid DESC
		 LIMIT ? OFFSET ?`,
		listArgs...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []models.OrderDetail
	for rows.Next() {
		var order models.OrderDetail
		if err := rows.Scan(&order.UUID, &order.UserID, &order.OrderType, &order.Amount, &order.TXBDiscount, &order.FinalAmount, &order.Status, &order.ServiceStatus,
			&order.PaymentMethod, &order.PaymentType, &order.UpstreamID, &order.Metadata, &order.AdminNote, &order.PaidAt, &order.CreatedAt, &order.UpdatedAt,
			&order.UserTelegramID, &order.UserTelegramName); err != nil {
			return nil, 0, err
		}
		orders = append(orders, order)
	}
	if orders == nil {
		orders = []models.OrderDetail{}
	}
	return orders, total, nil
}

func (s *PaymentService) UpdateOrderByAdmin(orderUUID string, actorUserID int64, updates AdminOrderUpdate) (*models.OrderDetail, error) {
	assignments := []string{}
	args := []interface{}{}
	eventPayload := map[string]interface{}{}

	addField := func(column string, value interface{}, key string) {
		assignments = append(assignments, column+" = ?")
		args = append(args, value)
		eventPayload[key] = value
	}

	if updates.Status != nil {
		addField("status", *updates.Status, "status")
	}
	if updates.ServiceStatus != nil {
		addField("service_status", *updates.ServiceStatus, "service_status")
	}
	if updates.Amount != nil {
		addField("amount", *updates.Amount, "amount")
	}
	if updates.FinalAmount != nil {
		addField("final_amount", *updates.FinalAmount, "final_amount")
	}
	if updates.PaymentMethod != nil {
		addField("payment_method", *updates.PaymentMethod, "payment_method")
	}
	if updates.PaymentType != nil {
		addField("payment_type", *updates.PaymentType, "payment_type")
	}
	if updates.UpstreamID != nil {
		addField("upstream_id", *updates.UpstreamID, "upstream_id")
	}
	if updates.AdminNote != nil {
		addField("admin_note", *updates.AdminNote, "admin_note")
	}

	if len(assignments) == 0 {
		return s.GetOrderDetail(orderUUID)
	}

	assignments = append(assignments, "updated_at = ?")
	args = append(args, time.Now(), orderUUID)
	if _, err := database.DB().Exec("UPDATE orders SET "+strings.Join(assignments, ", ")+" WHERE uuid = ?", args...); err != nil {
		return nil, err
	}

	s.recordOrderEvent(orderUUID, &actorUserID, "admin_updated", "Admin updated order fields", eventPayload)
	return s.GetOrderDetail(orderUUID)
}

func (s *PaymentService) ApplyCustomOrderCredit(orderUUID string, actorUserID int64) (*models.OrderDetail, error) {
	order, err := s.GetOrder(orderUUID)
	if err != nil {
		return nil, err
	}
	if order.OrderType != "custom" {
		return nil, fmt.Errorf("credit can only be applied to custom orders")
	}
	if order.Status != "paid" {
		return nil, fmt.Errorf("order must be paid before credit is applied")
	}
	if order.ServiceStatus == "fulfilled" {
		return nil, fmt.Errorf("credit has already been applied")
	}

	if _, err := s.credit.AddCredit(order.UserID, order.FinalAmount, fmt.Sprintf("custom top-up +%.2f (order %s)", order.FinalAmount, order.UUID[:8])); err != nil {
		return nil, err
	}

	if _, err := database.DB().Exec(
		"UPDATE orders SET service_status = 'fulfilled', updated_at = ? WHERE uuid = ?",
		time.Now(), orderUUID,
	); err != nil {
		return nil, err
	}

	s.recordOrderEvent(orderUUID, &actorUserID, "credit_applied", "Admin applied custom top-up credit", map[string]interface{}{
		"credit_amount": order.FinalAmount,
	})
	return s.GetOrderDetail(orderUUID)
}

func (s *PaymentService) ResendCustomOrderNotification(orderUUID string, actorUserID int64) (*models.OrderDetail, error) {
	order, err := s.GetOrder(orderUUID)
	if err != nil {
		return nil, err
	}
	if order.OrderType != "custom" {
		return nil, fmt.Errorf("notification resend is only available for custom orders")
	}

	if _, err := s.fulfillCustomOrder(order); err != nil {
		return nil, err
	}
	s.recordOrderEvent(orderUUID, &actorUserID, "admin_notified", "Admin notification resent", nil)
	return s.GetOrderDetail(orderUUID)
}

func (s *PaymentService) RefundOrder(orderUUID string, actorUserID int64) (*models.OrderDetail, error) {
	order, err := s.GetOrder(orderUUID)
	if err != nil {
		return nil, err
	}
	if order.Status != "paid" {
		return nil, fmt.Errorf("only paid orders can be refunded")
	}
	if order.OrderType != "custom" {
		return nil, fmt.Errorf("refund is only supported for custom orders in this build")
	}
	if order.ServiceStatus == "fulfilled" {
		if _, err := s.credit.ConsumeCredit(order.UserID, order.FinalAmount, fmt.Sprintf("custom top-up refund -%.2f (order %s)", order.FinalAmount, order.UUID[:8])); err != nil {
			return nil, fmt.Errorf("cannot refund after credit has been spent: %w", err)
		}
	}

	if _, err := database.DB().Exec(
		"UPDATE orders SET status = 'refunded', service_status = 'refunded', updated_at = ? WHERE uuid = ?",
		time.Now(), orderUUID,
	); err != nil {
		return nil, err
	}

	s.recordOrderEvent(orderUUID, &actorUserID, "refunded", "Order refunded by admin", nil)
	return s.GetOrderDetail(orderUUID)
}

func sendTelegramPaymentMessage(botToken string, chatID int64, text string) {
	if botToken == "" {
		return
	}

	body, _ := json.Marshal(map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	})

	_, _ = http.Post(
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken),
		"application/json",
		bytes.NewReader(body),
	)
}
