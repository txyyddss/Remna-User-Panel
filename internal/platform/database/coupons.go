package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/coupons"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

const couponSelect = `SELECT id,code,name,kind,discount_mode,value_minor_or_bps,percent_cap_minor,
	eligible_combo_ids,eligible_squad_ids,expires_at,global_use_limit,per_user_use_limit,active,created_at,updated_at
	FROM coupon_definitions`

// SaveCoupon creates or updates a canonical coupon definition.
func (s *Store) SaveCoupon(ctx context.Context, input coupons.CouponInput, now time.Time) (coupons.Coupon, error) {
	normalized, err := input.Normalize()
	if err != nil {
		return coupons.Coupon{}, err
	}
	now = now.UTC()
	comboJSON, err := json.Marshal(normalized.EligibleComboIDs)
	if err != nil {
		return coupons.Coupon{}, fmt.Errorf("encode coupon combo eligibility: %w", err)
	}
	squadJSON, err := json.Marshal(normalized.EligibleSquadIDs)
	if err != nil {
		return coupons.Coupon{}, fmt.Errorf("encode coupon squad eligibility: %w", err)
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if normalized.ID == "" {
		normalized.ID, err = ids.New()
		if err != nil {
			return coupons.Coupon{}, err
		}
		_, err = s.db.ExecContext(ctx, `INSERT INTO coupon_definitions(id,code,name,kind,discount_mode,value_minor_or_bps,percent_cap_minor,
			eligible_combo_ids,eligible_squad_ids,expires_at,global_use_limit,per_user_use_limit,active,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, normalized.ID, normalized.Code, normalized.Name, normalized.Kind, normalized.DiscountMode,
			normalized.ValueMinorOrBPS, nullableInt64(normalized.PercentCapMinor), string(comboJSON), string(squadJSON), nullableStamp(normalized.ExpiresAt),
			nullableInt64(normalized.GlobalUseLimit), nullableInt64(normalized.PerUserUseLimit), boolInt(normalized.Active), stamp(now), stamp(now))
	} else {
		var result sql.Result
		result, err = s.db.ExecContext(ctx, `UPDATE coupon_definitions SET code=?,name=?,kind=?,discount_mode=?,value_minor_or_bps=?,percent_cap_minor=?,
			eligible_combo_ids=?,eligible_squad_ids=?,expires_at=?,global_use_limit=?,per_user_use_limit=?,active=?,updated_at=? WHERE id=?`,
			normalized.Code, normalized.Name, normalized.Kind, normalized.DiscountMode, normalized.ValueMinorOrBPS, nullableInt64(normalized.PercentCapMinor),
			string(comboJSON), string(squadJSON), nullableStamp(normalized.ExpiresAt), nullableInt64(normalized.GlobalUseLimit), nullableInt64(normalized.PerUserUseLimit),
			boolInt(normalized.Active), stamp(now), normalized.ID)
		if err == nil {
			if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
				return coupons.Coupon{}, fmt.Errorf("inspect coupon update: %w", rowsErr)
			} else if affected == 0 {
				return coupons.Coupon{}, ErrNotFound
			}
		}
	}
	if err != nil {
		if isUniqueConstraint(err) {
			return coupons.Coupon{}, ErrConflict
		}
		return coupons.Coupon{}, fmt.Errorf("save coupon: %w", err)
	}
	return couponByID(ctx, s.db, normalized.ID)
}

// ListCoupons returns coupon definitions newest first.
func (s *Store) ListCoupons(ctx context.Context, activeOnly bool) ([]coupons.Coupon, error) {
	query := couponSelect
	args := make([]any, 0, 1)
	if activeOnly {
		query += ` WHERE active=1 AND (expires_at IS NULL OR expires_at>?)`
		args = append(args, stamp(time.Now().UTC()))
	}
	query += ` ORDER BY created_at DESC,id DESC`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list coupons: %w", err)
	}
	defer func() { _ = rows.Close() }()
	result := make([]coupons.Coupon, 0)
	for rows.Next() {
		coupon, scanErr := scanCoupon(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, coupon)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate coupons: %w", err)
	}
	return result, nil
}

// RedeemCoupon atomically creates a wallet grant or applies an immediate balance effect.
func (s *Store) RedeemCoupon(ctx context.Context, userID, code, idempotencyKey string, now time.Time) (coupons.RedemptionResult, error) {
	canonical, err := coupons.CanonicalCode(code)
	if err != nil {
		return coupons.RedemptionResult{}, err
	}
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(idempotencyKey) == "" || len(idempotencyKey) > 128 {
		return coupons.RedemptionResult{}, coupons.ErrInvalidInput
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	now = now.UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return coupons.RedemptionResult{}, fmt.Errorf("begin coupon redemption: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if existing, loadErr := redemptionByKeyTx(ctx, tx, userID, idempotencyKey); loadErr == nil {
		existing.Replayed = true
		return existing, nil
	} else if !errors.Is(loadErr, ErrNotFound) {
		return coupons.RedemptionResult{}, loadErr
	}
	coupon, err := couponByCodeTx(ctx, tx, canonical)
	if err != nil {
		return coupons.RedemptionResult{}, err
	}
	if err := couponAvailable(coupon, now); err != nil {
		return coupons.RedemptionResult{}, err
	}
	redemptionID, err := ids.New()
	if err != nil {
		return coupons.RedemptionResult{}, err
	}
	balance, err := balanceTx(ctx, tx, userID)
	if err != nil {
		return coupons.RedemptionResult{}, err
	}
	var grantID *string
	delta := int64(0)
	switch coupon.Kind {
	case coupons.KindPurchaseRecurring, coupons.KindPurchaseOnce:
		grant, grantErr := grantCouponTx(ctx, tx, userID, coupon, "code", coupon.Code, now)
		if grantErr != nil {
			return coupons.RedemptionResult{}, grantErr
		}
		grantID = &grant.ID
	case coupons.KindBalanceAdd, coupons.KindBalanceMultiply:
		if err := ensureCouponUseAvailableTx(ctx, tx, coupon, userID); err != nil {
			return coupons.RedemptionResult{}, err
		}
		if coupon.Kind == coupons.KindBalanceAdd {
			delta = coupon.ValueMinorOrBPS
		} else {
			delta, err = coupons.CalculateBalanceMultiplyCredit(balance, coupon.ValueMinorOrBPS)
			if err != nil {
				return coupons.RedemptionResult{}, err
			}
		}
		balance, err = changeBalanceTx(ctx, tx, userID, delta, now)
		if err != nil {
			return coupons.RedemptionResult{}, err
		}
		ledgerKind := "coupon_balance_add"
		if coupon.Kind == coupons.KindBalanceMultiply {
			ledgerKind = "coupon_balance_multiply"
		}
		if _, err := insertLedgerTx(ctx, tx, userID, delta, balance, ledgerKind, redemptionID, coupon.Code, now); err != nil {
			return coupons.RedemptionResult{}, err
		}
	default:
		return coupons.RedemptionResult{}, coupons.ErrInvalidInput
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO coupon_redemptions(id,coupon_id,grant_id,user_id,balance_delta_minor,balance_after_minor,idempotency_key,created_at)
		VALUES(?,?,?,?,?,?,?,?)`, redemptionID, coupon.ID, nullableStringValue(grantID), userID, delta, balance, idempotencyKey, stamp(now))
	if err != nil {
		return coupons.RedemptionResult{}, fmt.Errorf("record coupon redemption: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return coupons.RedemptionResult{}, fmt.Errorf("commit coupon redemption: %w", err)
	}
	return s.redemptionByID(ctx, redemptionID)
}

// GrantCoupon idempotently puts a purchase-discount coupon in a member wallet.
func (s *Store) GrantCoupon(ctx context.Context, userID, couponID, sourceType, sourceID string, now time.Time) (coupons.Grant, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(couponID) == "" || strings.TrimSpace(sourceType) == "" || strings.TrimSpace(sourceID) == "" {
		return coupons.Grant{}, coupons.ErrInvalidInput
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return coupons.Grant{}, err
	}
	defer func() { _ = tx.Rollback() }()
	coupon, err := couponByID(ctx, tx, couponID)
	if err != nil {
		return coupons.Grant{}, err
	}
	if err := couponAvailable(coupon, now); err != nil {
		return coupons.Grant{}, err
	}
	grant, err := grantCouponTx(ctx, tx, userID, coupon, sourceType, sourceID, now.UTC())
	if err != nil {
		return coupons.Grant{}, err
	}
	if err := tx.Commit(); err != nil {
		return coupons.Grant{}, err
	}
	return grant, nil
}

// ListCouponGrants returns currently usable purchase-discount grants.
func (s *Store) ListCouponGrants(ctx context.Context, userID string, now time.Time) ([]coupons.Grant, error) {
	if _, err := s.db.ExecContext(ctx, `UPDATE coupon_grants SET status='expired' WHERE user_id=? AND status='active'
		AND coupon_id IN (SELECT id FROM coupon_definitions WHERE expires_at IS NOT NULL AND expires_at<=?)`, userID, stamp(now)); err != nil {
		return nil, fmt.Errorf("expire coupon grants: %w", err)
	}
	rows, err := s.db.QueryContext(ctx, grantSelect+` WHERE coupon_grants.user_id=? AND coupon_grants.status='active'
		AND coupon_definitions.active=1 AND (coupon_definitions.expires_at IS NULL OR coupon_definitions.expires_at>?) ORDER BY coupon_grants.created_at DESC`, userID, stamp(now))
	if err != nil {
		return nil, fmt.Errorf("list coupon grants: %w", err)
	}
	defer func() { _ = rows.Close() }()
	result := make([]coupons.Grant, 0)
	for rows.Next() {
		grant, scanErr := scanGrant(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, grant)
	}
	return result, rows.Err()
}

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
		&percentCap, &comboJSON, &squadJSON, &expires, &globalLimit, &userLimit, &active, &created, &updated); err != nil {
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

const grantSelect = `SELECT coupon_grants.id,coupon_grants.user_id,coupon_grants.source_type,coupon_grants.source_id,coupon_grants.status,
	coupon_grants.use_count,coupon_grants.created_at,coupon_grants.consumed_at,
	coupon_definitions.id,coupon_definitions.code,coupon_definitions.name,coupon_definitions.kind,coupon_definitions.discount_mode,
	coupon_definitions.value_minor_or_bps,coupon_definitions.percent_cap_minor,coupon_definitions.eligible_combo_ids,
	coupon_definitions.eligible_squad_ids,coupon_definitions.expires_at,coupon_definitions.global_use_limit,
	coupon_definitions.per_user_use_limit,coupon_definitions.active,coupon_definitions.created_at,coupon_definitions.updated_at
	FROM coupon_grants JOIN coupon_definitions ON coupon_definitions.id=coupon_grants.coupon_id`

func grantByIDTx(ctx context.Context, tx *sql.Tx, grantID string) (coupons.Grant, error) {
	return scanGrant(tx.QueryRowContext(ctx, grantSelect+` WHERE coupon_grants.id=?`, grantID))
}

func grantBySourceTx(ctx context.Context, tx *sql.Tx, userID, couponID, sourceType, sourceID string) (coupons.Grant, error) {
	return scanGrant(tx.QueryRowContext(ctx, grantSelect+` WHERE coupon_grants.user_id=? AND coupon_grants.coupon_id=? AND coupon_grants.source_type=? AND coupon_grants.source_id=?`, userID, couponID, sourceType, sourceID))
}

func scanGrant(row rowScanner) (coupons.Grant, error) {
	var grant coupons.Grant
	var consumed sql.NullString
	var grantCreated string
	var percentCap, globalLimit, userLimit sql.NullInt64
	var comboJSON, squadJSON string
	var expires sql.NullString
	var active int
	var couponCreated, couponUpdated string
	if err := row.Scan(&grant.ID, &grant.UserID, &grant.SourceType, &grant.SourceID, &grant.Status, &grant.UseCount, &grantCreated, &consumed,
		&grant.Coupon.ID, &grant.Coupon.Code, &grant.Coupon.Name, &grant.Coupon.Kind, &grant.Coupon.DiscountMode,
		&grant.Coupon.ValueMinorOrBPS, &percentCap, &comboJSON, &squadJSON, &expires, &globalLimit, &userLimit, &active, &couponCreated, &couponUpdated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return coupons.Grant{}, ErrNotFound
		}
		return coupons.Grant{}, fmt.Errorf("scan coupon grant: %w", err)
	}
	grant.Coupon.PercentCapMinor = int64Pointer(percentCap)
	grant.Coupon.GlobalUseLimit = int64Pointer(globalLimit)
	grant.Coupon.PerUserUseLimit = int64Pointer(userLimit)
	grant.Coupon.Active = active == 1
	if err := json.Unmarshal([]byte(comboJSON), &grant.Coupon.EligibleComboIDs); err != nil {
		return coupons.Grant{}, err
	}
	if err := json.Unmarshal([]byte(squadJSON), &grant.Coupon.EligibleSquadIDs); err != nil {
		return coupons.Grant{}, err
	}
	var err error
	if grant.CreatedAt, err = parseStamp(grantCreated); err != nil {
		return coupons.Grant{}, err
	}
	if consumed.Valid {
		value, parseErr := parseStamp(consumed.String)
		if parseErr != nil {
			return coupons.Grant{}, parseErr
		}
		grant.ConsumedAt = &value
	}
	if expires.Valid {
		value, parseErr := parseStamp(expires.String)
		if parseErr != nil {
			return coupons.Grant{}, parseErr
		}
		grant.Coupon.ExpiresAt = &value
	}
	if grant.Coupon.CreatedAt, err = parseStamp(couponCreated); err != nil {
		return coupons.Grant{}, err
	}
	grant.Coupon.UpdatedAt, err = parseStamp(couponUpdated)
	return grant, err
}

func (s *Store) redemptionByID(ctx context.Context, redemptionID string) (coupons.RedemptionResult, error) {
	result, err := scanRedemption(s.db.QueryRowContext(ctx, redemptionSelect+` WHERE coupon_redemptions.id=?`, redemptionID))
	if err != nil {
		return coupons.RedemptionResult{}, err
	}
	if result.Grant != nil {
		grant, grantErr := scanGrant(s.db.QueryRowContext(ctx, grantSelect+` WHERE coupon_grants.id=?`, result.Grant.ID))
		if grantErr != nil {
			return coupons.RedemptionResult{}, grantErr
		}
		result.Grant = &grant
	}
	return result, nil
}

const redemptionSelect = `SELECT coupon_redemptions.id,coupon_redemptions.balance_delta_minor,coupon_redemptions.balance_after_minor,
	coupon_redemptions.idempotency_key,coupon_redemptions.created_at,coupon_redemptions.grant_id,
	coupon_definitions.id,coupon_definitions.code,coupon_definitions.name,coupon_definitions.kind,coupon_definitions.discount_mode,
	coupon_definitions.value_minor_or_bps,coupon_definitions.percent_cap_minor,coupon_definitions.eligible_combo_ids,
	coupon_definitions.eligible_squad_ids,coupon_definitions.expires_at,coupon_definitions.global_use_limit,
	coupon_definitions.per_user_use_limit,coupon_definitions.active,coupon_definitions.created_at,coupon_definitions.updated_at
	FROM coupon_redemptions JOIN coupon_definitions ON coupon_definitions.id=coupon_redemptions.coupon_id`

func redemptionByKeyTx(ctx context.Context, tx *sql.Tx, userID, key string) (coupons.RedemptionResult, error) {
	result, err := scanRedemption(tx.QueryRowContext(ctx, redemptionSelect+` WHERE coupon_redemptions.user_id=? AND coupon_redemptions.idempotency_key=?`, userID, key))
	if err != nil {
		return coupons.RedemptionResult{}, err
	}
	if result.Grant != nil {
		grant, grantErr := grantByIDTx(ctx, tx, result.Grant.ID)
		if grantErr != nil {
			return coupons.RedemptionResult{}, grantErr
		}
		result.Grant = &grant
	}
	return result, nil
}

func scanRedemption(row rowScanner) (coupons.RedemptionResult, error) {
	var result coupons.RedemptionResult
	var grantID sql.NullString
	var created string
	var percentCap, globalLimit, userLimit sql.NullInt64
	var comboJSON, squadJSON string
	var expires sql.NullString
	var active int
	var couponCreated, couponUpdated string
	if err := row.Scan(&result.ID, &result.BalanceDeltaMinor, &result.BalanceAfterMinor, &result.IdempotencyKey, &created, &grantID,
		&result.Coupon.ID, &result.Coupon.Code, &result.Coupon.Name, &result.Coupon.Kind, &result.Coupon.DiscountMode,
		&result.Coupon.ValueMinorOrBPS, &percentCap, &comboJSON, &squadJSON, &expires, &globalLimit, &userLimit, &active, &couponCreated, &couponUpdated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return coupons.RedemptionResult{}, ErrNotFound
		}
		return coupons.RedemptionResult{}, fmt.Errorf("scan coupon redemption: %w", err)
	}
	result.Coupon.PercentCapMinor = int64Pointer(percentCap)
	result.Coupon.GlobalUseLimit = int64Pointer(globalLimit)
	result.Coupon.PerUserUseLimit = int64Pointer(userLimit)
	result.Coupon.Active = active == 1
	if grantID.Valid {
		result.Grant = &coupons.Grant{ID: grantID.String}
	}
	if err := json.Unmarshal([]byte(comboJSON), &result.Coupon.EligibleComboIDs); err != nil {
		return coupons.RedemptionResult{}, err
	}
	if err := json.Unmarshal([]byte(squadJSON), &result.Coupon.EligibleSquadIDs); err != nil {
		return coupons.RedemptionResult{}, err
	}
	var err error
	if result.CreatedAt, err = parseStamp(created); err != nil {
		return coupons.RedemptionResult{}, err
	}
	if expires.Valid {
		value, parseErr := parseStamp(expires.String)
		if parseErr != nil {
			return coupons.RedemptionResult{}, parseErr
		}
		result.Coupon.ExpiresAt = &value
	}
	if result.Coupon.CreatedAt, err = parseStamp(couponCreated); err != nil {
		return coupons.RedemptionResult{}, err
	}
	result.Coupon.UpdatedAt, err = parseStamp(couponUpdated)
	return result, err
}

func balanceTx(ctx context.Context, tx *sql.Tx, userID string) (int64, error) {
	var balance int64
	if err := tx.QueryRowContext(ctx, `SELECT txb_minor FROM balances WHERE user_id=?`, userID).Scan(&balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("read TXB balance: %w", err)
	}
	return balance, nil
}

func checkedBalanceAddition(current, delta int64) (int64, error) {
	if (delta > 0 && current > math.MaxInt64-delta) || (delta < 0 && current < math.MinInt64-delta) {
		return 0, fmt.Errorf("TXB balance overflow")
	}
	return current + delta, nil
}

func setBalanceTx(ctx context.Context, tx *sql.Tx, userID string, current, next int64, now time.Time) error {
	result, err := tx.ExecContext(ctx, `UPDATE balances SET txb_minor=?,updated_at=? WHERE user_id=? AND txb_minor=?`, next, stamp(now), userID, current)
	if err != nil {
		return fmt.Errorf("change TXB balance: %w", err)
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return rowsErr
	} else if affected != 1 {
		return ErrConflict
	}
	return nil
}

// adjustBalanceTx performs a signed balance change without silently allowing
// SQLite to promote an overflowing INTEGER expression to REAL. It is reserved
// for operations that may intentionally create debt, such as administrator
// adjustments and provider reversals.
func adjustBalanceTx(ctx context.Context, tx *sql.Tx, userID string, delta int64, now time.Time) (int64, error) {
	current, err := balanceTx(ctx, tx, userID)
	if err != nil {
		return 0, err
	}
	next, err := checkedBalanceAddition(current, delta)
	if err != nil {
		return 0, err
	}
	if err := setBalanceTx(ctx, tx, userID, current, next, now); err != nil {
		return 0, err
	}
	return next, nil
}

func changeBalanceTx(ctx context.Context, tx *sql.Tx, userID string, delta int64, now time.Time) (int64, error) {
	current, err := balanceTx(ctx, tx, userID)
	if err != nil {
		return 0, err
	}
	next, err := checkedBalanceAddition(current, delta)
	if err != nil {
		return 0, err
	}
	if delta < 0 && next < 0 {
		return 0, ErrInsufficientBalance
	}
	if err := setBalanceTx(ctx, tx, userID, current, next, now); err != nil {
		return 0, err
	}
	return next, nil
}

func nullableInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableStamp(value *time.Time) any {
	if value == nil {
		return nil
	}
	return stamp(*value)
}

func nullableStringValue(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func int64Pointer(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Int64
	return &result
}

func isUniqueConstraint(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint")
}
