// Package coupons implements canonical coupon codes, wallet grants, and effects.
package coupons

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ErrInvalidInput indicates that a coupon or redemption request is unsafe.
var ErrInvalidInput = errors.New("invalid coupon input")

// CouponKind identifies how a coupon affects a member account.
type CouponKind string

const (
	// KindPurchaseRecurring creates a reusable purchase-discount grant.
	KindPurchaseRecurring CouponKind = "purchase_recurring"
	// KindPurchaseOnce creates a grant consumed by one purchase.
	KindPurchaseOnce CouponKind = "purchase_once"
	// KindBalanceAdd credits a fixed TXB amount when redeemed.
	KindBalanceAdd CouponKind = "balance_add"
	// KindBalanceMultiply multiplies the current balance when redeemed.
	KindBalanceMultiply CouponKind = "balance_multiply"
)

// DiscountMode identifies fixed or percentage purchase pricing.
type DiscountMode string

const (
	// DiscountFixed subtracts a fixed number of TXB minor units.
	DiscountFixed DiscountMode = "fixed"
	// DiscountPercent subtracts a basis-point percentage of the gross price.
	DiscountPercent DiscountMode = "percent"
)

var codePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{3,63}$`)

// CanonicalCode trims and uppercases a coupon code.
func CanonicalCode(code string) (string, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if !codePattern.MatchString(code) {
		return "", fmt.Errorf("%w: code must contain 4 to 64 ASCII letters, digits, underscore, or hyphen", ErrInvalidInput)
	}
	return code, nil
}

// CouponInput is the administrator-authored coupon definition.
// ValueMinorOrBPS is TXB minor units for fixed/add coupons and basis points for
// percentage/multiply coupons.
type CouponInput struct {
	ID               string       `json:"id,omitempty"`
	Code             string       `json:"code"`
	Name             string       `json:"name"`
	Kind             CouponKind   `json:"kind"`
	DiscountMode     DiscountMode `json:"discountMode,omitempty"`
	ValueMinorOrBPS  int64        `json:"valueMinorOrBps"`
	PercentCapMinor  *int64       `json:"percentCapMinor,omitempty"`
	EligibleComboIDs []string     `json:"eligibleComboIds"`
	EligibleSquadIDs []string     `json:"eligibleSquadIds"`
	ExpiresAt        *time.Time   `json:"expiresAt,omitempty"`
	GlobalUseLimit   *int64       `json:"globalUseLimit,omitempty"`
	PerUserUseLimit  *int64       `json:"perUserUseLimit,omitempty"`
	Active           bool         `json:"active"`
}

// Validate rejects ambiguous coupon definitions.
func (input CouponInput) Validate() error {
	if _, err := CanonicalCode(input.Code); err != nil {
		return err
	}
	if name := strings.TrimSpace(input.Name); name == "" || len(name) > 100 {
		return fmt.Errorf("%w: coupon name must be 1 to 100 bytes", ErrInvalidInput)
	}
	if input.GlobalUseLimit != nil && *input.GlobalUseLimit <= 0 {
		return fmt.Errorf("%w: global use limit must be positive", ErrInvalidInput)
	}
	if input.PerUserUseLimit != nil && *input.PerUserUseLimit <= 0 {
		return fmt.Errorf("%w: per-user use limit must be positive", ErrInvalidInput)
	}
	if len(input.EligibleComboIDs) > 500 || len(input.EligibleSquadIDs) > 500 {
		return fmt.Errorf("%w: too many eligibility targets", ErrInvalidInput)
	}
	if hasBlank(input.EligibleComboIDs) || hasBlank(input.EligibleSquadIDs) {
		return fmt.Errorf("%w: eligibility IDs cannot be blank", ErrInvalidInput)
	}
	switch input.Kind {
	case KindPurchaseRecurring, KindPurchaseOnce:
		if input.DiscountMode != DiscountFixed && input.DiscountMode != DiscountPercent {
			return fmt.Errorf("%w: purchase coupon needs fixed or percent discount mode", ErrInvalidInput)
		}
		if input.ValueMinorOrBPS <= 0 {
			return fmt.Errorf("%w: purchase discount must be positive", ErrInvalidInput)
		}
		if input.DiscountMode == DiscountPercent {
			if input.ValueMinorOrBPS > 10_000 {
				return fmt.Errorf("%w: percentage discount cannot exceed 100 percent", ErrInvalidInput)
			}
			if input.PercentCapMinor != nil && *input.PercentCapMinor <= 0 {
				return fmt.Errorf("%w: percentage cap must be positive", ErrInvalidInput)
			}
		} else if input.PercentCapMinor != nil {
			return fmt.Errorf("%w: fixed discount cannot have a percentage cap", ErrInvalidInput)
		}
	case KindBalanceAdd:
		if input.ValueMinorOrBPS <= 0 || input.DiscountMode != "" || input.PercentCapMinor != nil {
			return fmt.Errorf("%w: balance-add coupon needs only a positive TXB value", ErrInvalidInput)
		}
	case KindBalanceMultiply:
		if input.ValueMinorOrBPS <= 10_000 || input.ValueMinorOrBPS > 1_000_000 || input.DiscountMode != "" || input.PercentCapMinor != nil {
			return fmt.Errorf("%w: balance multiplier must be greater than 1x and at most 100x", ErrInvalidInput)
		}
	default:
		return fmt.Errorf("%w: unsupported coupon kind %q", ErrInvalidInput, input.Kind)
	}
	if input.Kind == KindBalanceAdd || input.Kind == KindBalanceMultiply {
		if len(input.EligibleComboIDs) != 0 || len(input.EligibleSquadIDs) != 0 {
			return fmt.Errorf("%w: balance coupon cannot have purchase eligibility", ErrInvalidInput)
		}
	}
	return nil
}

// Normalize returns a validated canonical representation suitable for storage.
func (input CouponInput) Normalize() (CouponInput, error) {
	if err := input.Validate(); err != nil {
		return CouponInput{}, err
	}
	code, _ := CanonicalCode(input.Code)
	input.Code = code
	input.Name = strings.TrimSpace(input.Name)
	input.EligibleComboIDs = uniqueSorted(input.EligibleComboIDs)
	input.EligibleSquadIDs = uniqueSorted(input.EligibleSquadIDs)
	if input.ExpiresAt != nil {
		value := input.ExpiresAt.UTC()
		input.ExpiresAt = &value
	}
	return input, nil
}

// Coupon is a persisted definition.
type Coupon struct {
	CouponInput
	UsageCount int64     `json:"usageCount"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// Grant is a coupon held in a member's wallet.
type Grant struct {
	ID         string     `json:"id"`
	Coupon     Coupon     `json:"coupon"`
	UserID     string     `json:"userId,omitempty"`
	SourceType string     `json:"sourceType"`
	SourceID   string     `json:"sourceId"`
	Status     string     `json:"status"`
	UseCount   int64      `json:"useCount"`
	CreatedAt  time.Time  `json:"createdAt"`
	ConsumedAt *time.Time `json:"consumedAt,omitempty"`
}

// RedemptionResult contains either a wallet grant or an immediate balance effect.
type RedemptionResult struct {
	ID                string    `json:"id"`
	Coupon            Coupon    `json:"coupon"`
	Grant             *Grant    `json:"grant,omitempty"`
	BalanceDeltaMinor int64     `json:"balanceDeltaMinor"`
	BalanceAfterMinor int64     `json:"balanceAfterMinor"`
	IdempotencyKey    string    `json:"idempotencyKey,omitempty"`
	Replayed          bool      `json:"replayed"`
	CreatedAt         time.Time `json:"createdAt"`
}

// PurchaseContext contains only server-priced inputs used for eligibility.
type PurchaseContext struct {
	UserID          string
	GrantID         string
	ComboID         string
	AddonSquadIDs   []string
	GrossPriceMinor int64
}

// Discount is a server-calculated coupon quote or consumed purchase effect.
type Discount struct {
	GrantID       string `json:"grantId"`
	CouponID      string `json:"couponId"`
	CouponCode    string `json:"couponCode"`
	GrossMinor    int64  `json:"grossMinor"`
	DiscountMinor int64  `json:"discountMinor"`
	NetMinor      int64  `json:"netMinor"`
	Recurring     bool   `json:"recurring"`
}

// EligibleFor returns whether a coupon applies to a combo or any selected add-on.
func (coupon Coupon) EligibleFor(comboID string, addonSquadIDs []string) bool {
	if len(coupon.EligibleComboIDs) == 0 && len(coupon.EligibleSquadIDs) == 0 {
		return true
	}
	for _, allowed := range coupon.EligibleComboIDs {
		if allowed == comboID {
			return true
		}
	}
	selected := make(map[string]struct{}, len(addonSquadIDs))
	for _, id := range addonSquadIDs {
		selected[id] = struct{}{}
	}
	for _, allowed := range coupon.EligibleSquadIDs {
		if _, ok := selected[allowed]; ok {
			return true
		}
	}
	return false
}

// CalculateDiscount applies fixed or basis-point pricing without floating point.
func CalculateDiscount(coupon Coupon, grossMinor int64) (int64, error) {
	if grossMinor < 0 {
		return 0, fmt.Errorf("%w: gross price cannot be negative", ErrInvalidInput)
	}
	if coupon.Kind != KindPurchaseRecurring && coupon.Kind != KindPurchaseOnce {
		return 0, fmt.Errorf("%w: coupon is not a purchase discount", ErrInvalidInput)
	}
	var discount int64
	switch coupon.DiscountMode {
	case DiscountFixed:
		discount = coupon.ValueMinorOrBPS
	case DiscountPercent:
		value, err := multiplyDivideFloor(grossMinor, coupon.ValueMinorOrBPS, 10_000)
		if err != nil {
			return 0, err
		}
		discount = value
		if coupon.PercentCapMinor != nil && discount > *coupon.PercentCapMinor {
			discount = *coupon.PercentCapMinor
		}
	default:
		return 0, fmt.Errorf("%w: unsupported discount mode", ErrInvalidInput)
	}
	if discount > grossMinor {
		discount = grossMinor
	}
	return discount, nil
}

// CalculateBalanceMultiplyCredit returns floor(balance*factor)-balance.
func CalculateBalanceMultiplyCredit(balanceMinor, multiplierBPS int64) (int64, error) {
	if balanceMinor < 0 || multiplierBPS <= 10_000 {
		return 0, fmt.Errorf("%w: invalid balance multiplier", ErrInvalidInput)
	}
	finalBalance, err := multiplyDivideFloor(balanceMinor, multiplierBPS, 10_000)
	if err != nil {
		return 0, err
	}
	return finalBalance - balanceMinor, nil
}

func multiplyDivideFloor(left, right, divisor int64) (int64, error) {
	if left < 0 || right < 0 || divisor <= 0 {
		return 0, fmt.Errorf("%w: invalid fixed-point operands", ErrInvalidInput)
	}
	value := new(big.Int).Mul(big.NewInt(left), big.NewInt(right))
	value.Quo(value, big.NewInt(divisor))
	if !value.IsInt64() {
		return 0, fmt.Errorf("%w: fixed-point result overflows int64", ErrInvalidInput)
	}
	return value.Int64(), nil
}

// Store is the narrow persistence contract required by Service.
type Store interface {
	SaveCoupon(context.Context, CouponInput, time.Time) (Coupon, error)
	ListCoupons(context.Context, bool) ([]Coupon, error)
	RedeemCoupon(context.Context, string, string, string, time.Time) (RedemptionResult, error)
	GrantCoupon(context.Context, string, string, string, string, time.Time) (Grant, error)
	ListCouponGrants(context.Context, string, time.Time) ([]Grant, error)
	QuotePurchaseCoupon(context.Context, PurchaseContext, time.Time) (Discount, error)
}

// Service validates coupon requests and delegates atomic effects to storage.
type Service struct {
	store Store
	now   func() time.Time
}

// NewService constructs the coupon application service.
func NewService(store Store, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, now: now}
}

// Save validates and persists one coupon definition.
func (service *Service) Save(ctx context.Context, input CouponInput) (Coupon, error) {
	normalized, err := input.Normalize()
	if err != nil {
		return Coupon{}, err
	}
	return service.store.SaveCoupon(ctx, normalized, service.now().UTC())
}

// List returns active member-visible coupons or all administrator definitions.
func (service *Service) List(ctx context.Context, activeOnly bool) ([]Coupon, error) {
	return service.store.ListCoupons(ctx, activeOnly)
}

// Redeem canonicalizes a code and atomically grants or applies it.
func (service *Service) Redeem(ctx context.Context, userID, code, idempotencyKey string) (RedemptionResult, error) {
	canonical, err := CanonicalCode(code)
	if err != nil {
		return RedemptionResult{}, err
	}
	if strings.TrimSpace(userID) == "" || !validIdempotencyKey(idempotencyKey) {
		return RedemptionResult{}, fmt.Errorf("%w: incomplete redemption request", ErrInvalidInput)
	}
	return service.store.RedeemCoupon(ctx, userID, canonical, idempotencyKey, service.now().UTC())
}

// Wallet returns currently usable purchase-discount grants.
func (service *Service) Wallet(ctx context.Context, userID string) ([]Grant, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("%w: missing user", ErrInvalidInput)
	}
	return service.store.ListCouponGrants(ctx, userID, service.now().UTC())
}

// Quote calculates one explicitly selected grant against a server-priced basket.
func (service *Service) Quote(ctx context.Context, input PurchaseContext) (Discount, error) {
	if strings.TrimSpace(input.UserID) == "" || strings.TrimSpace(input.GrantID) == "" || strings.TrimSpace(input.ComboID) == "" || input.GrossPriceMinor < 0 {
		return Discount{}, fmt.Errorf("%w: incomplete coupon quote", ErrInvalidInput)
	}
	input.AddonSquadIDs = uniqueSorted(input.AddonSquadIDs)
	return service.store.QuotePurchaseCoupon(ctx, input, service.now().UTC())
}

func hasBlank(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return true
		}
	}
	return false
}

func uniqueSorted(values []string) []string {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[strings.TrimSpace(value)] = struct{}{}
	}
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func validIdempotencyKey(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && len(value) <= 128
}
