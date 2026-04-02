package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/models"
	"github.com/user/remna-user-panel/internal/sdk/bepusdt"
	"github.com/user/remna-user-panel/internal/sdk/ezpay"
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

// CreatePayment creates a new payment order
func (s *PaymentService) CreatePayment(userID int64, req CreatePaymentRequest) (*CreatePaymentResponse, error) {
	cfg := config.Get()

	if req.Amount < 0 {
		return nil, fmt.Errorf("金额不能为负数")
	}

	// Apply TXB discount
	var discountRMB, consumedTXB, finalAmount float64
	var err error

	if req.UseTXB && req.Amount > 0 {
		discountRMB, consumedTXB, finalAmount, err = s.credit.ApplyDiscount(userID, req.Amount, true)
		if err != nil {
			return nil, fmt.Errorf("calculate discount: %w", err)
		}
	} else {
		finalAmount = req.Amount
	}

	// Generate order UUID
	orderUUID := uuid.New().String()

	// Create order record
	_, err = database.DB().Exec(
		`INSERT INTO orders (uuid, user_id, order_type, amount, txb_discount, final_amount, status, payment_method, payment_type, metadata, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, 'pending', ?, ?, ?, ?, ?)`,
		orderUUID, userID, req.OrderType, req.Amount, discountRMB, finalAmount,
		req.PaymentMethod, req.PaymentType, req.Metadata, time.Now(), time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	// If amount is 0, auto-complete
	if finalAmount <= 0 {
		if consumedTXB > 0 {
			s.credit.AddCredit(userID, -consumedTXB, fmt.Sprintf("订单折扣 -%s (订单%s)", fmt.Sprintf("%.2f", consumedTXB), orderUUID[:8]))
		}
		err = s.CompleteOrder(orderUUID)
		if err != nil {
			return nil, err
		}
		return &CreatePaymentResponse{
			OrderUUID:   orderUUID,
			FinalAmount: 0,
			TXBDiscount: discountRMB,
			TXBUsed:     consumedTXB,
			PaymentURL:  "",
		}, nil
	}

	// Deduct TXB if used
	if consumedTXB > 0 {
		_, err = s.credit.AddCredit(userID, -consumedTXB, fmt.Sprintf("订单折扣 -%s (订单%s)", fmt.Sprintf("%.2f", consumedTXB), orderUUID[:8]))
		if err != nil {
			return nil, fmt.Errorf("deduct TXB: %w", err)
		}
	}

	// Create upstream payment
	var paymentURL, tradeID string

	switch req.PaymentMethod {
	case "bepusdt":
		client := bepusdt.NewClient(cfg.BEPusdt.URL, cfg.BEPusdt.Token, cfg.BEPusdt.NotifyURL, cfg.BEPusdt.RedirectURL)
		resp, err := client.CreateTransaction(orderUUID, finalAmount, "订单支付")
		if err != nil {
			return nil, fmt.Errorf("create USDT payment: %w", err)
		}
		paymentURL = resp.Data.PaymentURL
		tradeID = resp.Data.TradeID

		// Save upstream trade ID
		database.DB().Exec("UPDATE orders SET upstream_id = ?, updated_at = ? WHERE uuid = ?",
			tradeID, time.Now(), orderUUID)

	case "ezpay":
		client := ezpay.NewClient(cfg.EZPay.URL, cfg.EZPay.PID, cfg.EZPay.Key, cfg.EZPay.NotifyURL, cfg.EZPay.ReturnURL)
		if req.PaymentType == "" {
			req.PaymentType = "alipay"
		}
		resp, err := client.CreatePayment(orderUUID, req.PaymentType, "订单支付", fmt.Sprintf("%.2f", finalAmount), req.ClientIP)
		if err != nil {
			// Fallback to page redirect
			paymentURL = client.GetPaymentURL(orderUUID, req.PaymentType, "订单支付", fmt.Sprintf("%.2f", finalAmount))
		} else {
			paymentURL = resp.PayURL
			if paymentURL == "" {
				paymentURL = resp.QRCode
			}
			tradeID = resp.TradeNo
		}

		database.DB().Exec("UPDATE orders SET upstream_id = ?, updated_at = ? WHERE uuid = ?",
			tradeID, time.Now(), orderUUID)

	default:
		return nil, fmt.Errorf("unsupported payment method: %s", req.PaymentMethod)
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
	var order models.Order
	err := database.DB().QueryRow(
		"SELECT uuid, user_id, order_type, amount, txb_discount, final_amount, status, metadata FROM orders WHERE uuid = ?",
		orderUUID,
	).Scan(&order.UUID, &order.UserID, &order.OrderType, &order.Amount, &order.TXBDiscount, &order.FinalAmount, &order.Status, &order.Metadata)
	if err != nil {
		return fmt.Errorf("find order: %w", err)
	}

	if order.Status == "paid" {
		return nil // Already processed (idempotent)
	}

	// Mark as paid
	_, err = database.DB().Exec("UPDATE orders SET status = 'paid', updated_at = ? WHERE uuid = ?", time.Now(), orderUUID)
	if err != nil {
		return err
	}

	// Add TXB for payment (only if TXB was not used for discount)
	if order.TXBDiscount == 0 && order.FinalAmount > 0 {
		s.credit.ConvertPaymentToCredit(order.UserID, order.FinalAmount)
	}

	// Trigger fulfillment based on order type
	go s.fulfillOrder(&order)

	return nil
}

// fulfillOrder handles post-payment fulfillment
func (s *PaymentService) fulfillOrder(order *models.Order) {
	switch order.OrderType {
	case "combo":
		log.Printf("[payment] fulfilling combo order %s for user %d", order.UUID, order.UserID)
		// Combo fulfillment handled by combo handler

	case "jellyfin":
		log.Printf("[payment] fulfilling jellyfin order %s for user %d", order.UUID, order.UserID)
		// Jellyfin fulfillment handled by jellyfin handler

	case "renewal":
		log.Printf("[payment] fulfilling renewal order %s for user %d", order.UUID, order.UserID)

	case "traffic_reset":
		log.Printf("[payment] fulfilling traffic reset order %s for user %d", order.UUID, order.UserID)

	default:
		log.Printf("[payment] unknown order type: %s", order.OrderType)
	}
}

// GetOrder gets an order by UUID
func (s *PaymentService) GetOrder(orderUUID string) (*models.Order, error) {
	var order models.Order
	err := database.DB().QueryRow(
		`SELECT uuid, user_id, order_type, amount, txb_discount, final_amount, status, payment_method, payment_type, upstream_id, metadata, created_at, updated_at
		 FROM orders WHERE uuid = ?`, orderUUID,
	).Scan(&order.UUID, &order.UserID, &order.OrderType, &order.Amount, &order.TXBDiscount, &order.FinalAmount,
		&order.Status, &order.PaymentMethod, &order.PaymentType, &order.UpstreamID, &order.Metadata, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetUserOrders gets orders for a user
func (s *PaymentService) GetUserOrders(userID int64, limit, offset int) ([]models.Order, error) {
	rows, err := database.DB().Query(
		`SELECT uuid, user_id, order_type, amount, txb_discount, final_amount, status, payment_method, payment_type, upstream_id, metadata, created_at, updated_at
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
			&o.Status, &o.PaymentMethod, &o.PaymentType, &o.UpstreamID, &o.Metadata, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

// Ensure json import is used
var _ = json.Marshal
