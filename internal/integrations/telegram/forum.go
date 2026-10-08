package telegram

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

type ForumTopic struct {
	MessageThreadID int64 `json:"message_thread_id"`
}
type ChatInfo struct {
	ID      int64  `json:"id"`
	Type    string `json:"type"`
	IsForum bool   `json:"is_forum"`
}

func (c *Client) GetChat(ctx context.Context, chatID int64) (ChatInfo, error) {
	if chatID == 0 {
		return ChatInfo{}, errors.New("chat id is required")
	}
	var result ChatInfo
	err := c.call(ctx, "getChat", struct {
		ChatID int64 `json:"chat_id"`
	}{chatID}, &result)
	return result, err
}

func (c *Client) CreateForumTopic(ctx context.Context, chatID int64, name string) (ForumTopic, error) {
	name = strings.TrimSpace(name)
	if chatID >= 0 || name == "" || utf8.RuneCountInString(name) > 128 {
		return ForumTopic{}, errors.New("invalid forum topic identity")
	}
	var topic ForumTopic
	err := c.call(ctx, "createForumTopic", struct {
		ChatID int64  `json:"chat_id"`
		Name   string `json:"name"`
	}{chatID, name}, &topic)
	if err == nil && topic.MessageThreadID <= 1 {
		return topic, errors.New("Telegram returned an invalid forum topic id")
	}
	return topic, err
}

type CopyMessageRequest struct {
	ChatID              int64 `json:"chat_id"`
	MessageThreadID     int64 `json:"message_thread_id,omitempty"`
	FromChatID          int64 `json:"from_chat_id"`
	MessageID           int64 `json:"message_id"`
	DisableNotification bool  `json:"disable_notification"`
}

// CopyMessage retains original formatting and captions by sending only references.
func (c *Client) CopyMessage(ctx context.Context, input CopyMessageRequest) (int64, error) {
	if input.ChatID == 0 || input.FromChatID == 0 || input.MessageID <= 0 || input.MessageThreadID < 0 {
		return 0, errors.New("invalid copyMessage references")
	}
	var result struct {
		MessageID int64 `json:"message_id"`
	}
	err := c.call(ctx, "copyMessage", input, &result)
	if err == nil && result.MessageID <= 0 {
		return 0, errors.New("Telegram returned an invalid copied message id")
	}
	return result.MessageID, err
}
