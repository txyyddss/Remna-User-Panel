package telegram

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

const telegramMessageLimit = 4096

// BotCommand is one command advertised by Telegram's command menu.
type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

// BotCommandScope selects the chats receiving a command menu.
type BotCommandScope struct {
	Type   string `json:"type"`
	ChatID int64  `json:"chat_id,omitempty"`
}

// SetMyCommands installs one localized command list for a Bot API scope.
func (c *Client) SetMyCommands(ctx context.Context, commands []BotCommand, scope BotCommandScope, languageCode string) error {
	if len(commands) == 0 || len(commands) > 100 {
		return errors.New("telegram command list must contain 1-100 commands")
	}
	if scope.Type != "all_private_chats" && scope.Type != "chat" {
		return errors.New("telegram command scope is invalid")
	}
	if scope.Type == "chat" && scope.ChatID == 0 {
		return errors.New("telegram chat command scope requires a chat id")
	}
	for _, command := range commands {
		if !validCommand(command.Command) || strings.TrimSpace(command.Description) == "" || utf8.RuneCountInString(command.Description) > 256 {
			return errors.New("telegram command contains an invalid name or description")
		}
	}
	payload := struct {
		Commands     []BotCommand    `json:"commands"`
		Scope        BotCommandScope `json:"scope"`
		LanguageCode string          `json:"language_code,omitempty"`
	}{Commands: commands, Scope: scope, LanguageCode: strings.TrimSpace(languageCode)}
	return c.booleanCall(ctx, "setMyCommands", payload)
}

// SendMessage sends a bounded plain-text reply to a Telegram message.
func (c *Client) SendMessage(ctx context.Context, chatID, replyToMessageID int64, text string) error {
	text = strings.TrimSpace(text)
	if chatID == 0 || text == "" || utf8.RuneCountInString(text) > telegramMessageLimit {
		return errors.New("telegram reply requires a chat id and 1-4096 characters")
	}
	payload := struct {
		ChatID          int64  `json:"chat_id"`
		Text            string `json:"text"`
		ReplyParameters *struct {
			MessageID                int64 `json:"message_id"`
			AllowSendingWithoutReply bool  `json:"allow_sending_without_reply"`
		} `json:"reply_parameters,omitempty"`
	}{ChatID: chatID, Text: text}
	if replyToMessageID > 0 {
		payload.ReplyParameters = &struct {
			MessageID                int64 `json:"message_id"`
			AllowSendingWithoutReply bool  `json:"allow_sending_without_reply"`
		}{MessageID: replyToMessageID, AllowSendingWithoutReply: true}
	}
	return c.booleanCall(ctx, "sendMessage", payload)
}

func validCommand(value string) bool {
	if len(value) < 1 || len(value) > 32 {
		return false
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '_' {
			continue
		}
		return false
	}
	return true
}
