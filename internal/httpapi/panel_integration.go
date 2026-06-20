package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"remna-user-panel/internal/config"
	"remna-user-panel/internal/mail"
	"remna-user-panel/internal/payments"
	"remna-user-panel/internal/remnawave"
	appsettings "remna-user-panel/internal/settings"
	"remna-user-panel/internal/tariffs"
)

func panelUserForWebUser(ctx context.Context, pool *pgxpool.Pool, panel *remnawave.Client, user webappUser) (map[string]any, bool, error) {
	if panel == nil || !panel.Configured(ctx) {
		return nil, false, nil
	}
	if user.PanelUserUUID != "" {
		panelUser, ok, err := panel.GetUserByUUID(ctx, user.PanelUserUUID)
		if err != nil {
			return nil, false, err
		}
		if ok {
			return panelUser, true, nil
		}
	}
	if user.TelegramID != 0 {
		users, err := panel.GetUsersByTelegramID(ctx, user.TelegramID)
		if err != nil {
			return nil, false, err
		}
		if len(users) == 1 {
			bindPanelUUID(ctx, pool, user.UserID, stringValue(users[0], "uuid"))
			return users[0], true, nil
		}
	}
	if user.Email != "" {
		users, err := panel.GetUsersByEmail(ctx, user.Email)
		if err != nil {
			return nil, false, err
		}
		if len(users) == 1 {
			bindPanelUUID(ctx, pool, user.UserID, stringValue(users[0], "uuid"))
			return users[0], true, nil
		}
	}
	return nil, false, nil
}

func subscriptionFromPanelUser(ctx context.Context, pool *pgxpool.Pool, user webappUser, panelUser map[string]any) map[string]any {
	if panelUser == nil {
		return map[string]any{"active": false}
	}
	plan := latestPaidPlanForUser(ctx, pool, user.UserID)
	expireAt := parsePanelTime(panelUser["expireAt"])
	now := time.Now().UTC()
	status := strings.ToUpper(strings.TrimSpace(stringValue(panelUser, "status")))
	active := status == "ACTIVE" || status == "LIMITED"
	if !expireAt.IsZero() && expireAt.Before(now) {
		active = false
		if status == "" || status == "ACTIVE" {
			status = "EXPIRED"
		}
	}
	traffic := mapValue(panelUser, "userTraffic")
	usedBytes := int64Value(traffic, "usedTrafficBytes")
	if usedBytes == 0 {
		usedBytes = int64Value(panelUser, "usedTrafficBytes")
	}
	limitBytes := int64Value(panelUser, "trafficLimitBytes")
	days := 0
	if !expireAt.IsZero() && expireAt.After(now) {
		days = int(math.Ceil(expireAt.Sub(now).Hours() / 24))
	}
	tariffName := plan.TariffName
	if tariffName == "" {
		tariffName = plan.Title
	}
	override := loadUserTrafficOverride(ctx, pool, user.UserID)
	payload := map[string]any{
		"active":                     active,
		"status":                     status,
		"end_date":                   timeString(expireAt),
		"end_date_text":              dateText(expireAt),
		"days_left":                  days,
		"remaining_text":             remainingText(days),
		"config_link":                stringValue(panelUser, "subscriptionUrl"),
		"connect_url":                stringValue(panelUser, "subscriptionUrl"),
		"panel_short_uuid":           stringValue(panelUser, "shortUuid"),
		"traffic_used":               bytesText(usedBytes),
		"traffic_limit":              bytesText(limitBytes),
		"traffic_used_bytes":         usedBytes,
		"traffic_limit_bytes":        limitBytes,
		"traffic_limit_strategy":     stringValue(panelUser, "trafficLimitStrategy"),
		"can_topup_traffic":          active && limitBytes > 0,
		"can_topup_regular_traffic":  active && limitBytes > 0,
		"can_topup_premium_traffic":  false,
		"premium_unlimited_override": override.PremiumUnlimited,
		"premium_bonus_bytes":        override.PremiumBonusBytes,
		"premium_used_bytes":         int64(0),
		"premium_limit_bytes":        premiumLimitBytes(override),
		"premium_is_limited":         false,
		"regular_unlimited_override": override.RegularUnlimited,
		"regular_bonus_bytes":        override.RegularBonusBytes,
		"auto_renew_enabled":         userAutoRenewEnabled(ctx, pool, user.UserID),
		"auto_renew_available":       true,
		"tariff_key":                 plan.TariffKey,
		"tariff_name":                tariffName,
		"billing_model":              plan.BillingModel,
		"tariff_description":         plan.Description,
		"provider":                   plan.Provider,
	}
	if onlineAt := parsePanelTime(traffic["onlineAt"]); !onlineAt.IsZero() {
		payload["online_at"] = onlineAt
	}
	return payload
}

func latestPaidPlanForUser(ctx context.Context, pool *pgxpool.Pool, userID int64) paidPlan {
	if pool == nil {
		return paidPlan{}
	}
	var order payments.Order
	err := pool.QueryRow(ctx, `
SELECT COALESCE(provider,''), COALESCE(plan_snapshot,'{}'::jsonb), COALESCE(tariff_key,''), COALESCE(sale_mode,''),
	COALESCE(months,0), COALESCE(traffic_gb,0)::float8
FROM payment_orders
WHERE user_id=$1 AND status IN ('paid','succeeded')
ORDER BY paid_at DESC NULLS LAST, created_at DESC
LIMIT 1`, userID).Scan(&order.Provider, &order.PlanSnapshot, &order.TariffKey, &order.SaleMode, &order.Months, &order.TrafficGB)
	if err != nil {
		return paidPlan{}
	}
	return paidPlanFromOrder(order)
}

func paidPlanFromOrder(order payments.Order) paidPlan {
	plan := paidPlan{}
	if len(order.PlanSnapshot) > 0 {
		_ = json.Unmarshal(order.PlanSnapshot, &plan)
	}
	if plan.TariffKey == "" {
		plan.TariffKey = order.TariffKey
	}
	if plan.SaleMode == "" {
		plan.SaleMode = order.SaleMode
	}
	if plan.Months == 0 {
		plan.Months = order.Months
	}
	if plan.TrafficGB == 0 {
		plan.TrafficGB = order.TrafficGB
	}
	if plan.BillingModel == "" {
		if plan.SaleMode == "traffic_package" {
			plan.BillingModel = "traffic"
		} else {
			plan.BillingModel = "period"
		}
	}
	if plan.Title != "" && plan.TariffName == "" {
		plan.TariffName = plan.Title
	}
	plan.Provider = order.Provider
	return plan
}

func provisionPaidOrder(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client, order payments.Order) error {
	if order.ProvisionedAt != nil || !paymentOrderPaid(order) {
		return nil
	}
	if panel == nil || !panel.Configured(ctx) {
		err := remnawave.ErrNotConfigured
		markPaymentProvisionError(ctx, pool, order.PaymentID, err)
		return err
	}
	user, err := loadWebappUser(ctx, pool, order.UserID, settings)
	if err != nil {
		markPaymentProvisionError(ctx, pool, order.PaymentID, err)
		return err
	}
	plan := paidPlanFromOrder(order)
	panelUser, err := ensurePanelUserForPlan(ctx, pool, panel, user, plan)
	if err != nil {
		markPaymentProvisionError(ctx, pool, order.PaymentID, err)
		return err
	}
	if uuid := stringValue(panelUser, "uuid"); uuid != "" {
		bindPanelUUID(ctx, pool, user.UserID, uuid)
	}
	_, err = pool.Exec(ctx, "UPDATE payment_orders SET provisioned_at=NOW(), provision_error=NULL, updated_at=NOW() WHERE payment_id=$1 AND provisioned_at IS NULL", order.PaymentID)
	return err
}

// ProvisionPendingPaidOrders provisions paid orders that have not yet been
// applied to Remnawave. It is safe to call from webhooks and workers; failures
// are recorded on each order and returned as a joined error for logging.
func ProvisionPendingPaidOrders(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client, limit int) (ProvisionResult, error) {
	result := ProvisionResult{}
	if pool == nil {
		return result, fmt.Errorf("database_not_configured")
	}
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	orders, err := loadPendingProvisionOrders(ctx, pool, limit)
	if err != nil {
		return result, err
	}
	result.Scanned = len(orders)
	var joined error
	for _, order := range orders {
		lock, locked, err := tryPaymentProvisionLock(ctx, pool, order.PaymentID)
		if err != nil {
			result.Failed++
			joined = errors.Join(joined, fmt.Errorf("payment_id=%d: %w", order.PaymentID, err))
			continue
		}
		if !locked {
			continue
		}
		if err := provisionPaidOrder(ctx, settings, pool, panel, order); err != nil {
			unlockPaymentProvisionLock(ctx, lock, order.PaymentID)
			result.Failed++
			joined = errors.Join(joined, fmt.Errorf("payment_id=%d: %w", order.PaymentID, err))
			continue
		}
		unlockPaymentProvisionLock(ctx, lock, order.PaymentID)
		result.Provisioned++
	}
	return result, joined
}

func tryPaymentProvisionLock(ctx context.Context, pool *pgxpool.Pool, paymentID int64) (*pgxpool.Conn, bool, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, false, err
	}
	var locked bool
	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", paymentProvisionLockKey(paymentID)).Scan(&locked); err != nil {
		conn.Release()
		return nil, false, err
	}
	if !locked {
		conn.Release()
		return nil, false, nil
	}
	return conn, true, nil
}

func unlockPaymentProvisionLock(_ context.Context, conn *pgxpool.Conn, paymentID int64) {
	if conn == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var unlocked bool
	if err := conn.QueryRow(ctx, "SELECT pg_advisory_unlock($1)", paymentProvisionLockKey(paymentID)).Scan(&unlocked); err != nil || !unlocked {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer closeCancel()
		_ = conn.Hijack().Close(closeCtx)
		return
	}
	conn.Release()
}

func paymentProvisionLockKey(paymentID int64) int64 {
	return paymentProvisionLockNamespace ^ paymentID
}

// RunPanelSync reconciles local users with Remnawave and provisions paid orders.
func RunPanelSync(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client, limit int) (PanelSyncResult, error) {
	result := PanelSyncResult{
		Status:       "ok",
		LastSyncTime: time.Now().UTC().Format(time.RFC3339),
	}
	if pool == nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, "database_not_configured")
		return result, fmt.Errorf("database_not_configured")
	}
	if limit <= 0 || limit > 2000 {
		limit = 500
	}

	provision, err := ProvisionPendingPaidOrders(ctx, settings, pool, panel, 100)
	result.PaymentsScanned = provision.Scanned
	result.PaymentsProvisioned = provision.Provisioned
	result.PaymentsFailed = provision.Failed
	if err != nil {
		result.Status = "partial"
		result.Errors = appendSyncError(result.Errors, err)
	}

	if panel == nil || !panel.Configured(ctx) {
		if result.Status == "ok" {
			result.Status = "partial"
		}
		result.Errors = append(result.Errors, "panel_not_configured")
		_ = savePanelSyncStatus(ctx, pool, result)
		return result, nil
	}

	users, err := loadPanelSyncUsers(ctx, pool, settings, limit)
	if err != nil {
		result.Status = "failed"
		result.Errors = appendSyncError(result.Errors, err)
		_ = savePanelSyncStatus(ctx, pool, result)
		return result, err
	}
	for _, user := range users {
		result.UsersProcessed++
		panelUser, found, err := panelUserForWebUser(ctx, pool, panel, user)
		if err != nil {
			result.Status = "partial"
			result.Errors = appendSyncError(result.Errors, err)
			continue
		}
		if found && stringValue(panelUser, "uuid") != "" {
			result.SubscriptionsSynced++
			cachePanelUserState(ctx, pool, user.UserID, panelUser)
		} else {
			_, _ = pool.Exec(ctx, "UPDATE users SET panel_status=NULL,panel_expire_at=NULL,panel_state_synced_at=NOW() WHERE user_id=$1", user.UserID)
		}
	}
	if err := savePanelSyncStatus(ctx, pool, result); err != nil {
		result.Status = "partial"
		result.Errors = appendSyncError(result.Errors, err)
	}
	return result, nil
}

func cachePanelUserState(ctx context.Context, pool *pgxpool.Pool, userID int64, panelUser map[string]any) {
	if pool == nil || userID == 0 {
		return
	}
	traffic := mapValue(panelUser, "userTraffic")
	used := int64Value(traffic, "lifetimeUsedTrafficBytes")
	if used == 0 {
		used = int64Value(traffic, "usedTrafficBytes")
	}
	status := strings.ToUpper(strings.TrimSpace(stringValue(panelUser, "status")))
	expireAt := parsePanelTime(panelUser["expireAt"])
	var expireValue any
	if !expireAt.IsZero() {
		expireValue = expireAt
	}
	_, _ = pool.Exec(ctx, `UPDATE users SET panel_status=$2,panel_expire_at=$3,panel_state_synced_at=NOW(),
lifetime_used_traffic_bytes=$4,lifetime_used_traffic_synced_at=NOW() WHERE user_id=$1`, userID, emptyStringToNil(status), expireValue, used)
}

// LastPanelSyncStatus returns the most recent sync status payload for admin stats.
func LastPanelSyncStatus(ctx context.Context, pool *pgxpool.Pool) PanelSyncResult {
	result := PanelSyncResult{Status: "idle"}
	if pool == nil {
		return result
	}
	raw, ok, err := appsettings.NewStore(pool).Get(ctx, panelSyncStatusSettingKey)
	if err != nil || !ok {
		return result
	}
	if json.Unmarshal(raw, &result) != nil || result.Status == "" {
		return PanelSyncResult{Status: "idle"}
	}
	return result
}

func loadPanelSyncUsers(ctx context.Context, pool *pgxpool.Pool, settings config.Settings, limit int) ([]webappUser, error) {
	rows, err := pool.Query(ctx, `
SELECT user_id, COALESCE(telegram_id,0), COALESCE(username,''), COALESCE(email,''), COALESCE(first_name,''), COALESCE(last_name,''),
	COALESCE(language_code,''), COALESCE(telegram_photo_url,''), COALESCE(panel_user_uuid,'')
FROM users
WHERE COALESCE(panel_user_uuid,'') = '' OR COALESCE(telegram_id,0) <> 0 OR COALESCE(email,'') <> ''
ORDER BY CASE WHEN COALESCE(panel_user_uuid,'') = '' THEN 0 ELSE 1 END, registration_date DESC
LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []webappUser{}
	for rows.Next() {
		var user webappUser
		if err := rows.Scan(&user.UserID, &user.TelegramID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.LanguageCode, &user.PhotoURL, &user.PanelUserUUID); err != nil {
			return nil, err
		}
		user.LanguageCode = normalizeWebLanguage(user.LanguageCode, effectiveDefaultLanguage(ctx, pool, settings))
		user.IsAdmin = isAdminID(settings.AdminIDs, user.UserID) || isAdminID(settings.AdminIDs, user.TelegramID)
		users = append(users, user)
	}
	return users, rows.Err()
}

func savePanelSyncStatus(ctx context.Context, pool *pgxpool.Pool, result PanelSyncResult) error {
	// Skip write when status hasn't changed to avoid redundant app_settings rows.
	prev := LastPanelSyncStatus(ctx, pool)
	if prev.Status == result.Status &&
		prev.UsersProcessed == result.UsersProcessed &&
		prev.SubscriptionsSynced == result.SubscriptionsSynced &&
		prev.PaymentsScanned == result.PaymentsScanned &&
		prev.PaymentsProvisioned == result.PaymentsProvisioned &&
		prev.PaymentsFailed == result.PaymentsFailed &&
		prev.LastSyncTime != "" && result.LastSyncTime != "" {
		// Only update the timestamp without touching the rest.
		result = prev
		result.LastSyncTime = time.Now().UTC().Format(time.RFC3339)
	}
	return appsettings.NewStore(pool).Upsert(ctx, panelSyncStatusSettingKey, result)
}

func appendSyncError(errorsList []string, err error) []string {
	if err == nil || len(errorsList) >= 10 {
		return errorsList
	}
	message := err.Error()
	if len(message) > 300 {
		message = message[:300]
	}
	return append(errorsList, message)
}

func loadPendingProvisionOrders(ctx context.Context, pool *pgxpool.Pool, limit int) ([]payments.Order, error) {
	rows, err := pool.Query(ctx, `
SELECT payment_id, order_id, user_id, COALESCE(provider,''), COALESCE(method,''), COALESCE(payment_type,''),
	amount::float8, COALESCE(currency,''), COALESCE(base_amount, amount)::float8, COALESCE(base_currency, currency),
	COALESCE(display_cny_amount,0)::float8, COALESCE(fx_rate,0)::float8, COALESCE(fx_source,''), fx_updated_at,
	COALESCE(plan_hash,''), COALESCE(plan_snapshot,'{}'::jsonb), COALESCE(status,''), COALESCE(description,''),
	COALESCE(tariff_key,''), COALESCE(sale_mode,''), COALESCE(months,0), COALESCE(traffic_gb,0)::float8,
	COALESCE(device_count,0), COALESCE(provider_payment_id,''), COALESCE(payment_url,''), COALESCE(qr_content,''),
	COALESCE(display_amount,''), COALESCE(display_currency,''), COALESCE(payment_address,''), COALESCE(network,''),
	COALESCE(url_scheme,''), COALESCE(raw_webhook,'{}'::jsonb), provisioned_at, COALESCE(provision_error,''), created_at, updated_at, paid_at
FROM payment_orders
WHERE status IN ('paid','succeeded') AND provisioned_at IS NULL
ORDER BY paid_at ASC NULLS LAST, updated_at ASC
LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders := []payments.Order{}
	for rows.Next() {
		var order payments.Order
		var deviceCount int
		if err := rows.Scan(
			&order.PaymentID, &order.OrderID, &order.UserID, &order.Provider, &order.Method, &order.PaymentType,
			&order.Amount, &order.Currency, &order.BaseAmount, &order.BaseCurrency, &order.DisplayCNYAmount,
			&order.FXRate, &order.FXSource, &order.FXUpdatedAt, &order.PlanHash, &order.PlanSnapshot,
			&order.Status, &order.Description, &order.TariffKey, &order.SaleMode, &order.Months, &order.TrafficGB,
			&deviceCount, &order.ProviderPaymentID, &order.PaymentURL, &order.QRContent, &order.DisplayAmount,
			&order.DisplayCurrency, &order.PaymentAddress, &order.Network, &order.URLScheme, &order.RawWebhook,
			&order.ProvisionedAt, &order.ProvisionError, &order.CreatedAt, &order.UpdatedAt, &order.PaidAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func ensurePanelUserForPlan(ctx context.Context, pool *pgxpool.Pool, panel *remnawave.Client, user webappUser, plan paidPlan) (map[string]any, error) {
	cfg := panel.EffectiveConfig(ctx)
	existing, found, err := panelUserForWebUser(ctx, pool, panel, user)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	existingExpire := time.Time{}
	currentTrafficLimit := int64(0)
	if found {
		existingExpire = parsePanelTime(existing["expireAt"])
		currentTrafficLimit = int64Value(existing, "trafficLimitBytes")
	}
	expireAt := nextExpireAt(plan, existingExpire, now)
	trafficLimit := trafficLimitBytesForPlan(plan, cfg, currentTrafficLimit)
	payload := map[string]any{
		"status":               "ACTIVE",
		"expireAt":             expireAt.Format(time.RFC3339Nano),
		"trafficLimitBytes":    trafficLimit,
		"trafficLimitStrategy": cfg.UserTrafficStrategy,
		"description":          panelDescription(user),
	}
	if user.TelegramID != 0 {
		payload["telegramId"] = user.TelegramID
	}
	if user.Email != "" {
		payload["email"] = user.Email
	}
	if squads := effectiveSquads(plan, cfg); len(squads) > 0 {
		payload["activeInternalSquads"] = squads
	}
	if external := firstNonEmpty(plan.ExternalSquadUUID, cfg.UserExternalSquadUUID); external != "" {
		payload["externalSquadUuid"] = external
	}

	if found {
		uuid := stringValue(existing, "uuid")
		payload["uuid"] = uuid
		return panel.UpdateUser(ctx, payload)
	}
	payload["username"] = panelUsername(user)
	return panel.CreateUser(ctx, payload)
}

func grantPanelAccessDays(ctx context.Context, pool *pgxpool.Pool, panel *remnawave.Client, user webappUser, days int, options accessGrantOptions) (map[string]any, error) {
	if panel == nil || !panel.Configured(ctx) {
		return nil, remnawave.ErrNotConfigured
	}
	if days <= 0 {
		return nil, fmt.Errorf("invalid_days")
	}
	cfg := panel.EffectiveConfig(ctx)
	existing, found, err := panelUserForWebUser(ctx, pool, panel, user)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	base := now
	currentTrafficLimit := int64(0)
	if found {
		if currentExpire := parsePanelTime(existing["expireAt"]); currentExpire.After(now) {
			base = currentExpire
		}
		currentTrafficLimit = int64Value(existing, "trafficLimitBytes")
	}
	trafficLimit := currentTrafficLimit
	if options.SetTrafficLimit || trafficLimit == 0 {
		trafficGB := options.TrafficLimitGB
		if trafficGB <= 0 {
			trafficGB = cfg.UserTrafficLimitGB
		}
		trafficLimit = gbToBytes(trafficGB)
	}
	payload := map[string]any{
		"status":               "ACTIVE",
		"expireAt":             base.AddDate(0, 0, days).Format(time.RFC3339Nano),
		"trafficLimitStrategy": firstNonEmpty(options.TrafficStrategy, cfg.UserTrafficStrategy),
		"description":          panelDescription(user),
	}
	if trafficLimit > 0 || options.SetTrafficLimit {
		payload["trafficLimitBytes"] = trafficLimit
	}
	if user.TelegramID != 0 {
		payload["telegramId"] = user.TelegramID
	}
	if user.Email != "" {
		payload["email"] = user.Email
	}
	squads := cleanStringSlice(options.SquadUUIDs)
	if len(squads) == 0 {
		squads = cleanStringSlice(cfg.UserSquadUUIDs)
	}
	if len(squads) > 0 {
		payload["activeInternalSquads"] = squads
	}
	if cfg.UserExternalSquadUUID != "" {
		payload["externalSquadUuid"] = cfg.UserExternalSquadUUID
	}
	if source := strings.TrimSpace(options.Source); source != "" {
		payload["description"] = strings.TrimSpace(panelDescription(user) + " [" + source + "]")
	}
	if found {
		payload["uuid"] = stringValue(existing, "uuid")
		updated, err := panel.UpdateUser(ctx, payload)
		if err != nil {
			return nil, err
		}
		return updated, nil
	}
	payload["username"] = panelUsername(user)
	created, err := panel.CreateUser(ctx, payload)
	if err != nil {
		return nil, err
	}
	if uuid := stringValue(created, "uuid"); uuid != "" {
		bindPanelUUID(ctx, pool, user.UserID, uuid)
	}
	return created, nil
}

func nextExpireAt(plan paidPlan, existing time.Time, now time.Time) time.Time {
	if plan.SaleMode == "traffic_package" || strings.EqualFold(plan.BillingModel, "traffic") {
		if existing.After(now) {
			return existing
		}
		return now.AddDate(10, 0, 0)
	}
	base := now
	if existing.After(now) {
		base = existing
	}
	months := plan.Months
	if months <= 0 {
		months = 1
	}
	return base.AddDate(0, months, 0)
}

func trafficLimitBytesForPlan(plan paidPlan, cfg remnawave.EffectiveConfig, current int64) int64 {
	if plan.SaleMode == "traffic_package" || strings.EqualFold(plan.BillingModel, "traffic") {
		added := gbToBytes(plan.TrafficGB)
		if added > 0 {
			return current + added
		}
	}
	if plan.MonthlyGB > 0 {
		return gbToBytes(plan.MonthlyGB)
	}
	return gbToBytes(cfg.UserTrafficLimitGB)
}

func effectiveSquads(plan paidPlan, cfg remnawave.EffectiveConfig) []string {
	if len(plan.SquadUUIDs) > 0 {
		return cleanStringSlice(plan.SquadUUIDs)
	}
	return cleanStringSlice(cfg.UserSquadUUIDs)
}

func paymentOrderPaid(order payments.Order) bool {
	status := strings.ToLower(strings.TrimSpace(order.Status))
	return status == "paid" || status == "succeeded"
}

func markPaymentProvisionError(ctx context.Context, pool *pgxpool.Pool, paymentID int64, err error) {
	if pool == nil || err == nil {
		return
	}
	msg := err.Error()
	if len(msg) > 1000 {
		msg = msg[:1000]
	}
	_, _ = pool.Exec(ctx, "UPDATE payment_orders SET provision_error=$2, updated_at=NOW() WHERE payment_id=$1", paymentID, msg)
}

func bindPanelUUID(ctx context.Context, pool *pgxpool.Pool, userID int64, uuid string) {
	uuid = strings.TrimSpace(uuid)
	if pool == nil || userID == 0 || uuid == "" {
		return
	}
	_, _ = pool.Exec(ctx, "UPDATE users SET panel_user_uuid=$2 WHERE user_id=$1 AND (panel_user_uuid IS NULL OR panel_user_uuid <> $2)", userID, uuid)
}

func panelUsername(user webappUser) string {
	if user.TelegramID != 0 {
		return "tg_" + strconv.FormatInt(user.TelegramID, 10)
	}
	base := "u_" + strconv.FormatInt(user.UserID, 10)
	if user.Email != "" {
		base = "em_" + sanitizePanelUsername(strings.Split(user.Email, "@")[0])
	}
	if len(base) < 3 {
		base += "_user"
	}
	if len(base) > 36 {
		base = base[:36]
	}
	return base
}

func sanitizePanelUsername(value string) string {
	value = panelUsernameInvalid.ReplaceAllString(strings.TrimSpace(value), "_")
	value = strings.Trim(value, "_-")
	if value == "" {
		return "user"
	}
	return value
}

func panelDescription(user webappUser) string {
	parts := []string{}
	if user.Username != "" {
		parts = append(parts, "@"+strings.TrimPrefix(user.Username, "@"))
	}
	name := strings.TrimSpace(strings.Join([]string{user.FirstName, user.LastName}, " "))
	if name != "" {
		parts = append(parts, name)
	}
	if user.Email != "" {
		parts = append(parts, user.Email)
	}
	if len(parts) == 0 {
		return fmt.Sprintf("Remna User Panel user %d", user.UserID)
	}
	return strings.Join(parts, " | ")
}

func loadPanelUUIDForUser(ctx context.Context, pool *pgxpool.Pool, userID int64) string {
	if pool == nil {
		return ""
	}
	var uuid string
	_ = pool.QueryRow(ctx, "SELECT COALESCE(panel_user_uuid,'') FROM users WHERE user_id=$1", userID).Scan(&uuid)
	return uuid
}

func loadUserTrafficOverride(ctx context.Context, pool *pgxpool.Pool, userID int64) userTrafficOverride {
	overrides := readUserTrafficOverrides(ctx, pool)
	return overrides[strconv.FormatInt(userID, 10)]
}

func saveUserTrafficOverride(ctx context.Context, pool *pgxpool.Pool, userID int64, override userTrafficOverride) error {
	if pool == nil {
		return fmt.Errorf("database_not_configured")
	}
	overrides := readUserTrafficOverrides(ctx, pool)
	override.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	overrides[strconv.FormatInt(userID, 10)] = override
	return appsettings.NewStore(pool).Upsert(ctx, userTrafficOverridesSettingKey, overrides)
}

func readUserTrafficOverrides(ctx context.Context, pool *pgxpool.Pool) map[string]userTrafficOverride {
	raw, ok, err := appsettings.NewStore(pool).Get(ctx, userTrafficOverridesSettingKey)
	if err != nil || !ok {
		return map[string]userTrafficOverride{}
	}
	var overrides map[string]userTrafficOverride
	if json.Unmarshal(raw, &overrides) == nil && overrides != nil {
		return overrides
	}
	return map[string]userTrafficOverride{}
}

func premiumLimitBytes(override userTrafficOverride) int64 {
	if override.PremiumUnlimited {
		return 0
	}
	if override.PremiumBonusBytes > 0 {
		return override.PremiumBonusBytes
	}
	return 0
}

func userAutoRenewEnabled(ctx context.Context, pool *pgxpool.Pool, userID int64) bool {
	raw, ok, err := appsettings.NewStore(pool).Get(ctx, userAutoRenewSettingKey)
	if err != nil || !ok {
		return false
	}
	var values map[string]bool
	if json.Unmarshal(raw, &values) != nil || values == nil {
		return false
	}
	return values[strconv.FormatInt(userID, 10)]
}

func saveUserAutoRenew(ctx context.Context, pool *pgxpool.Pool, userID int64, enabled bool) error {
	if pool == nil {
		return fmt.Errorf("database_not_configured")
	}
	store := appsettings.NewStore(pool)
	raw, ok, err := store.Get(ctx, userAutoRenewSettingKey)
	if err != nil {
		return err
	}
	values := map[string]bool{}
	if ok {
		_ = json.Unmarshal(raw, &values)
		if values == nil {
			values = map[string]bool{}
		}
	}
	key := strconv.FormatInt(userID, 10)
	if enabled {
		values[key] = true
	} else {
		delete(values, key)
	}
	return store.Upsert(ctx, userAutoRenewSettingKey, values)
}

func loadUserNotificationPrefs(ctx context.Context, pool *pgxpool.Pool, userID int64) UserNotificationPrefs {
	prefs := UserNotificationPrefs{
		ExpiryEnabled:       true,
		ExpiryDaysBefore:    defaultExpiryDaysBefore,
		TrafficEnabled:      true,
		TrafficThresholdPct: defaultTrafficThresholdPct,
	}
	if pool == nil {
		return prefs
	}
	store := appsettings.NewStore(pool)
	userKey := strconv.FormatInt(userID, 10)

	// Expiry enabled
	if raw, ok, _ := store.Get(ctx, userNotifyExpiryEnabledKey); ok {
		var values map[string]bool
		if json.Unmarshal(raw, &values) == nil && values != nil {
			if v, exists := values[userKey]; exists {
				prefs.ExpiryEnabled = v
			}
		}
	}
	// Expiry days before
	if raw, ok, _ := store.Get(ctx, userNotifyExpiryDaysBeforeKey); ok {
		var values map[string]float64
		if json.Unmarshal(raw, &values) == nil && values != nil {
			if v, exists := values[userKey]; exists && v > 0 {
				prefs.ExpiryDaysBefore = int(v)
			}
		}
	}
	// Traffic enabled
	if raw, ok, _ := store.Get(ctx, userNotifyTrafficEnabledKey); ok {
		var values map[string]bool
		if json.Unmarshal(raw, &values) == nil && values != nil {
			if v, exists := values[userKey]; exists {
				prefs.TrafficEnabled = v
			}
		}
	}
	// Traffic threshold pct
	if raw, ok, _ := store.Get(ctx, userNotifyTrafficThresholdPctKey); ok {
		var values map[string]float64
		if json.Unmarshal(raw, &values) == nil && values != nil {
			if v, exists := values[userKey]; exists && v > 0 && v <= 100 {
				prefs.TrafficThresholdPct = int(v)
			}
		}
	}
	return prefs
}

func saveUserNotificationPrefs(ctx context.Context, pool *pgxpool.Pool, userID int64, prefs UserNotificationPrefs) error {
	if pool == nil {
		return fmt.Errorf("database_not_configured")
	}
	store := appsettings.NewStore(pool)
	userKey := strconv.FormatInt(userID, 10)

	// Expiry enabled
	if err := saveUserBoolMapSetting(ctx, store, userNotifyExpiryEnabledKey, userKey, prefs.ExpiryEnabled); err != nil {
		return err
	}
	// Expiry days before
	if err := saveUserFloatMapSetting(ctx, store, userNotifyExpiryDaysBeforeKey, userKey, float64(prefs.ExpiryDaysBefore)); err != nil {
		return err
	}
	// Traffic enabled
	if err := saveUserBoolMapSetting(ctx, store, userNotifyTrafficEnabledKey, userKey, prefs.TrafficEnabled); err != nil {
		return err
	}
	// Traffic threshold pct
	if err := saveUserFloatMapSetting(ctx, store, userNotifyTrafficThresholdPctKey, userKey, float64(prefs.TrafficThresholdPct)); err != nil {
		return err
	}
	return nil
}

func saveUserBoolMapSetting(ctx context.Context, store appsettings.Store, settingKey, userKey string, value bool) error {
	raw, ok, _ := store.Get(ctx, settingKey)
	values := map[string]bool{}
	if ok {
		_ = json.Unmarshal(raw, &values)
		if values == nil {
			values = map[string]bool{}
		}
	}
	values[userKey] = value
	return store.Upsert(ctx, settingKey, values)
}

func saveUserFloatMapSetting(ctx context.Context, store appsettings.Store, settingKey, userKey string, value float64) error {
	raw, ok, _ := store.Get(ctx, settingKey)
	values := map[string]float64{}
	if ok {
		_ = json.Unmarshal(raw, &values)
		if values == nil {
			values = map[string]float64{}
		}
	}
	values[userKey] = value
	return store.Upsert(ctx, settingKey, values)
}

func loadRecentPaymentsForUser(ctx context.Context, pool *pgxpool.Pool, userID int64) []payments.Order {
	if pool == nil {
		return []payments.Order{}
	}
	rows, err := pool.Query(ctx, `
SELECT payment_id, order_id, user_id, COALESCE(provider,''), COALESCE(method,''), COALESCE(payment_type,''),
	amount::float8, COALESCE(currency,''), COALESCE(base_amount, amount)::float8, COALESCE(base_currency, currency),
	COALESCE(display_cny_amount,0)::float8, COALESCE(fx_rate,0)::float8, COALESCE(fx_source,''), fx_updated_at,
	COALESCE(plan_hash,''), COALESCE(plan_snapshot,'{}'::jsonb), COALESCE(status,''), COALESCE(description,''),
	COALESCE(tariff_key,''), COALESCE(sale_mode,''), COALESCE(months,0), COALESCE(traffic_gb,0)::float8,
	COALESCE(device_count,0), COALESCE(provider_payment_id,''), COALESCE(payment_url,''), COALESCE(qr_content,''),
	COALESCE(display_amount,''), COALESCE(display_currency,''), COALESCE(payment_address,''), COALESCE(network,''),
	COALESCE(url_scheme,''), COALESCE(raw_webhook,'{}'::jsonb), provisioned_at, COALESCE(provision_error,''), created_at, updated_at, paid_at
FROM payment_orders WHERE user_id=$1 ORDER BY created_at DESC LIMIT 20`, userID)
	if err != nil {
		return []payments.Order{}
	}
	defer rows.Close()
	result := []payments.Order{}
	for rows.Next() {
		var order payments.Order
		var deviceCount int
		if err := rows.Scan(
			&order.PaymentID, &order.OrderID, &order.UserID, &order.Provider, &order.Method, &order.PaymentType,
			&order.Amount, &order.Currency, &order.BaseAmount, &order.BaseCurrency, &order.DisplayCNYAmount,
			&order.FXRate, &order.FXSource, &order.FXUpdatedAt, &order.PlanHash, &order.PlanSnapshot,
			&order.Status, &order.Description, &order.TariffKey, &order.SaleMode, &order.Months, &order.TrafficGB,
			&deviceCount, &order.ProviderPaymentID, &order.PaymentURL, &order.QRContent, &order.DisplayAmount,
			&order.DisplayCurrency, &order.PaymentAddress, &order.Network, &order.URLScheme, &order.RawWebhook,
			&order.ProvisionedAt, &order.ProvisionError, &order.CreatedAt, &order.UpdatedAt, &order.PaidAt,
		); err == nil {
			result = append(result, order)
		}
	}
	return result
}

func userTotalPaid(ctx context.Context, pool *pgxpool.Pool, userID int64) float64 {
	if pool == nil {
		return 0
	}
	var total float64
	_ = pool.QueryRow(ctx, "SELECT COALESCE(SUM(base_amount),0)::float8 FROM payment_orders WHERE user_id=$1 AND status IN ('paid','succeeded')", userID).Scan(&total)
	return total
}

func panelAwareAdminUser(ctx context.Context, pool *pgxpool.Pool, panel *remnawave.Client, user map[string]any) map[string]any {
	userID, _ := strconv.ParseInt(fmt.Sprint(user["user_id"]), 10, 64)
	override := loadUserTrafficOverride(ctx, pool, userID)
	user["premium_traffic"] = premiumTrafficPayload(override)
	panelUUID := loadPanelUUIDForUser(ctx, pool, userID)
	if panelUUID != "" {
		user["panel_user_uuid"] = panelUUID
	}
	if panel == nil || !panel.Configured(ctx) || panelUUID == "" {
		return user
	}
	panelUser, ok, err := panel.GetUserByUUID(ctx, panelUUID)
	if err != nil || !ok {
		user["panel_status"] = "unknown"
		return user
	}
	enrichAdminUserFromPanel(user, panelUser, override)
	return user
}

func enrichAdminUserFromPanel(user map[string]any, panelUser map[string]any, override userTrafficOverride) {
	status := strings.ToLower(stringValue(panelUser, "status"))
	if status == "" {
		status = "unknown"
	}
	user["panel_status"] = status
	user["subscription_expires_at"] = timeString(parsePanelTime(panelUser["expireAt"]))
	user["panel_status_expired_at"] = user["subscription_expires_at"]
	traffic := mapValue(panelUser, "userTraffic")
	used := int64Value(traffic, "usedTrafficBytes")
	limit := int64Value(panelUser, "trafficLimitBytes")
	user["traffic_used_bytes"] = used
	user["traffic_limit_bytes"] = limit
	user["premium_traffic"] = premiumTrafficPayload(override)
}

func premiumTrafficPayload(override userTrafficOverride) map[string]any {
	if override.PremiumUnlimited {
		return map[string]any{
			"state":       "unlimited",
			"used_bytes":  int64(0),
			"limit_bytes": int64(0),
		}
	}
	if override.PremiumBonusBytes > 0 {
		return map[string]any{
			"state":       "good",
			"used_bytes":  int64(0),
			"limit_bytes": override.PremiumBonusBytes,
		}
	}
	return map[string]any{"state": "none"}
}

func parsePanelTime(value any) time.Time {
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "" || text == "<nil>" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05-07"} {
		if parsed, err := time.Parse(layout, text); err == nil {
			return parsed.UTC()
		}
	}
	return time.Time{}
}

func timeString(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func dateText(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format("2006-01-02")
}

func remainingText(days int) string {
	if days <= 0 {
		return ""
	}
	return strconv.Itoa(days) + " d."
}

func bytesText(bytes int64) string {
	if bytes <= 0 {
		return "0 GB"
	}
	gb := float64(bytes) / bytesPerGB
	return compactFloat(math.Round(gb*10)/10) + " GB"
}

func gbToBytes(gb float64) int64 {
	if gb <= 0 {
		return 0
	}
	return int64(math.Round(gb * bytesPerGB))
}

func stringValue(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	value, ok := m[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func int64Value(m map[string]any, key string) int64 {
	if m == nil {
		return 0
	}
	switch value := m[key].(type) {
	case int:
		return int64(value)
	case int64:
		return value
	case float64:
		return int64(value)
	case json.Number:
		parsed, _ := value.Int64()
		return parsed
	case string:
		parsed, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		return parsed
	default:
		return 0
	}
}

func mapValue(m map[string]any, key string) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	if value, ok := m[key].(map[string]any); ok {
		return value
	}
	return map[string]any{}
}

func cleanStringSlice(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func panelErrorCode(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, remnawave.ErrNotConfigured):
		return "panel_not_configured"
	case errors.Is(err, remnawave.ErrNotFound):
		return "panel_user_not_found"
	default:
		return "panel_request_failed"
	}
}

func adminExtendPanelUser(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client, userID int64, r *http.Request) error {
	if panel == nil || !panel.Configured(ctx) {
		return remnawave.ErrNotConfigured
	}
	var payload struct {
		Days      int    `json:"days"`
		TariffKey string `json:"tariff_key"`
	}
	if err := decodeJSONBody(r, &payload); err != nil {
		return err
	}
	if payload.Days <= 0 {
		return fmt.Errorf("invalid_days")
	}
	user, panelUser, err := adminPanelUser(ctx, settings, pool, panel, userID)
	if err != nil {
		return err
	}
	currentExpire := parsePanelTime(panelUser["expireAt"])
	base := time.Now().UTC()
	if currentExpire.After(base) {
		base = currentExpire
	}
	update := map[string]any{
		"uuid":     stringValue(panelUser, "uuid"),
		"status":   "ACTIVE",
		"expireAt": base.AddDate(0, 0, payload.Days).Format(time.RFC3339Nano),
	}
	if payload.TariffKey != "" {
		if plan, ok := firstPlanForTariff(ctx, settings, user.LanguageCode, payload.TariffKey); ok {
			mergePlanUpdate(update, plan, panel.EffectiveConfig(ctx))
		}
	}
	_, err = panel.UpdateUser(ctx, update)
	return err
}

func adminChangePanelTariff(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client, userID int64, r *http.Request) error {
	if panel == nil || !panel.Configured(ctx) {
		return remnawave.ErrNotConfigured
	}
	var payload struct {
		TariffKey string `json:"tariff_key"`
	}
	if err := decodeJSONBody(r, &payload); err != nil {
		return err
	}
	if strings.TrimSpace(payload.TariffKey) == "" {
		return fmt.Errorf("tariff_key_required")
	}
	user, panelUser, err := adminPanelUser(ctx, settings, pool, panel, userID)
	if err != nil {
		return err
	}
	plan, ok := firstPlanForTariff(ctx, settings, user.LanguageCode, payload.TariffKey)
	if !ok {
		return fmt.Errorf("tariff_not_found")
	}
	update := map[string]any{"uuid": stringValue(panelUser, "uuid"), "status": "ACTIVE"}
	mergePlanUpdate(update, plan, panel.EffectiveConfig(ctx))
	_, err = panel.UpdateUser(ctx, update)
	return err
}

func adminSetPremiumTrafficOverride(ctx context.Context, pool *pgxpool.Pool, userID int64, r *http.Request) error {
	var payload struct {
		Unlimited  bool    `json:"unlimited"`
		BonusGB    float64 `json:"bonus_gb"`
		BonusBytes *int64  `json:"bonus_bytes"`
	}
	if err := decodeJSONBody(r, &payload); err != nil {
		return err
	}
	bonusBytes := gbToBytes(payload.BonusGB)
	if payload.BonusBytes != nil {
		bonusBytes = *payload.BonusBytes
	}
	if bonusBytes < 0 {
		return fmt.Errorf("invalid_bonus")
	}
	override := loadUserTrafficOverride(ctx, pool, userID)
	override.PremiumUnlimited = payload.Unlimited
	override.PremiumBonusBytes = bonusBytes
	if override.PremiumUnlimited {
		override.PremiumBonusBytes = 0
	}
	return saveUserTrafficOverride(ctx, pool, userID, override)
}

func adminSetRegularTrafficOverride(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client, userID int64, r *http.Request) error {
	if panel == nil || !panel.Configured(ctx) {
		return remnawave.ErrNotConfigured
	}
	var payload struct {
		Unlimited         bool    `json:"unlimited"`
		RegularBonusGB    float64 `json:"regular_bonus_gb"`
		RegularBonusBytes *int64  `json:"regular_bonus_bytes"`
	}
	if err := decodeJSONBody(r, &payload); err != nil {
		return err
	}
	bonusBytes := gbToBytes(payload.RegularBonusGB)
	if payload.RegularBonusBytes != nil {
		bonusBytes = *payload.RegularBonusBytes
	}
	if bonusBytes < 0 {
		return fmt.Errorf("invalid_regular_bonus")
	}
	user, panelUser, err := adminPanelUser(ctx, settings, pool, panel, userID)
	if err != nil {
		return err
	}
	override := loadUserTrafficOverride(ctx, pool, userID)
	baseLimit := regularBaseTrafficLimit(ctx, pool, panel, user, panelUser, override)
	nextLimit := baseLimit + bonusBytes
	if payload.Unlimited {
		nextLimit = 0
		bonusBytes = 0
	}
	update := map[string]any{
		"uuid":                 stringValue(panelUser, "uuid"),
		"trafficLimitBytes":    nextLimit,
		"trafficLimitStrategy": panel.EffectiveConfig(ctx).UserTrafficStrategy,
	}
	if _, err := panel.UpdateUser(ctx, update); err != nil {
		return err
	}
	override.RegularUnlimited = payload.Unlimited
	override.RegularBonusBytes = bonusBytes
	return saveUserTrafficOverride(ctx, pool, userID, override)
}

func adminGrantPanelTraffic(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client, userID int64, r *http.Request) error {
	var payload struct {
		Kind  string  `json:"kind"`
		GB    float64 `json:"gb"`
		Bytes *int64  `json:"bytes"`
	}
	if err := decodeJSONBody(r, &payload); err != nil {
		return err
	}
	grantBytes := gbToBytes(payload.GB)
	if payload.Bytes != nil {
		grantBytes = *payload.Bytes
	}
	if grantBytes <= 0 {
		return fmt.Errorf("invalid_gb")
	}
	override := loadUserTrafficOverride(ctx, pool, userID)
	if strings.EqualFold(payload.Kind, "premium") {
		override.PremiumBonusBytes += grantBytes
		if override.PremiumUnlimited {
			override.PremiumBonusBytes = 0
		}
		return saveUserTrafficOverride(ctx, pool, userID, override)
	}
	if panel == nil || !panel.Configured(ctx) {
		return remnawave.ErrNotConfigured
	}
	_, panelUser, err := adminPanelUser(ctx, settings, pool, panel, userID)
	if err != nil {
		return err
	}
	current := int64Value(panelUser, "trafficLimitBytes")
	nextLimit := current + grantBytes
	if override.RegularUnlimited {
		nextLimit = 0
	} else {
		override.RegularBonusBytes += grantBytes
	}
	update := map[string]any{
		"uuid":              stringValue(panelUser, "uuid"),
		"trafficLimitBytes": nextLimit,
	}
	if _, err = panel.UpdateUser(ctx, update); err != nil {
		return err
	}
	return saveUserTrafficOverride(ctx, pool, userID, override)
}

func regularBaseTrafficLimit(ctx context.Context, pool *pgxpool.Pool, panel *remnawave.Client, user webappUser, panelUser map[string]any, override userTrafficOverride) int64 {
	plan := latestPaidPlanForUser(ctx, pool, user.UserID)
	if plan.TariffKey != "" || plan.Title != "" || plan.MonthlyGB > 0 || plan.TrafficGB > 0 {
		return trafficLimitBytesForPlan(plan, panel.EffectiveConfig(ctx), 0)
	}
	current := int64Value(panelUser, "trafficLimitBytes")
	base := current - override.RegularBonusBytes
	if base < 0 {
		return 0
	}
	return base
}

func adminPanelUser(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client, userID int64) (webappUser, map[string]any, error) {
	user, err := loadWebappUser(ctx, pool, userID, settings)
	if err != nil {
		return webappUser{}, nil, err
	}
	panelUser, found, err := panelUserForWebUser(ctx, pool, panel, user)
	if err != nil {
		return webappUser{}, nil, err
	}
	if !found {
		return webappUser{}, nil, remnawave.ErrNotFound
	}
	return user, panelUser, nil
}

func firstPlanForTariff(_ context.Context, settings config.Settings, language string, tariffKey string) (tariffs.Plan, bool) {
	catalog, err := tariffs.Load("data/tariffs.json")
	if err != nil {
		return tariffs.Plan{}, false
	}
	plans := catalog.Plans(language, settings.DefaultCurrency)
	for _, plan := range plans {
		if plan.TariffKey == tariffKey && plan.SaleMode == "subscription" {
			return plan, true
		}
	}
	for _, plan := range plans {
		if plan.TariffKey == tariffKey {
			return plan, true
		}
	}
	return tariffs.Plan{}, false
}

func mergePlanUpdate(update map[string]any, plan tariffs.Plan, cfg remnawave.EffectiveConfig) {
	if plan.MonthlyGB > 0 {
		update["trafficLimitBytes"] = gbToBytes(plan.MonthlyGB)
		update["trafficLimitStrategy"] = cfg.UserTrafficStrategy
	}
	if len(plan.SquadUUIDs) > 0 {
		update["activeInternalSquads"] = cleanStringSlice(plan.SquadUUIDs)
	} else if len(cfg.UserSquadUUIDs) > 0 {
		update["activeInternalSquads"] = cleanStringSlice(cfg.UserSquadUUIDs)
	}
	if external := firstNonEmpty(plan.ExternalSquadUUID, cfg.UserExternalSquadUUID); external != "" {
		update["externalSquadUuid"] = external
	}
}

// RunSubscriptionNotifications checks active subscriptions and sends expiry
// and traffic-exhaustion notifications. Users can toggle and configure each
// notification type through their Web App settings. Notifications are only
// sent when each per-user toggle is enabled.
func RunSubscriptionNotifications(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) (int, error) {
	if pool == nil {
		return 0, fmt.Errorf("database_not_configured")
	}
	if panel == nil || !panel.Configured(ctx) {
		return 0, nil
	}
	if !appsettings.NewStore(pool).Bool(ctx, "SUBSCRIPTION_NOTIFICATIONS_ENABLED", true) {
		return 0, nil
	}

	users, err := loadSubscribableUsers(ctx, pool, settings, 500)
	if err != nil {
		return 0, err
	}
	notified := 0
	store := appsettings.NewStore(pool)
	notificationMailer := mail.NewMailer(mailerConfigFromSettings(ctx, store))
	mailEnabled := store.Bool(ctx, "SMTP_ENABLED", false) && notificationMailer.IsConfigured()
	notifyHoursBefore := store.Int(ctx, "SUBSCRIPTION_NOTIFY_HOURS_BEFORE", settings.SubscriptionNotifyHoursBefore)
	notifyDaysBefore := store.Int(ctx, "SUBSCRIPTION_NOTIFY_DAYS_BEFORE", settings.SubscriptionNotifyDaysBefore)

	for _, user := range users {
		panelUser, found, err := panelUserForWebUser(ctx, pool, panel, user)
		if err != nil || !found {
			continue
		}

		chatID := telegramChatIDForWebUser(user)
		userEmail := strings.TrimSpace(user.Email)

		// Load per-user notification preferences
		prefs := loadUserNotificationPrefs(ctx, pool, user.UserID)

		// ---- Expiry notifications ----
		if prefs.ExpiryEnabled {
			expireAt := parsePanelTime(panelUser["expireAt"])
			if !expireAt.IsZero() {
				status := strings.ToUpper(strings.TrimSpace(stringValue(panelUser, "status")))
				now := time.Now().UTC()
				secondsLeft := expireAt.Sub(now).Seconds()

				// Use per-user days-before preference
				effectiveDaysBefore := notifyDaysBefore
				if prefs.ExpiryDaysBefore > 0 {
					effectiveDaysBefore = prefs.ExpiryDaysBefore
				}
				stages := subscriptionNotificationStagesWithOverrides(settings, secondsLeft, status, expireAt, now, notifyHoursBefore, effectiveDaysBefore)
				for _, stage := range stages {
					notificationKey := fmt.Sprintf("%d:%s", user.UserID, stage.key)
					if subscriptionNotificationAlreadySent(ctx, pool, notificationKey, 24*time.Hour) {
						continue
					}
					text := subscriptionNotificationText(stage, expireAt, settings.DefaultLanguage)
					sent := false
					if chatID != 0 {
						if err := sendTelegramText(ctx, settings, chatID, text); err == nil {
							sent = true
						}
					}
					if !sent && userEmail != "" && mailEnabled {
						sent = trySendSubscriptionEmail(ctx, pool, notificationMailer, user, store.String(ctx, "BRAND_NAME", "Remna"), text)
					}
					if sent {
						recordSubscriptionNotification(ctx, pool, notificationKey)
						notified++
					}
					break // only send the most urgent notification per user
				}
			}
		}

		// ---- Traffic exhaustion notifications ----
		if prefs.TrafficEnabled {
			traffic := mapValue(panelUser, "userTraffic")
			used := int64Value(traffic, "usedTrafficBytes")
			limit := int64Value(panelUser, "trafficLimitBytes")
			if limit > 0 {
				pct := int(float64(used) / float64(limit) * 100)
				thresholdPct := prefs.TrafficThresholdPct
				if thresholdPct <= 0 {
					thresholdPct = defaultTrafficThresholdPct
				}
				if pct >= thresholdPct {
					trafficKey := fmt.Sprintf("%d:traffic_%d", user.UserID, thresholdPct)
					if !subscriptionNotificationAlreadySent(ctx, pool, trafficKey, 48*time.Hour) {
						text := fmtTrafficExhaustionText(used, limit, pct, thresholdPct)
						sent := false
						if chatID != 0 {
							if err := sendTelegramText(ctx, settings, chatID, text); err == nil {
								sent = true
							}
						}
						if !sent && userEmail != "" && mailEnabled {
							sent = trySendSubscriptionEmail(ctx, pool, notificationMailer, user, store.String(ctx, "BRAND_NAME", "Remna"), text)
						}
						if sent {
							recordSubscriptionNotification(ctx, pool, trafficKey)
							notified++
						}
					}
				}
			}
		}

		// ---- Trial traffic depletion (always notify, regardless of prefs) ----
		if isTrialSubscription(panelUser) {
			traffic := mapValue(panelUser, "userTraffic")
			used := int64Value(traffic, "usedTrafficBytes")
			limit := int64Value(panelUser, "trafficLimitBytes")
			if limit > 0 && used >= limit {
				trialKey := fmt.Sprintf("%d:trial_traffic_depleted", user.UserID)
				if !subscriptionNotificationAlreadySent(ctx, pool, trialKey, 48*time.Hour) {
					text := "⚠️ Your trial traffic has been fully used. Upgrade to a paid plan to continue using the service."
					sent := false
					if chatID != 0 {
						if err := sendTelegramText(ctx, settings, chatID, text); err == nil {
							sent = true
						}
					}
					if !sent && userEmail != "" && mailEnabled {
						sent = trySendSubscriptionEmail(ctx, pool, notificationMailer, user, store.String(ctx, "BRAND_NAME", "Remna"), text)
					}
					if sent {
						recordSubscriptionNotification(ctx, pool, trialKey)
						notified++
					}
				}
			}
		}
	}
	return notified, nil
}

// subscriptionNotificationStagesWithOverrides allows overriding notification
// thresholds from app_settings values. When hoursBefore/daysBefore are 0,
// the env config defaults are used.
func subscriptionNotificationStagesWithOverrides(settings config.Settings, secondsLeft float64, status string, expireAt, now time.Time, hoursBeforeOverride, daysBeforeOverride int) []notifyStage {
	hoursBefore := settings.SubscriptionNotifyHoursBefore
	daysBefore := settings.SubscriptionNotifyDaysBefore
	if hoursBeforeOverride > 0 {
		hoursBefore = hoursBeforeOverride
	}
	if daysBeforeOverride > 0 {
		daysBefore = daysBeforeOverride
	}

	var stages []notifyStage

	if secondsLeft > 0 {
		// Before expiry notifications
		if hoursBefore > 0 && hoursBefore <= 23 && secondsLeft <= float64(hoursBefore)*3600 {
			stages = append(stages, notifyStage{
				key:       fmt.Sprintf("before_%dh", hoursBefore),
				hoursLeft: hoursBefore,
			})
		}

		dayStages := []struct {
			days int
			key  string
		}{
			{1, "before_1d"},
			{2, "before_2d"},
			{3, "before_3d"},
		}
		for _, ds := range dayStages {
			if ds.days <= daysBefore && secondsLeft <= float64(ds.days)*86400 {
				stages = append(stages, notifyStage{
					key:      ds.key,
					daysLeft: ds.days,
				})
				break
			}
		}
	} else {
		// Expired notifications
		expiredFor := now.Sub(expireAt)
		if expiredFor <= 24*time.Hour && (status == "EXPIRED" || status == "DISABLED" || status == "") {
			stages = append(stages, notifyStage{
				key:       "expired",
				isExpired: true,
			})
		} else if expiredFor > 24*time.Hour && expiredFor <= 48*time.Hour {
			stages = append(stages, notifyStage{
				key:       "expired_24h_after",
				isPostExp: true,
			})
		}
	}
	return stages
}

func subscriptionNotificationText(stage notifyStage, expireAt time.Time, _ string) string {
	dateText := expireAt.Format("2006-01-02")
	switch {
	case stage.isExpired:
		return fmt.Sprintf("⏰ Your subscription has expired on %s. Renew now to restore access.", dateText)
	case stage.isPostExp:
		return fmt.Sprintf("🔔 Your subscription expired on %s. Don't lose your configuration — renew now!", dateText)
	case stage.hoursLeft > 0:
		return fmt.Sprintf("⏳ Your subscription expires in %d hours (on %s). Renew now to avoid interruption.", stage.hoursLeft, dateText)
	case stage.daysLeft > 0:
		return fmt.Sprintf("📅 Your subscription expires in %d day(s) (on %s). Renew now to stay connected.", stage.daysLeft, dateText)
	default:
		return fmt.Sprintf("📅 Your subscription expires soon (on %s).", dateText)
	}
}

// fmtTrafficExhaustionText formats a traffic exhaustion notification message.
func fmtTrafficExhaustionText(usedBytes, limitBytes int64, usedPct, _ int) string {
	usedGB := float64(usedBytes) / bytesPerGB
	limitGB := float64(limitBytes) / bytesPerGB
	return fmt.Sprintf("📊 You have used %.1f GB out of %.1f GB (%d%%). Your traffic is running low!", usedGB, limitGB, usedPct)
}

func sendSubscriptionEmail(mailer *mail.Mailer, recipient, brand, text string) error {
	return mailer.Send(mail.Message{
		To:        []string{recipient},
		Subject:   brand + " subscription notification",
		BodyPlain: text,
	})
}

func trySendSubscriptionEmail(ctx context.Context, pool *pgxpool.Pool, mailer *mail.Mailer, user webappUser, brand, text string) bool {
	err := sendSubscriptionEmail(mailer, strings.TrimSpace(user.Email), brand, text)
	if err == nil {
		return true
	}
	recordMessageLog(ctx, pool, messageLogEntry{UserID: user.UserID, TargetUserID: user.UserID, EventType: "subscription_email_failed", Content: text, Payload: map[string]any{"error": err.Error()}})
	return false
}

func isTrialSubscription(panelUser map[string]any) bool {
	status := strings.ToUpper(strings.TrimSpace(stringValue(panelUser, "status")))
	return status == "TRIAL" || strings.Contains(strings.ToLower(stringValue(panelUser, "description")), "trial")
}

func loadSubscribableUsers(ctx context.Context, pool *pgxpool.Pool, settings config.Settings, limit int) ([]webappUser, error) {
	if pool == nil {
		return nil, nil
	}
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	rows, err := pool.Query(ctx, `
SELECT user_id, COALESCE(telegram_id,0), COALESCE(username,''), COALESCE(email,''), COALESCE(first_name,''), COALESCE(last_name,''),
	COALESCE(language_code,''), COALESCE(telegram_photo_url,''), COALESCE(panel_user_uuid,'')
FROM users
WHERE COALESCE(panel_user_uuid,'') <> '' AND is_banned=FALSE
ORDER BY registration_date DESC
LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []webappUser{}
	for rows.Next() {
		var user webappUser
		if err := rows.Scan(&user.UserID, &user.TelegramID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.LanguageCode, &user.PhotoURL, &user.PanelUserUUID); err != nil {
			return nil, err
		}
		user.LanguageCode = normalizeWebLanguage(user.LanguageCode, effectiveDefaultLanguage(ctx, pool, settings))
		user.IsAdmin = isAdminID(settings.AdminIDs, user.UserID) || isAdminID(settings.AdminIDs, user.TelegramID)
		users = append(users, user)
	}
	return users, rows.Err()
}

func subscriptionNotificationAlreadySent(ctx context.Context, pool *pgxpool.Pool, notificationKey string, within time.Duration) bool {
	sent := readSubscriptionNotifications(ctx, pool)
	if timestamp, ok := sent[notificationKey]; ok {
		if parsed, err := time.Parse(time.RFC3339, timestamp); err == nil {
			if time.Since(parsed) < within {
				return true
			}
		}
	}
	return false
}

func recordSubscriptionNotification(ctx context.Context, pool *pgxpool.Pool, notificationKey string) {
	sent := readSubscriptionNotifications(ctx, pool)
	if sent == nil {
		sent = map[string]string{}
	}
	// Skip if the same key was already recorded recently (no-op).
	if ts, exists := sent[notificationKey]; exists {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			if time.Since(parsed) < time.Hour {
				return
			}
		}
	}
	// Clean old entries (older than 7 days)
	cutoff := time.Now().UTC().Add(-7 * 24 * time.Hour)
	dirty := false
	for key, timestamp := range sent {
		if parsed, err := time.Parse(time.RFC3339, timestamp); err == nil && parsed.Before(cutoff) {
			delete(sent, key)
			dirty = true
		}
	}
	// Only write if entries actually changed.
	if !dirty {
		if _, ok := sent[notificationKey]; ok {
			// Already have the key and no cleanup needed — skip write entirely.
			return
		}
	}
	sent[notificationKey] = time.Now().UTC().Format(time.RFC3339)
	_ = appsettings.NewStore(pool).Upsert(ctx, subscriptionNotificationsSentKey, sent)
}

func readSubscriptionNotifications(ctx context.Context, pool *pgxpool.Pool) map[string]string {
	raw, ok, err := appsettings.NewStore(pool).Get(ctx, subscriptionNotificationsSentKey)
	if err != nil || !ok {
		return map[string]string{}
	}
	var sent map[string]string
	if json.Unmarshal(raw, &sent) == nil && sent != nil {
		return sent
	}
	return map[string]string{}
}
