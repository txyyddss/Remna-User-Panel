package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"remna-user-panel/internal/config"
	"remna-user-panel/internal/payments"
	"remna-user-panel/internal/remnawave"
	"remna-user-panel/internal/tariffs"
)

const bytesPerGB = 1024 * 1024 * 1024

type paidPlan struct {
	TariffKey         string   `json:"tariff_key"`
	TariffName        string   `json:"tariff_name"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	BillingModel      string   `json:"billing_model"`
	SaleMode          string   `json:"sale_mode"`
	Months            int      `json:"months"`
	TrafficGB         float64  `json:"traffic_gb"`
	MonthlyGB         float64  `json:"monthly_gb"`
	SquadUUIDs        []string `json:"squad_uuids"`
	ExternalSquadUUID string   `json:"external_squad_uuid"`
	HWIDDeviceLimit   *int     `json:"hwid_device_limit"`
	Provider          string   `json:"provider"`
}

type accessGrantOptions struct {
	Source          string
	TrafficLimitGB  float64
	TrafficStrategy string
	SquadUUIDs      []string
	SetTrafficLimit bool
}

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
	hwidLimit, hasHWIDLimit := optionalIntValue(panelUser, "hwidDeviceLimit")
	if !hasHWIDLimit && plan.HWIDDeviceLimit != nil {
		hwidLimit = *plan.HWIDDeviceLimit
		hasHWIDLimit = true
	}
	maxDevices := any(nil)
	if hasHWIDLimit {
		maxDevices = hwidLimit
	}
	days := 0
	if !expireAt.IsZero() && expireAt.After(now) {
		days = int(math.Ceil(expireAt.Sub(now).Hours() / 24))
	}
	tariffName := plan.TariffName
	if tariffName == "" {
		tariffName = plan.Title
	}
	payload := map[string]any{
		"active":                    active,
		"status":                    status,
		"end_date":                  timeString(expireAt),
		"end_date_text":             dateText(expireAt),
		"days_left":                 days,
		"remaining_text":            remainingText(days),
		"config_link":               stringValue(panelUser, "subscriptionUrl"),
		"connect_url":               stringValue(panelUser, "subscriptionUrl"),
		"panel_short_uuid":          stringValue(panelUser, "shortUuid"),
		"traffic_used":              bytesText(usedBytes),
		"traffic_limit":             bytesText(limitBytes),
		"traffic_used_bytes":        usedBytes,
		"traffic_limit_bytes":       limitBytes,
		"traffic_limit_strategy":    stringValue(panelUser, "trafficLimitStrategy"),
		"can_topup_traffic":         active && limitBytes > 0,
		"can_topup_regular_traffic": active && limitBytes > 0,
		"can_topup_premium_traffic": false,
		"can_topup_devices":         active && hasHWIDLimit && hwidLimit != 0,
		"max_devices":               maxDevices,
		"extra_hwid_devices":        0,
		"auto_renew_enabled":        false,
		"auto_renew_available":      false,
		"tariff_key":                plan.TariffKey,
		"tariff_name":               tariffName,
		"billing_model":             plan.BillingModel,
		"tariff_description":        plan.Description,
		"provider":                  plan.Provider,
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
	if plan.HWIDDeviceLimit != nil {
		payload["hwidDeviceLimit"] = *plan.HWIDDeviceLimit
	} else if cfg.UserHWIDDeviceLimit != nil {
		payload["hwidDeviceLimit"] = *cfg.UserHWIDDeviceLimit
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
	if cfg.UserHWIDDeviceLimit != nil {
		payload["hwidDeviceLimit"] = *cfg.UserHWIDDeviceLimit
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

var panelUsernameInvalid = regexp.MustCompile(`[^A-Za-z0-9_-]+`)

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
		if err := rows.Scan(
			&order.PaymentID, &order.OrderID, &order.UserID, &order.Provider, &order.Method, &order.PaymentType,
			&order.Amount, &order.Currency, &order.BaseAmount, &order.BaseCurrency, &order.DisplayCNYAmount,
			&order.FXRate, &order.FXSource, &order.FXUpdatedAt, &order.PlanHash, &order.PlanSnapshot,
			&order.Status, &order.Description, &order.TariffKey, &order.SaleMode, &order.Months, &order.TrafficGB,
			&order.DeviceCount, &order.ProviderPaymentID, &order.PaymentURL, &order.QRContent, &order.DisplayAmount,
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
	enrichAdminUserFromPanel(user, panelUser)
	return user
}

func enrichAdminUserFromPanel(user map[string]any, panelUser map[string]any) {
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
	user["premium_traffic"] = map[string]any{"state": "none"}
}

func mapDevicePayload(panelPayload map[string]any, panelUser map[string]any) map[string]any {
	rawDevices := arrayValue(panelPayload, "devices")
	devices := make([]map[string]any, 0, len(rawDevices))
	for index, raw := range rawDevices {
		device, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		hwid := stringValue(device, "hwid")
		devices = append(devices, map[string]any{
			"index":           index + 1,
			"token":           hwid,
			"hwid":            hwid,
			"hwid_short":      shortHWID(hwid),
			"display_name":    firstNonEmpty(stringValue(device, "deviceModel"), "Device "+strconv.Itoa(index+1)),
			"platform_label":  platformLabel(device),
			"user_agent":      stringValue(device, "userAgent"),
			"created_at":      timeString(parsePanelTime(device["createdAt"])),
			"created_at_text": dateTimeText(parsePanelTime(device["createdAt"])),
			"can_disconnect":  hwid != "",
		})
	}
	maxDevices, hasMax := optionalIntValue(panelUser, "hwidDeviceLimit")
	maxLabel := ""
	if hasMax {
		if maxDevices == 0 {
			maxLabel = "∞"
		} else {
			maxLabel = strconv.Itoa(maxDevices)
		}
	}
	return map[string]any{
		"ok":                true,
		"enabled":           true,
		"current_devices":   len(devices),
		"max_devices":       maxDevices,
		"max_devices_label": maxLabel,
		"devices":           devices,
	}
}

func platformLabel(device map[string]any) string {
	platform := stringValue(device, "platform")
	osVersion := stringValue(device, "osVersion")
	if platform == "" {
		return osVersion
	}
	if osVersion == "" {
		return platform
	}
	return platform + " " + osVersion
}

func shortHWID(hwid string) string {
	hwid = strings.TrimSpace(hwid)
	if len(hwid) <= 14 {
		return hwid
	}
	return hwid[:8] + "..." + hwid[len(hwid)-6:]
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

func dateTimeText(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format("2006-01-02 15:04")
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

func optionalIntValue(m map[string]any, key string) (int, bool) {
	if m == nil {
		return 0, false
	}
	value, ok := m[key]
	if !ok || value == nil {
		return 0, false
	}
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	case string:
		if strings.TrimSpace(typed) == "" {
			return 0, false
		}
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		return parsed, err == nil
	default:
		return 0, false
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

func arrayValue(m map[string]any, key string) []any {
	if m == nil {
		return []any{}
	}
	switch value := m[key].(type) {
	case []any:
		return value
	case []map[string]any:
		result := make([]any, 0, len(value))
		for _, item := range value {
			result = append(result, item)
		}
		return result
	default:
		return []any{}
	}
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

func adminSetPanelHWIDLimit(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client, userID int64, r *http.Request) error {
	if panel == nil || !panel.Configured(ctx) {
		return remnawave.ErrNotConfigured
	}
	var payload struct {
		Unlimited       bool `json:"unlimited"`
		UseDefault      bool `json:"use_default"`
		HWIDDeviceLimit *int `json:"hwid_device_limit"`
		Limit           *int `json:"limit"`
	}
	if err := decodeJSONBody(r, &payload); err != nil {
		return err
	}
	_, panelUser, err := adminPanelUser(ctx, settings, pool, panel, userID)
	if err != nil {
		return err
	}
	var limit any
	switch {
	case payload.Unlimited:
		limit = 0
	case payload.UseDefault:
		cfg := panel.EffectiveConfig(ctx)
		if cfg.UserHWIDDeviceLimit == nil {
			limit = nil
		} else {
			limit = *cfg.UserHWIDDeviceLimit
		}
	case payload.HWIDDeviceLimit != nil:
		limit = *payload.HWIDDeviceLimit
	case payload.Limit != nil:
		limit = *payload.Limit
	default:
		limit = nil
	}
	update := map[string]any{"uuid": stringValue(panelUser, "uuid"), "hwidDeviceLimit": limit}
	_, err = panel.UpdateUser(ctx, update)
	return err
}

func adminGrantPanelTraffic(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client, userID int64, r *http.Request) error {
	if panel == nil || !panel.Configured(ctx) {
		return remnawave.ErrNotConfigured
	}
	var payload struct {
		Kind string  `json:"kind"`
		GB   float64 `json:"gb"`
	}
	if err := decodeJSONBody(r, &payload); err != nil {
		return err
	}
	if payload.GB <= 0 {
		return fmt.Errorf("invalid_gb")
	}
	if strings.EqualFold(payload.Kind, "premium") {
		return nil
	}
	_, panelUser, err := adminPanelUser(ctx, settings, pool, panel, userID)
	if err != nil {
		return err
	}
	current := int64Value(panelUser, "trafficLimitBytes")
	update := map[string]any{
		"uuid":              stringValue(panelUser, "uuid"),
		"trafficLimitBytes": current + gbToBytes(payload.GB),
	}
	_, err = panel.UpdateUser(ctx, update)
	return err
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

func firstPlanForTariff(ctx context.Context, settings config.Settings, language string, tariffKey string) (tariffs.Plan, bool) {
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
	if plan.HWIDDeviceLimit != nil {
		update["hwidDeviceLimit"] = *plan.HWIDDeviceLimit
	} else if cfg.UserHWIDDeviceLimit != nil {
		update["hwidDeviceLimit"] = *cfg.UserHWIDDeviceLimit
	}
}
