package telegram

import (
	"context"
	"errors"
	"strings"
	"time"
)

// CreateJoinRequestInvite creates a revocable invite which requires administrator approval.
func (c *Client) CreateJoinRequestInvite(ctx context.Context, chatID, name string, expireAt time.Time) (*ChatInviteLink, error) {
	if strings.TrimSpace(chatID) == "" {
		return nil, errors.New("telegram chat id is empty")
	}
	if expireAt.IsZero() {
		return nil, errors.New("telegram invite expiration is empty")
	}
	if len([]rune(name)) > 32 {
		return nil, errors.New("telegram invite name exceeds 32 characters")
	}
	payload := struct {
		ChatID             string `json:"chat_id"`
		Name               string `json:"name,omitempty"`
		ExpireDate         int64  `json:"expire_date"`
		CreatesJoinRequest bool   `json:"creates_join_request"`
	}{ChatID: chatID, Name: name, ExpireDate: expireAt.Unix(), CreatesJoinRequest: true}
	var result ChatInviteLink
	if err := c.call(ctx, "createChatInviteLink", payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ApproveJoinRequest approves a single Telegram user for a chat.
func (c *Client) ApproveJoinRequest(ctx context.Context, chatID string, userID int64) error {
	return c.booleanCall(ctx, "approveChatJoinRequest", memberRequest{ChatID: chatID, UserID: userID})
}

// RevokeInviteLink revokes a bot-created chat invite link.
func (c *Client) RevokeInviteLink(ctx context.Context, chatID, inviteLink string) (*ChatInviteLink, error) {
	if strings.TrimSpace(chatID) == "" || strings.TrimSpace(inviteLink) == "" {
		return nil, errors.New("telegram chat id and invite link are required")
	}
	payload := struct {
		ChatID     string `json:"chat_id"`
		InviteLink string `json:"invite_link"`
	}{ChatID: chatID, InviteLink: inviteLink}
	var result ChatInviteLink
	if err := c.call(ctx, "revokeChatInviteLink", payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetChatMember fetches canonical membership state for a user.
func (c *Client) GetChatMember(ctx context.Context, chatID string, userID int64) (*ChatMember, error) {
	if err := validateMemberRequest(chatID, userID); err != nil {
		return nil, err
	}
	var result ChatMember
	if err := c.call(ctx, "getChatMember", memberRequest{ChatID: chatID, UserID: userID}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type memberRequest struct {
	ChatID string `json:"chat_id"`
	UserID int64  `json:"user_id"`
}

func validateMemberRequest(chatID string, userID int64) error {
	if strings.TrimSpace(chatID) == "" || userID <= 0 {
		return errors.New("telegram chat id and positive user id are required")
	}
	return nil
}
