// Package telegram wraps the Telegram Bot API client used by the backend.
package telegram

import (
	"fmt"

	tgbot "github.com/go-telegram/bot"

	"remna-user-panel/internal/config"
)

// NewBot creates a Telegram bot client configured for webhook secret verification.
func NewBot(settings config.Settings) (*tgbot.Bot, error) {
	if settings.BotToken == "" {
		return nil, fmt.Errorf("BOT_TOKEN is required")
	}
	options := []tgbot.Option{}
	if settings.WebhookSecretToken != "" {
		options = append(options, tgbot.WithWebhookSecretToken(settings.WebhookSecretToken))
	}
	bot, err := tgbot.New(settings.BotToken, options...)
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}
	return bot, nil
}
