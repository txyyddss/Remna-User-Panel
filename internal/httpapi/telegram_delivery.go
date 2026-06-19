package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"remna-user-panel/internal/config"
)

func sendTelegramText(ctx context.Context, settings config.Settings, chatID int64, text string) error {
	if strings.TrimSpace(settings.BotToken) == "" {
		return fmt.Errorf("telegram_bot_not_configured")
	}
	if chatID == 0 {
		return fmt.Errorf("telegram_chat_not_found")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("empty_message")
	}
	payload := map[string]any{
		"chat_id":                  chatID,
		"text":                     text,
		"disable_web_page_preview": true,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	endpoint := "https://api.telegram.org/bot" + strings.TrimSpace(settings.BotToken) + "/sendMessage"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram_status_%d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if json.Unmarshal(responseBody, &result) == nil && !result.OK {
		if result.Description != "" {
			return fmt.Errorf("telegram_api_error: %s", result.Description)
		}
		return fmt.Errorf("telegram_api_error")
	}
	return nil
}
