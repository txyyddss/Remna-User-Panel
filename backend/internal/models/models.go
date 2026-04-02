package models

import "time"

// User represents a panel user linked to Telegram
type User struct {
	ID             int64     `json:"id" db:"id"`
	TelegramID     int64     `json:"telegram_id" db:"telegram_id"`
	TelegramName   string    `json:"telegram_name" db:"telegram_name"`
	RemnawaveUUID  string    `json:"remnawave_uuid,omitempty" db:"remnawave_uuid"`
	JellyfinUserID string    `json:"jellyfin_user_id,omitempty" db:"jellyfin_user_id"`
	Credit         float64   `json:"credit" db:"credit"`
	IsAdmin        bool      `json:"is_admin" db:"is_admin"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Combo represents a VPN subscription plan
type Combo struct {
	UUID        string  `json:"uuid" db:"uuid"`
	Name        string  `json:"name" db:"name"`
	Description string  `json:"description" db:"description"`
	SquadUUID   string  `json:"squad_uuid" db:"squad_uuid"`
	TrafficGB   int64   `json:"traffic_gb" db:"traffic_gb"`
	Strategy    string  `json:"strategy" db:"strategy"` // NO_RESET, DAY, WEEK, MONTH
	Cycle       string  `json:"cycle" db:"cycle"`       // monthly, quarterly, semiannual, annual
	PriceRMB    float64 `json:"price_rmb" db:"price_rmb"`
	ResetPrice  float64 `json:"reset_price" db:"reset_price"`
	Active      bool    `json:"active" db:"active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Subscription represents a user's active VPN subscription
type Subscription struct {
	ID             int64     `json:"id" db:"id"`
	UserID         int64     `json:"user_id" db:"user_id"`
	ComboUUID      string    `json:"combo_uuid" db:"combo_uuid"`
	RemnawaveUUID  string    `json:"remnawave_uuid" db:"remnawave_uuid"`
	Status         string    `json:"status" db:"status"` // active, expired, disabled
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// JellyfinAccount represents a user's Jellyfin account
type JellyfinAccount struct {
	ID              int64     `json:"id" db:"id"`
	UserID          int64     `json:"user_id" db:"user_id"`
	JellyfinUserID  string    `json:"jellyfin_user_id" db:"jellyfin_user_id"`
	Username        string    `json:"username" db:"username"`
	ParentalRating  int       `json:"parental_rating" db:"parental_rating"`
	ExpiresAt       time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// Order represents a payment order
type Order struct {
	UUID          string    `json:"uuid" db:"uuid"`
	UserID        int64     `json:"user_id" db:"user_id"`
	OrderType     string    `json:"order_type" db:"order_type"` // combo, jellyfin, credit, renewal, traffic_reset
	Amount        float64   `json:"amount" db:"amount"`
	TXBDiscount   float64   `json:"txb_discount" db:"txb_discount"`
	FinalAmount   float64   `json:"final_amount" db:"final_amount"`
	Status        string    `json:"status" db:"status"` // pending, paid, cancelled, expired
	PaymentMethod string    `json:"payment_method" db:"payment_method"` // bepusdt, ezpay
	PaymentType   string    `json:"payment_type" db:"payment_type"` // alipay, wxpay, usdt
	UpstreamID    string    `json:"upstream_id,omitempty" db:"upstream_id"`
	Metadata      string    `json:"metadata,omitempty" db:"metadata"` // JSON string with order-specific data
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// CreditLog records credit balance changes
type CreditLog struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Amount    float64   `json:"amount" db:"amount"`
	Balance   float64   `json:"balance" db:"balance"`
	Reason    string    `json:"reason" db:"reason"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// GroupMessage stores collected group messages for AI evaluation
type GroupMessage struct {
	ID           int64     `json:"id" db:"id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	TelegramMsgID int      `json:"telegram_msg_id" db:"telegram_msg_id"`
	TelegramName string    `json:"telegram_name" db:"telegram_name"`
	Text         string    `json:"text" db:"text"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// APIToken represents an admin API access token
type APIToken struct {
	ID          int64     `json:"id" db:"id"`
	TokenHash   string    `json:"-" db:"token_hash"`
	Name        string    `json:"name" db:"name"`
	Permissions string    `json:"permissions" db:"permissions"` // JSON array of permission strings
	CreatedBy   int64     `json:"created_by" db:"created_by"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// SignupLog tracks daily signup records
type SignupLog struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Date      string    `json:"date" db:"date"` // YYYY-MM-DD
	Value     float64   `json:"value" db:"value"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// IPChangeLog tracks IP change cooldowns
type IPChangeLog struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	OldIP     string    `json:"old_ip" db:"old_ip"`
	NewIP     string    `json:"new_ip" db:"new_ip"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
