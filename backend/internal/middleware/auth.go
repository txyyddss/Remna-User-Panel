package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/user/remna-user-panel/internal/config"
	"github.com/user/remna-user-panel/internal/database"
	"github.com/user/remna-user-panel/internal/models"
)

type contextKey string

const (
	UserContextKey      contextKey = "user"
	TokenPermissionsKey contextKey = "token_permissions"
)

// GetUser retrieves the authenticated user from context
func GetUser(r *http.Request) *models.User {
	if u, ok := r.Context().Value(UserContextKey).(*models.User); ok {
		return u
	}
	return nil
}

// TelegramAuth validates Telegram Mini App initData and sets user in context
func TelegramAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		initData := r.Header.Get("X-Telegram-Init-Data")
		if initData == "" {
			// Try query param
			initData = r.URL.Query().Get("initData")
		}
		if initData == "" {
			WriteError(w, http.StatusUnauthorized, "missing telegram init data")
			return
		}

		telegramID, name, err := validateInitData(initData)
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "invalid telegram auth: "+err.Error())
			return
		}

		// Find or create user
		user, err := getOrCreateUser(r.Context(), telegramID, name)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "failed to load user")
			return
		}

		if err := ensureGroupMembership(r.Context(), user); err != nil {
			WriteError(w, http.StatusForbidden, err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// APITokenAuth validates API token from Authorization header
func APITokenAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")
		if token == "" {
			WriteError(w, http.StatusUnauthorized, "missing api token")
			return
		}

		hash := hashToken(token)
		var apiToken models.APIToken
		err := database.DB().QueryRowContext(r.Context(),
			"SELECT id, token_hash, name, permissions, created_by FROM api_tokens WHERE token_hash = ?",
			hash,
		).Scan(&apiToken.ID, &apiToken.TokenHash, &apiToken.Name, &apiToken.Permissions, &apiToken.CreatedBy)
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "invalid api token")
			return
		}

		// Update last used
		database.DB().ExecContext(r.Context(), "UPDATE api_tokens SET last_used_at = ? WHERE id = ?", time.Now(), apiToken.ID)

		var user models.User
		err = database.DB().QueryRowContext(r.Context(),
			"SELECT id, telegram_id, telegram_name, remnawave_uuid, jellyfin_user_id, credit, is_admin, created_at, updated_at FROM users WHERE id = ?",
			apiToken.CreatedBy,
		).Scan(&user.ID, &user.TelegramID, &user.TelegramName, &user.RemnawaveUUID, &user.JellyfinUserID, &user.Credit, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "invalid token owner")
			return
		}

		permissions := parsePermissions(apiToken.Permissions)
		user.IsAdmin = user.IsAdmin && hasAnyPermission(permissions, "*", "admin")

		ctx := context.WithValue(r.Context(), UserContextKey, &user)
		ctx = context.WithValue(ctx, TokenPermissionsKey, permissions)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminOnly ensures the user is an admin
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r)
		if user == nil || !user.IsAdmin {
			WriteError(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}

// CombinedAuth tries Telegram auth first, then API token
func CombinedAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		initData := r.Header.Get("X-Telegram-Init-Data")
		if initData != "" {
			TelegramAuth(next).ServeHTTP(w, r)
			return
		}
		token := r.Header.Get("Authorization")
		if token != "" {
			APITokenAuth(next).ServeHTTP(w, r)
			return
		}
		WriteError(w, http.StatusUnauthorized, "authentication required")
	})
}

func validateInitData(initData string) (int64, string, error) {
	cfg := config.Get()
	botToken := cfg.Telegram.BotToken
	if botToken == "" {
		return 0, "", fmt.Errorf("bot token not configured")
	}

	// Parse the data
	values, err := url.ParseQuery(initData)
	if err != nil {
		return 0, "", fmt.Errorf("parse init data: %w", err)
	}

	hash := values.Get("hash")
	if hash == "" {
		return 0, "", fmt.Errorf("missing hash")
	}

	// Build check string: sort keys, exclude hash
	var keys []string
	for k := range values {
		if k != "hash" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+values.Get(k))
	}
	checkString := strings.Join(parts, "\n")

	// Compute HMAC
	secretKey := hmacSHA256([]byte("WebAppData"), []byte(botToken))
	computed := hex.EncodeToString(hmacSHA256(secretKey, []byte(checkString)))

	if computed != hash {
		return 0, "", fmt.Errorf("hash mismatch")
	}

	// Validate auth_date freshness (allow 24 hours)
	authDate := values.Get("auth_date")
	if authDate != "" {
		var ts int64
		fmt.Sscanf(authDate, "%d", &ts)
		if time.Now().Unix()-ts > 86400 {
			return 0, "", fmt.Errorf("init data expired")
		}
	}

	// Parse user
	userData := values.Get("user")
	if userData == "" {
		return 0, "", fmt.Errorf("missing user data")
	}

	var tgUser struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
	}
	if err := json.Unmarshal([]byte(userData), &tgUser); err != nil {
		return 0, "", fmt.Errorf("parse user data: %w", err)
	}

	name := tgUser.FirstName
	if tgUser.LastName != "" {
		name += " " + tgUser.LastName
	}
	if name == "" {
		name = tgUser.Username
	}

	return tgUser.ID, name, nil
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func parsePermissions(raw string) []string {
	if raw == "" {
		return nil
	}

	var permissions []string
	if err := json.Unmarshal([]byte(raw), &permissions); err != nil {
		return nil
	}
	return permissions
}

func hasAnyPermission(permissions []string, required ...string) bool {
	for _, permission := range permissions {
		for _, candidate := range required {
			if permission == candidate {
				return true
			}
		}
	}
	return false
}

func getOrCreateUser(ctx context.Context, telegramID int64, name string) (*models.User, error) {
	var user models.User
	err := database.DB().QueryRowContext(ctx,
		"SELECT id, telegram_id, telegram_name, remnawave_uuid, jellyfin_user_id, credit, is_admin, created_at, updated_at FROM users WHERE telegram_id = ?",
		telegramID,
	).Scan(&user.ID, &user.TelegramID, &user.TelegramName, &user.RemnawaveUUID, &user.JellyfinUserID, &user.Credit, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		// Check if user is admin
		cfg := config.Get()
		isAdmin := false
		for _, id := range cfg.Telegram.AdminIDs {
			if id == telegramID {
				isAdmin = true
				break
			}
		}

		adminInt := 0
		if isAdmin {
			adminInt = 1
		}

		result, err := database.DB().ExecContext(ctx,
			"INSERT INTO users (telegram_id, telegram_name, is_admin) VALUES (?, ?, ?)",
			telegramID, name, adminInt,
		)
		if err != nil {
			return nil, err
		}
		id, _ := result.LastInsertId()
		user = models.User{
			ID:           id,
			TelegramID:   telegramID,
			TelegramName: name,
			Credit:       0,
			IsAdmin:      isAdmin,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
	} else {
		// Update name if changed
		if user.TelegramName != name {
			database.DB().ExecContext(ctx, "UPDATE users SET telegram_name = ?, updated_at = ? WHERE id = ?", name, time.Now(), user.ID)
			user.TelegramName = name
		}
		// Check if admin status should update
		cfg := config.Get()
		for _, id := range cfg.Telegram.AdminIDs {
			if id == telegramID && !user.IsAdmin {
				database.DB().ExecContext(ctx, "UPDATE users SET is_admin = 1 WHERE id = ?", user.ID)
				user.IsAdmin = true
			}
		}
	}

	return &user, nil
}

func ensureGroupMembership(ctx context.Context, user *models.User) error {
	if user == nil || user.IsAdmin {
		return nil
	}

	cfg := config.Get()
	if cfg == nil || cfg.Telegram.BotToken == "" || cfg.Telegram.GroupID == 0 {
		return nil
	}

	status, err := fetchChatMemberStatus(ctx, cfg.Telegram.BotToken, cfg.Telegram.GroupID, user.TelegramID)
	if err != nil {
		return fmt.Errorf("failed to verify group membership")
	}
	if !isJoinedChatStatus(status) {
		return fmt.Errorf("group membership required to use this mini app")
	}
	return nil
}

func fetchChatMemberStatus(ctx context.Context, botToken string, chatID, userID int64) (string, error) {
	body, err := json.Marshal(map[string]int64{
		"chat_id": chatID,
		"user_id": userID,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.telegram.org/bot"+botToken+"/getChatMember", strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var payload struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
		Result      struct {
			Status string `json:"status"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", err
	}
	if !payload.OK {
		return "", fmt.Errorf(payload.Description)
	}
	return payload.Result.Status, nil
}

func isJoinedChatStatus(status string) bool {
	switch status {
	case "member", "administrator", "creator", "restricted":
		return true
	default:
		return false
	}
}
