package telegram

import (
	"context"
	"errors"
	"unicode/utf8"
)

// SetChatMemberTag applies the Bot API regular-member tag, including clearing.
func (c *Client) SetChatMemberTag(ctx context.Context, chatID string, userID int64, tag string) error {
	if err := validateMemberRequest(chatID, userID); err != nil {
		return err
	}
	if !utf8.ValidString(tag) || utf8.RuneCountInString(tag) > 16 {
		return errors.New("telegram member tag exceeds 16 characters")
	}
	return c.booleanCall(ctx, "setChatMemberTag", struct {
		ChatID string `json:"chat_id"`
		UserID int64  `json:"user_id"`
		Tag    string `json:"tag"`
	}{chatID, userID, tag})
}
