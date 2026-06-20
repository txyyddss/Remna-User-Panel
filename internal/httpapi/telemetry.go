package httpapi

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"remna-user-panel/internal/config"
	appsettings "remna-user-panel/internal/settings"
)

const visitorCookieName = "rw_anon_visitor"

type telemetryFingerprint struct {
	Full     string `json:"full"`
	Canvas   string `json:"canvas"`
	WebGL    string `json:"webgl"`
	Fonts    string `json:"fonts"`
	Audio    string `json:"audio"`
	Browser  string `json:"browser"`
	Platform string `json:"platform"`
	Timezone string `json:"timezone"`
	Screen   string `json:"screen"`
	Hardware string `json:"hardware"`
	Language string `json:"language"`
}

type telemetryHeartbeatPayload struct {
	Fingerprint telemetryFingerprint `json:"fingerprint"`
	InviteCode  string               `json:"invite_code"`
}

type storedFingerprint struct {
	Visitor, Full, Canvas, WebGL, Fonts, Audio, Network     string
	Browser, Platform, Timezone, Screen, Hardware, Language string
}

func telemetryEnabled(ctx context.Context, pool *pgxpool.Pool) bool {
	return appsettings.NewStore(pool).Bool(ctx, "TELEMETRY_ENABLED", true)
}

func telemetryRetention(ctx context.Context, pool *pgxpool.Pool) time.Duration {
	hours := appsettings.NewStore(pool).Int(ctx, "TELEMETRY_RETENTION_HOURS", 24)
	if hours < 1 {
		hours = 1
	}
	if hours > 720 {
		hours = 720
	}
	return time.Duration(hours) * time.Hour
}

func telemetryRejectScore(ctx context.Context, pool *pgxpool.Pool) int {
	score := appsettings.NewStore(pool).Int(ctx, "TELEMETRY_FINGERPRINT_REJECT_SCORE", 70)
	if score < 1 {
		return 1
	}
	if score > 100 {
		return 100
	}
	return score
}

func telemetrySecret(settings config.Settings) []byte {
	secret := strings.TrimSpace(settings.WebAppSessionSecret)
	if secret == "" {
		secret = strings.TrimSpace(settings.BotToken)
	}
	return []byte("telemetry-v1:" + secret)
}

func telemetryDigest(settings config.Settings, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	mac := hmac.New(sha256.New, telemetrySecret(settings))
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

func visitorCookie(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie(visitorCookieName); err == nil && len(cookie.Value) >= 32 {
		return cookie.Value
	}
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return ""
	}
	value := hex.EncodeToString(raw)
	http.SetCookie(w, &http.Cookie{Name: visitorCookieName, Value: value, Path: "/", MaxAge: 86400 * 30, HttpOnly: true, Secure: requestIsHTTPS(r), SameSite: http.SameSiteLaxMode}) //nolint:gosec // G124: attributes set dynamically.
	return value
}

func requestIsHTTPS(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https")
}

func networkBucket(raw string) string {
	host := strings.TrimSpace(raw)
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return "unknown"
	}
	if v4 := ip.To4(); v4 != nil {
		return fmt.Sprintf("%d.%d.%d.0/24", v4[0], v4[1], v4[2])
	}
	masked := ip.Mask(net.CIDRMask(64, 128))
	return masked.String() + "/64"
}

func optionalSessionUserID(r *http.Request, settings config.Settings) int64 {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return 0
	}
	claims, err := webappSessionManager(settings).Verify(cookie.Value)
	if err != nil {
		return 0
	}
	return claims.UserID
}

func telemetryHeartbeatHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if pool == nil || !telemetryEnabled(r.Context(), pool) {
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "enabled": false})
			return
		}
		var payload telemetryHeartbeatPayload
		if err := decodeJSONBody(r, &payload); err != nil || strings.TrimSpace(payload.Fingerprint.Full) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_fingerprint"})
			return
		}
		visitor := visitorCookie(w, r)
		if visitor == "" {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "visitor_cookie_failed"})
			return
		}
		fp := hashTelemetryFingerprint(settings, visitor, networkBucket(clientIP(r)), payload.Fingerprint)
		userID := optionalSessionUserID(r, settings)

		// Try INSERT; on conflict only bump last_seen_at to avoid redundant writes
		// when the fingerprint data hasn't changed.
		insertTag, err := pool.Exec(r.Context(), `
INSERT INTO visitor_telemetry (visitor_hash, full_fingerprint_hash, canvas_hash, webgl_hash, fonts_hash, audio_hash,
 network_hash, browser_hash, platform_hash, timezone_hash, screen_hash, hardware_hash, language_hash, user_id)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NULLIF($14,0))
ON CONFLICT (visitor_hash) DO UPDATE SET last_seen_at=NOW()`,
			fp.Visitor, fp.Full, fp.Canvas, fp.WebGL, fp.Fonts, fp.Audio, fp.Network, fp.Browser, fp.Platform,
			fp.Timezone, fp.Screen, fp.Hardware, fp.Language, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "telemetry_save_failed"})
			return
		}

		// If this was an INSERT (new visitor), no further action needed.
		// If conflict (RowsAffected==0), check whether fingerprint fields changed
		// and only then perform a full update.
		if insertTag.RowsAffected() == 0 {
			updateFingerprintIfChanged(r.Context(), pool, fp, userID)
		}

		if userID != 0 {
			_, _ = pool.Exec(r.Context(), `INSERT INTO visitor_user_links(visitor_hash,user_id) VALUES($1,$2)
ON CONFLICT(visitor_hash,user_id) DO UPDATE SET last_seen_at=NOW()`, fp.Visitor, userID)
		}
		if code := normalizeInviteCode(payload.InviteCode); code != "" {
			recordInviteVisit(r.Context(), pool, code, fp)
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "enabled": true, "expires_in_hours": int(telemetryRetention(r.Context(), pool).Hours())})
	}
}

// updateFingerprintIfChanged loads the current stored fingerprint and only performs
// a full UPDATE when at least one field differs, avoiding redundant WAL writes.
func updateFingerprintIfChanged(ctx context.Context, pool *pgxpool.Pool, fp storedFingerprint, userID int64) {
	var current storedFingerprint
	err := pool.QueryRow(ctx, `SELECT COALESCE(full_fingerprint_hash,''), COALESCE(canvas_hash,''),
 COALESCE(webgl_hash,''), COALESCE(fonts_hash,''), COALESCE(audio_hash,''), COALESCE(network_hash,''),
 COALESCE(browser_hash,''), COALESCE(platform_hash,''), COALESCE(timezone_hash,''), COALESCE(screen_hash,''),
 COALESCE(hardware_hash,''), COALESCE(language_hash,'') FROM visitor_telemetry WHERE visitor_hash=$1`, fp.Visitor).
		Scan(&current.Full, &current.Canvas, &current.WebGL, &current.Fonts, &current.Audio,
			&current.Network, &current.Browser, &current.Platform, &current.Timezone,
			&current.Screen, &current.Hardware, &current.Language)
	if err != nil {
		return
	}
	// Skip update if nothing changed (apart from last_seen_at which was already bumped).
	if current.Full == fp.Full && current.Canvas == fp.Canvas && current.WebGL == fp.WebGL &&
		current.Fonts == fp.Fonts && current.Audio == fp.Audio && current.Network == fp.Network &&
		current.Browser == fp.Browser && current.Platform == fp.Platform && current.Timezone == fp.Timezone &&
		current.Screen == fp.Screen && current.Hardware == fp.Hardware && current.Language == fp.Language {
		return
	}
	_, _ = pool.Exec(ctx, `UPDATE visitor_telemetry SET full_fingerprint_hash=$2, canvas_hash=$3, webgl_hash=$4,
 fonts_hash=$5, audio_hash=$6, network_hash=$7, browser_hash=$8, platform_hash=$9, timezone_hash=$10,
 screen_hash=$11, hardware_hash=$12, language_hash=$13, user_id=COALESCE(NULLIF($14,0), visitor_telemetry.user_id)
 WHERE visitor_hash=$1`, fp.Visitor, fp.Full, fp.Canvas, fp.WebGL, fp.Fonts, fp.Audio, fp.Network,
		fp.Browser, fp.Platform, fp.Timezone, fp.Screen, fp.Hardware, fp.Language, userID)
}

func hashTelemetryFingerprint(settings config.Settings, visitor, network string, in telemetryFingerprint) storedFingerprint {
	return storedFingerprint{
		Visitor: telemetryDigest(settings, visitor), Full: telemetryDigest(settings, in.Full),
		Canvas: telemetryDigest(settings, in.Canvas), WebGL: telemetryDigest(settings, in.WebGL),
		Fonts: telemetryDigest(settings, in.Fonts), Audio: telemetryDigest(settings, in.Audio),
		Network: telemetryDigest(settings, network), Browser: telemetryDigest(settings, in.Browser),
		Platform: telemetryDigest(settings, in.Platform), Timezone: telemetryDigest(settings, in.Timezone),
		Screen: telemetryDigest(settings, in.Screen), Hardware: telemetryDigest(settings, in.Hardware),
		Language: telemetryDigest(settings, in.Language),
	}
}

func normalizeInviteCode(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "REF_")
	value = strings.TrimPrefix(value, "REF-")
	if len(value) > 128 {
		return ""
	}
	return value
}

func recordInviteVisit(ctx context.Context, pool *pgxpool.Pool, code string, fp storedFingerprint) {
	kind := "campaign"
	var exists bool
	_ = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE UPPER(referral_code)=UPPER($1))", code).Scan(&exists)
	if exists {
		kind = "referral"
	}
	var id int64
	err := pool.QueryRow(ctx, `SELECT visit_id FROM invite_visits WHERE kind=$1 AND UPPER(code)=UPPER($2)
 AND (visitor_hash=$3 OR fingerprint_hash=$4) AND last_seen_at > NOW() - INTERVAL '24 hours' LIMIT 1`, kind, code, fp.Visitor, fp.Full).Scan(&id)
	if err == nil {
		_, _ = pool.Exec(ctx, "UPDATE invite_visits SET last_seen_at=NOW(), visit_count=visit_count+1 WHERE visit_id=$1", id)
		return
	}
	_, _ = pool.Exec(ctx, `INSERT INTO invite_visits(kind,code,visitor_hash,fingerprint_hash) VALUES($1,$2,$3,$4)`, kind, code, fp.Visitor, fp.Full)
}

func fingerprintSimilarity(a, b storedFingerprint) int {
	if a.Full != "" && a.Full == b.Full {
		return 100
	}
	score := 0
	for _, item := range []struct {
		a, b   string
		weight int
	}{
		{a.Canvas, b.Canvas, 18}, {a.WebGL, b.WebGL, 14}, {a.Fonts, b.Fonts, 14}, {a.Audio, b.Audio, 12},
		{a.Network, b.Network, 12}, {a.Browser, b.Browser, 8}, {a.Platform, b.Platform, 6}, {a.Timezone, b.Timezone, 4},
		{a.Screen, b.Screen, 4}, {a.Hardware, b.Hardware, 4}, {a.Language, b.Language, 4},
	} {
		if item.a != "" && item.a == item.b {
			score += item.weight
		}
	}
	if score > 100 {
		return 100
	}
	return score
}

func currentRequestFingerprint(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, r *http.Request) (storedFingerprint, bool) {
	cookie, err := r.Cookie(visitorCookieName)
	if err != nil || cookie.Value == "" {
		return storedFingerprint{}, false
	}
	visitor := telemetryDigest(settings, cookie.Value)
	var fp storedFingerprint
	err = pool.QueryRow(ctx, `SELECT visitor_hash,full_fingerprint_hash,COALESCE(canvas_hash,''),COALESCE(webgl_hash,''),
 COALESCE(fonts_hash,''),COALESCE(audio_hash,''),COALESCE(network_hash,''),COALESCE(browser_hash,''),
 COALESCE(platform_hash,''),COALESCE(timezone_hash,''),COALESCE(screen_hash,''),COALESCE(hardware_hash,''),COALESCE(language_hash,'')
	 FROM visitor_telemetry WHERE visitor_hash=$1 AND last_seen_at > NOW()-make_interval(hours => $2)`, visitor, int(telemetryRetention(ctx, pool).Hours())).Scan(
		&fp.Visitor, &fp.Full, &fp.Canvas, &fp.WebGL, &fp.Fonts, &fp.Audio, &fp.Network, &fp.Browser, &fp.Platform, &fp.Timezone, &fp.Screen, &fp.Hardware, &fp.Language)
	return fp, err == nil
}

func evaluateWelcomeRisk(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, r *http.Request, userID, referrerID int64) (storedFingerprint, int, string) {
	current, ok := currentRequestFingerprint(ctx, settings, pool, r)
	if !ok {
		return storedFingerprint{}, 100, "missing_telemetry"
	}
	var sameDevice bool
	_ = pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM visitor_user_links links JOIN visitor_telemetry telemetry USING(visitor_hash) WHERE links.user_id=$1 AND
	 (telemetry.visitor_hash=$2 OR telemetry.full_fingerprint_hash=$3) AND links.last_seen_at > NOW()-make_interval(hours => $4))`, referrerID, current.Visitor, current.Full, int(telemetryRetention(ctx, pool).Hours())).Scan(&sameDevice)
	if sameDevice {
		return current, 100, "referrer_device_match"
	}
	rows, err := pool.Query(ctx, `SELECT vt.visitor_hash,vt.full_fingerprint_hash,COALESCE(vt.canvas_hash,''),COALESCE(vt.webgl_hash,''),
 COALESCE(vt.fonts_hash,''),COALESCE(vt.audio_hash,''),COALESCE(vt.network_hash,''),COALESCE(vt.browser_hash,''),
 COALESCE(vt.platform_hash,''),COALESCE(vt.timezone_hash,''),COALESCE(vt.screen_hash,''),COALESCE(vt.hardware_hash,''),COALESCE(vt.language_hash,'')
 FROM referral_welcome_claims c JOIN visitor_telemetry vt ON vt.visitor_hash=c.visitor_hash
	 WHERE c.status='applied' AND c.user_id<>$1 AND vt.last_seen_at > NOW()-make_interval(hours => $2)`, userID, int(telemetryRetention(ctx, pool).Hours()))
	if err != nil {
		return current, 100, "risk_check_failed"
	}
	defer rows.Close()
	maxScore := 0
	for rows.Next() {
		var candidate storedFingerprint
		if rows.Scan(&candidate.Visitor, &candidate.Full, &candidate.Canvas, &candidate.WebGL, &candidate.Fonts, &candidate.Audio, &candidate.Network, &candidate.Browser, &candidate.Platform, &candidate.Timezone, &candidate.Screen, &candidate.Hardware, &candidate.Language) != nil {
			continue
		}
		if candidate.Visitor == current.Visitor {
			return current, 100, "visitor_reused"
		}
		score := fingerprintSimilarity(current, candidate)
		if score > maxScore {
			maxScore = score
		}
	}
	if maxScore >= telemetryRejectScore(ctx, pool) {
		return current, maxScore, "fingerprint_similarity"
	}
	return current, maxScore, ""
}

// RunTelemetryMaintenance stores one local installation heartbeat and removes expired anonymous data.
func RunTelemetryMaintenance(ctx context.Context, settings config.Settings, pool *pgxpool.Pool) error {
	if pool == nil {
		return nil
	}
	retention := telemetryRetention(ctx, pool)
	if telemetryEnabled(ctx, pool) {
		var count int64
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
		version := readSmallBuildFile(".build-version", "dev")
		provenance := readSmallBuildFile(".build-provenance", "custom")
		locale := appsettings.NewStore(pool).String(ctx, "DEFAULT_LANGUAGE", settings.DefaultLanguage)
		userRange := userCountRange(count)

		// Only write heartbeat when data has changed to avoid redundant WAL.
		if heartbeatChanged(ctx, pool, version, provenance, locale, userRange) {
			_, _ = pool.Exec(ctx, `INSERT INTO installation_heartbeats(heartbeat_date,version,provenance,os,locale,user_count_range)
VALUES(CURRENT_DATE,$1,$2,$3,$4,$5) ON CONFLICT(heartbeat_date) DO UPDATE SET version=EXCLUDED.version,
 provenance=EXCLUDED.provenance, os=EXCLUDED.os, locale=EXCLUDED.locale, user_count_range=EXCLUDED.user_count_range, last_seen_at=NOW()`,
				version, provenance, runtime.GOOS, locale, userRange)
		}
	}
	cutoff := time.Now().Add(-retention)
	// Only NULL out values that are not already NULL to avoid redundant writes.
	_, _ = pool.Exec(ctx, `UPDATE referral_welcome_claims SET visitor_hash=NULL, fingerprint_hash=NULL WHERE updated_at < $1 AND (visitor_hash IS NOT NULL OR fingerprint_hash IS NOT NULL)`, cutoff)
	_, _ = pool.Exec(ctx, "DELETE FROM invite_visits WHERE last_seen_at < $1", cutoff)
	_, _ = pool.Exec(ctx, "DELETE FROM visitor_user_links WHERE last_seen_at < $1", cutoff)
	_, _ = pool.Exec(ctx, "DELETE FROM visitor_telemetry WHERE last_seen_at < $1", cutoff)
	_, _ = pool.Exec(ctx, "DELETE FROM installation_heartbeats WHERE last_seen_at < $1", cutoff)
	return nil
}

// heartbeatChanged returns true when the heartbeat data differs from the last stored entry.
func heartbeatChanged(ctx context.Context, pool *pgxpool.Pool, version, provenance, locale, userRange string) bool {
	var prevVersion, prevProvenance, prevLocale, prevRange string
	err := pool.QueryRow(ctx, `SELECT COALESCE(version,''), COALESCE(provenance,''), COALESCE(locale,''), COALESCE(user_count_range,'')
 FROM installation_heartbeats WHERE heartbeat_date=CURRENT_DATE`).Scan(&prevVersion, &prevProvenance, &prevLocale, &prevRange)
	if err != nil {
		return true // no existing row, insert needed
	}
	return prevVersion != version || prevProvenance != provenance || prevLocale != locale || prevRange != userRange
}

func readSmallBuildFile(path, fallback string) string {
	body, err := osReadFile(path)
	if err != nil {
		return fallback
	}
	value := strings.TrimSpace(string(body))
	if value == "" {
		return fallback
	}
	if len(value) > 128 {
		value = value[:128]
	}
	return value
}

var osReadFile = func(path string) ([]byte, error) { return os.ReadFile(path) }

func userCountRange(count int64) string {
	switch {
	case count == 0:
		return "0"
	case count < 10:
		return "1-9"
	case count < 100:
		return "10-99"
	case count < 1000:
		return "100-999"
	case count < 10000:
		return "1000-9999"
	default:
		return "10000+"
	}
}
