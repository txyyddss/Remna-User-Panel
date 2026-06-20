// Package config loads Remnawave Minishop runtime configuration from environment variables.
package config

import (
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultLanguage = "zh"
	defaultDBName   = "vpn_shop_db"
)

// Settings contains process-wide runtime configuration.
type Settings struct {
	BotToken                      string
	AdminIDs                      []int64
	AdminEmail                    string
	AdminPassword                 string
	DefaultLanguage               string
	DefaultCurrency               string
	WebhookBaseURL                string
	WebhookSecretToken            string
	WebAppSessionSecret           string
	WebServerHost                 string
	WebServerPort                 int
	WebAppEnabled                 bool
	WebAppServerHost              string
	WebAppServerPort              int
	SubscriptionMiniApp           string
	PostgresUser                  string
	PostgresPassword              string
	PostgresHost                  string
	PostgresPort                  int
	PostgresDB                    string
	DatabaseURL                   string
	RedisURL                      string
	RedisKeyPrefix                string
	TrustedProxies                []string
	PanelAPIURL                   string
	PanelAPIKey                   string
	PanelAPITotalTimeout          time.Duration
	LogLevel                      string
	WorkerPanelSyncEvery          time.Duration
	WorkerPaymentProvisionEvery   time.Duration
	UserTrafficLimitGB            float64
	UserTrafficStrategy           string
	UserSquadUUIDs                []string
	UserExternalSquadUUID         string
	EZPay                         EZPaySettings
	BEPUSDT                       BEPUSDTSettings
	PaymentMethodsOrder           []string
	SubscriptionNotifyHoursBefore int
	SubscriptionNotifyDaysBefore  int
	StarsUSDRate                  float64
	StarsEnabled                  bool
}

// EZPaySettings contains EZPay merchant configuration.
type EZPaySettings struct {
	Enabled   bool
	BaseURL   string
	PID       int
	Key       string
	ReturnURL string
}

// BEPUSDTSettings contains BEPUSDT merchant configuration.
type BEPUSDTSettings struct {
	Enabled   bool
	BaseURL   string
	Token     string
	ReturnURL string
}

// Load reads settings from the current process environment.
func Load() (Settings, error) {
	publicURL := strings.TrimRight(env("PUBLIC_URL", ""), "/")
	webhookBaseURL := strings.TrimRight(env("WEBHOOK_BASE_URL", ""), "/")
	subscriptionMiniApp := env("SUBSCRIPTION_MINI_APP_URL", "")
	// PUBLIC_URL is a convenience variable for single-domain deployments.
	// When set, it provides defaults for both WEBHOOK_BASE_URL and
	// SUBSCRIPTION_MINI_APP_URL. Explicit values always take precedence.
	if webhookBaseURL == "" && publicURL != "" {
		webhookBaseURL = publicURL
	}
	if subscriptionMiniApp == "" && publicURL != "" {
		subscriptionMiniApp = publicURL + "/"
	}

	settings := Settings{
		BotToken:                    env("BOT_TOKEN", ""),
		AdminEmail:                  strings.ToLower(env("ADMIN_EMAIL", "")),
		AdminPassword:               env("ADMIN_PASSWORD", ""),
		DefaultLanguage:             NormalizeLanguage(env("DEFAULT_LANGUAGE", defaultLanguage)),
		DefaultCurrency:             env("DEFAULT_CURRENCY_SYMBOL", "USD"),
		WebhookBaseURL:              webhookBaseURL,
		WebhookSecretToken:          env("WEBHOOK_SECRET_TOKEN", ""),
		WebAppSessionSecret:         env("WEBAPP_SESSION_SECRET", ""),
		WebServerHost:               env("WEB_SERVER_HOST", "0.0.0.0"),
		WebServerPort:               envInt("WEB_SERVER_PORT", 8080),
		WebAppEnabled:               envBool("WEBAPP_ENABLED", true),
		WebAppServerHost:            env("WEBAPP_SERVER_HOST", "0.0.0.0"),
		WebAppServerPort:            envInt("WEBAPP_SERVER_PORT", 8081),
		SubscriptionMiniApp:         subscriptionMiniApp,
		PostgresUser:                env("POSTGRES_USER", ""),
		PostgresPassword:            env("POSTGRES_PASSWORD", ""),
		PostgresHost:                env("POSTGRES_HOST", "localhost"),
		PostgresPort:                envInt("POSTGRES_PORT", 5432),
		PostgresDB:                  env("POSTGRES_DB", defaultDBName),
		RedisURL:                    env("REDIS_URL", ""),
		RedisKeyPrefix:              env("REDIS_KEY_PREFIX", "remna-user-panel"),
		PanelAPIURL:                 strings.TrimRight(env("PANEL_API_URL", ""), "/"),
		PanelAPIKey:                 env("PANEL_API_KEY", ""),
		PanelAPITotalTimeout:        time.Duration(envInt("PANEL_API_TOTAL_TIMEOUT_SECONDS", 25)) * time.Second,
		LogLevel:                    env("LOG_LEVEL", "INFO"),
		WorkerPanelSyncEvery:        time.Duration(envInt("WORKER_PANEL_SYNC_INTERVAL_SECONDS", 900)) * time.Second,
		WorkerPaymentProvisionEvery: time.Duration(envInt("WORKER_PAYMENT_PROVISION_INTERVAL_SECONDS", 30)) * time.Second,
		UserTrafficLimitGB:          envFloat("USER_TRAFFIC_LIMIT_GB", 0),
		UserTrafficStrategy:         normalizeTrafficStrategy(env("USER_TRAFFIC_STRATEGY", "NO_RESET")),
		UserSquadUUIDs:              splitCSV(env("USER_SQUAD_UUIDS", "")),
		UserExternalSquadUUID:       strings.TrimSpace(env("USER_EXTERNAL_SQUAD_UUID", "")),
	
		EZPay: EZPaySettings{
			Enabled:   envBool("EZPAY_ENABLED", false),
			BaseURL:   strings.TrimRight(env("EZPAY_BASE_URL", ""), "/"),
			PID:       envInt("EZPAY_PID", 0),
			Key:       env("EZPAY_KEY", ""),
		},
		BEPUSDT: BEPUSDTSettings{
			Enabled:   envBool("BEPUSDT_ENABLED", false),
			BaseURL:   strings.TrimRight(env("BEPUSDT_BASE_URL", ""), "/"),
			Token:     env("BEPUSDT_TOKEN", ""),
		},
		PaymentMethodsOrder:           splitCSV(env("PAYMENT_METHODS_ORDER", "")),
		SubscriptionNotifyHoursBefore: envInt("SUBSCRIPTION_NOTIFY_HOURS_BEFORE", 0),
		SubscriptionNotifyDaysBefore:  envInt("SUBSCRIPTION_NOTIFY_DAYS_BEFORE", 3),
		StarsUSDRate:                  envFloat("STARS_USD_RATE", 0),
		StarsEnabled:                 envBool("STARS_ENABLED", false),
	}
	settings.AdminIDs = parseInt64List(env("ADMIN_IDS", ""))
	settings.TrustedProxies = splitCSV(env("TRUSTED_PROXIES", "127.0.0.1,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,fc00::/7"))

	// Database URL: SHOP_DATABASE_URL (User Panel 专属) 优先于 DATABASE_URL（可能与 Remnawave 面板冲突）。
	// 如果都未设置，则从 POSTGRES_* 组件变量构建。
	settings.DatabaseURL = env("SHOP_DATABASE_URL", "")
	if settings.DatabaseURL == "" {
		settings.DatabaseURL = env("DATABASE_URL", "")
	}
	if settings.DatabaseURL == "" {
		if settings.PostgresUser == "" || settings.PostgresPassword == "" {
			return Settings{}, fmt.Errorf("POSTGRES_USER and POSTGRES_PASSWORD are required when DATABASE_URL is empty")
		}
		settings.DatabaseURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			urlQueryEscape(settings.PostgresUser),
			urlQueryEscape(settings.PostgresPassword),
			settings.PostgresHost,
			settings.PostgresPort,
			settings.PostgresDB,
		)
	}

	// Redis URL: 支持 REDIS_URL 完整 URL，或从 REDIS_HOST / REDIS_PORT 构建。
	if settings.RedisURL == "" {
		redisHost := env("REDIS_HOST", "")
		if redisHost != "" {
			redisPort := envInt("REDIS_PORT", 6379)
			settings.RedisURL = fmt.Sprintf("redis://%s:%d/0", redisHost, redisPort)
		}
	}

	return settings, nil
}

// WebhookPath returns the Telegram webhook path used by the backend HTTP server.
func (s Settings) WebhookPath() string {
	if s.BotToken == "" {
		return "/webhook/telegram"
	}
	return "/webhook/" + s.BotToken
}

// WebhookURL returns the public Telegram webhook URL.
func (s Settings) WebhookURL() string {
	if s.WebhookBaseURL == "" {
		return ""
	}
	return strings.TrimRight(s.WebhookBaseURL, "/") + s.WebhookPath()
}

// WebListenAddr returns the backend webhook/health listen address.
func (s Settings) WebListenAddr() string {
	return net.JoinHostPort(s.WebServerHost, strconv.Itoa(s.WebServerPort))
}

// WebAppListenAddr returns the Mini App/admin listen address.
func (s Settings) WebAppListenAddr() string {
	return net.JoinHostPort(s.WebAppServerHost, strconv.Itoa(s.WebAppServerPort))
}

func env(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func envInt(key string, fallback int) int {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return value
}

func envFloat(key string, fallback float64) float64 {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return fallback
	}
	return value
}



func parseInt64List(raw string) []int64 {
	parts := splitCSV(raw)
	result := make([]int64, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.ParseInt(part, 10, 64)
		if err == nil {
			result = append(result, value)
		} else {
			slog.Warn("invalid admin id ignored", "value", part, "error", err)
		}
	}
	return result
}

func splitCSV(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		value := strings.TrimSpace(field)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func NormalizeLanguage(raw string) string {
	value := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(raw), "_", "-"))
	if value == "" {
		return defaultLanguage
	}
	return value
}

func normalizeTrafficStrategy(raw string) string {
	value := strings.ToUpper(strings.TrimSpace(raw))
	switch value {
	case "DAY", "WEEK", "MONTH", "MONTH_ROLLING":
		return value
	default:
		return "NO_RESET"
	}
}

func urlQueryEscape(value string) string {
	return url.QueryEscape(value)
}
