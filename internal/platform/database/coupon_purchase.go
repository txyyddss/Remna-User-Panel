package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/txyyddss/Remna-User-Panel/internal/coupons"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
	"strings"
	"time"
)

// QuotePurchaseCoupon validates one selected grant against a server-priced basket.
func (s *Store) QuotePurchaseCoupon(ctx context.Context, input coupons.PurchaseContext, now time.Time) (coupons.Discount, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return coupons.Discount{}, err
	}
	defer func() { _ = tx.Rollback() }()
	discount, err := quoteCouponGrantTx(ctx, tx, input, now)
	if err != nil {
		return coupons.Discount{}, err
	}
	return discount, nil
}

// applyPurchaseCouponTx consumes one selected grant inside the purchase transaction.
// Checkout must call this only after it has assigned the immutable purchase ID and
// before it debits the returned net price.
func applyPurchaseCouponTx(ctx context.Context, tx *sql.Tx, input coupons.PurchaseContext, purchaseID string, now time.Time) (coupons.Discount, error) {
	if strings.TrimSpace(purchaseID) == "" {
		return coupons.Discount{}, coupons.ErrInvalidInput
	}
	discount, err := quoteCouponGrantTx(ctx, tx, input, now)
	if err != nil {
		return coupons.Discount{}, err
	}
	grant, err := grantByIDTx(ctx, tx, input.GrantID)
	if err != nil {
		return coupons.Discount{}, err
	}
	if err := ensureCouponUseAvailableTx(ctx, tx, grant.Coupon, input.UserID); err != nil {
		return coupons.Discount{}, err
	}
	useID, err := ids.New()
	if err != nil {
		return coupons.Discount{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO coupon_uses(id,coupon_id,grant_id,user_id,purchase_id,gross_price_minor,discount_minor,net_price_minor,created_at)
		VALUES(?,?,?,?,?,?,?,?,?)`, useID, grant.Coupon.ID, grant.ID, input.UserID, purchaseID, discount.GrossMinor, discount.DiscountMinor, discount.NetMinor, stamp(now)); err != nil {
		if isUniqueConstraint(err) {
			return coupons.Discount{}, ErrConflict
		}
		return coupons.Discount{}, fmt.Errorf("record coupon use: %w", err)
	}
	status := "active"
	var consumed any
	if grant.Coupon.Kind == coupons.KindPurchaseOnce {
		status = "consumed"
		consumed = stamp(now)
	}
	result, err := tx.ExecContext(ctx, `UPDATE coupon_grants SET use_count=use_count+1,status=?,consumed_at=? WHERE id=? AND status='active'`, status, consumed, grant.ID)
	if err != nil {
		return coupons.Discount{}, fmt.Errorf("consume coupon grant: %w", err)
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return coupons.Discount{}, rowsErr
	} else if affected != 1 {
		return coupons.Discount{}, ErrConflict
	}
	return discount, nil
}

func quoteCouponGrantTx(ctx context.Context, tx *sql.Tx, input coupons.PurchaseContext, now time.Time) (coupons.Discount, error) {
	if input.GrossPriceMinor < 0 || strings.TrimSpace(input.UserID) == "" || strings.TrimSpace(input.GrantID) == "" || strings.TrimSpace(input.ComboID) == "" {
		return coupons.Discount{}, coupons.ErrInvalidInput
	}
	grant, err := grantByIDTx(ctx, tx, input.GrantID)
	if err != nil {
		return coupons.Discount{}, err
	}
	if grant.UserID != input.UserID || grant.Status != "active" {
		return coupons.Discount{}, ErrConflict
	}
	if err := couponAvailable(grant.Coupon, now); err != nil {
		return coupons.Discount{}, err
	}
	if !grant.Coupon.EligibleFor(input.ComboID, input.AddonSquadIDs) {
		return coupons.Discount{}, ErrConflict
	}
	if err := ensureCouponUseAvailableTx(ctx, tx, grant.Coupon, input.UserID); err != nil {
		return coupons.Discount{}, err
	}
	value, err := coupons.CalculateDiscount(grant.Coupon, input.GrossPriceMinor)
	if err != nil {
		return coupons.Discount{}, err
	}
	return coupons.Discount{GrantID: grant.ID, CouponID: grant.Coupon.ID, CouponCode: grant.Coupon.Code,
		GrossMinor: input.GrossPriceMinor, DiscountMinor: value, NetMinor: input.GrossPriceMinor - value,
		Recurring: grant.Coupon.Kind == coupons.KindPurchaseRecurring}, nil
}

func grantCouponTx(ctx context.Context, tx *sql.Tx, userID string, coupon coupons.Coupon, sourceType, sourceID string, now time.Time) (coupons.Grant, error) {
	if coupon.Kind != coupons.KindPurchaseRecurring && coupon.Kind != coupons.KindPurchaseOnce {
		return coupons.Grant{}, fmt.Errorf("%w: only purchase coupons can be wallet grants", coupons.ErrInvalidInput)
	}
	if existing, err := grantBySourceTx(ctx, tx, userID, coupon.ID, sourceType, sourceID); err == nil {
		return existing, nil
	} else if !errors.Is(err, ErrNotFound) {
		return coupons.Grant{}, err
	}
	grantID, err := ids.New()
	if err != nil {
		return coupons.Grant{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO coupon_grants(id,coupon_id,user_id,source_type,source_id,status,use_count,created_at)
		VALUES(?,?,?,?,?,'active',0,?)`, grantID, coupon.ID, userID, sourceType, sourceID, stamp(now))
	if err != nil {
		return coupons.Grant{}, fmt.Errorf("grant coupon: %w", err)
	}
	return coupons.Grant{ID: grantID, Coupon: coupon, UserID: userID, SourceType: sourceType, SourceID: sourceID, Status: "active", CreatedAt: now}, nil
}

func ensureCouponUseAvailableTx(ctx context.Context, tx *sql.Tx, coupon coupons.Coupon, userID string) error {
	var globalCount, userCount int64
	query := `SELECT
		(SELECT COUNT(*) FROM coupon_uses WHERE coupon_id=?) +
		(SELECT COUNT(*) FROM coupon_redemptions r JOIN coupon_definitions d ON d.id=r.coupon_id WHERE r.coupon_id=? AND d.kind IN ('balance_add','balance_multiply')),
		(SELECT COUNT(*) FROM coupon_uses WHERE coupon_id=? AND user_id=?) +
		(SELECT COUNT(*) FROM coupon_redemptions r JOIN coupon_definitions d ON d.id=r.coupon_id WHERE r.coupon_id=? AND r.user_id=? AND d.kind IN ('balance_add','balance_multiply'))`
	if err := tx.QueryRowContext(ctx, query, coupon.ID, coupon.ID, coupon.ID, userID, coupon.ID, userID).Scan(&globalCount, &userCount); err != nil {
		return fmt.Errorf("count coupon uses: %w", err)
	}
	if coupon.GlobalUseLimit != nil && globalCount >= *coupon.GlobalUseLimit {
		return ErrConflict
	}
	if coupon.PerUserUseLimit != nil && userCount >= *coupon.PerUserUseLimit {
		return ErrConflict
	}
	return nil
}

func couponAvailable(coupon coupons.Coupon, now time.Time) error {
	if !coupon.Active || (coupon.ExpiresAt != nil && !coupon.ExpiresAt.After(now)) {
		return ErrConflict
	}
	return nil
}

func couponByID(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, couponID string) (coupons.Coupon, error) {
	return scanCoupon(queryer.QueryRowContext(ctx, couponSelect+` WHERE id=?`, couponID))
}

func couponByCodeTx(ctx context.Context, tx *sql.Tx, code string) (coupons.Coupon, error) {
	return scanCoupon(tx.QueryRowContext(ctx, couponSelect+` WHERE code=?`, code))
}

func scanCoupon(row rowScanner) (coupons.Coupon, error) {
	var coupon coupons.Coupon
	var percentCap, globalLimit, userLimit sql.NullInt64
	var comboJSON, squadJSON string
	var expires sql.NullString
	var active int
	var created, updated string
	if err := row.Scan(&coupon.ID, &coupon.Code, &coupon.Name, &coupon.Kind, &coupon.DiscountMode, &coupon.ValueMinorOrBPS,
		&percentCap, &comboJSON, &squadJSON, &expires, &globalLimit, &userLimit, &active, &created, &updated, &coupon.UsageCount); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return coupons.Coupon{}, ErrNotFound
		}
		return coupons.Coupon{}, fmt.Errorf("scan coupon: %w", err)
	}
	coupon.Active = active == 1
	coupon.PercentCapMinor = int64Pointer(percentCap)
	coupon.GlobalUseLimit = int64Pointer(globalLimit)
	coupon.PerUserUseLimit = int64Pointer(userLimit)
	if err := json.Unmarshal([]byte(comboJSON), &coupon.EligibleComboIDs); err != nil {
		return coupons.Coupon{}, fmt.Errorf("decode coupon combo eligibility: %w", err)
	}
	if err := json.Unmarshal([]byte(squadJSON), &coupon.EligibleSquadIDs); err != nil {
		return coupons.Coupon{}, fmt.Errorf("decode coupon squad eligibility: %w", err)
	}
	var err error
	if expires.Valid {
		value, parseErr := parseStamp(expires.String)
		if parseErr != nil {
			return coupons.Coupon{}, parseErr
		}
		coupon.ExpiresAt = &value
	}
	if coupon.CreatedAt, err = parseStamp(created); err != nil {
		return coupons.Coupon{}, err
	}
	coupon.UpdatedAt, err = parseStamp(updated)
	return coupon, err
}
