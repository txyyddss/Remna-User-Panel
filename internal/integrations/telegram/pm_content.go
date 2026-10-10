package telegram

import (
	"context"
	"errors"
	"unicode/utf16"
)

// MessageEntity preserves Telegram's UTF-16 offsets and optional entity data.
type MessageEntity struct {
	Type          string `json:"type"`
	Offset        int    `json:"offset"`
	Length        int    `json:"length"`
	URL           string `json:"url,omitempty"`
	User          *User  `json:"user,omitempty"`
	Language      string `json:"language,omitempty"`
	CustomEmojiID string `json:"custom_emoji_id,omitempty"`
}

// PMTextRequest sends an attributed text or a separately tracked footer reply.
type PMTextRequest struct {
	ChatID              int64              `json:"chat_id"`
	Text                string             `json:"text"`
	Entities            []MessageEntity    `json:"entities,omitempty"`
	DisableNotification bool               `json:"disable_notification"`
	ReplyParameters     *PMReplyParameters `json:"reply_parameters,omitempty"`
}

// PMReplyParameters links an exception footer to the copied message.
type PMReplyParameters struct {
	MessageID int64 `json:"message_id"`
}

// UTF16Length matches Telegram entity offsets and conservatively bounds text limits.
func UTF16Length(text string) int { return len(utf16.Encode([]rune(text))) }

// SendPMText preserves entities without parse-mode reinterpretation or truncation.
func (c *Client) SendPMText(ctx context.Context, input PMTextRequest) (int64, error) {
	if input.ChatID <= 0 || input.Text == "" || UTF16Length(input.Text) > 4096 {
		return 0, errors.New("invalid attributed PM text")
	}
	if input.ReplyParameters != nil && input.ReplyParameters.MessageID <= 0 {
		return 0, errors.New("invalid PM reply reference")
	}
	var result Message
	err := c.call(ctx, "sendMessage", input, &result)
	if err == nil && (result.MessageID <= 0 || result.Chat.ID != input.ChatID) {
		return 0, errors.New("invalid attributed PM response identity")
	}
	return result.MessageID, err
}
