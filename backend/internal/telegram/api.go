package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func IsJoinedChatStatus(status string) bool {
	switch status {
	case "member", "administrator", "creator", "restricted":
		return true
	default:
		return false
	}
}

func GetChatMemberStatus(ctx context.Context, botToken string, chatID, userID int64) (string, error) {
	var payload struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
		Result      struct {
			Status string `json:"status"`
		} `json:"result"`
	}

	if err := doTelegramJSON(ctx, botToken, "getChatMember", map[string]int64{
		"chat_id": chatID,
		"user_id": userID,
	}, &payload); err != nil {
		return "", err
	}
	if !payload.OK {
		return "", fmt.Errorf(payload.Description)
	}
	return payload.Result.Status, nil
}

func CreateChatInviteLink(ctx context.Context, botToken string, chatID int64, memberLimit int) (string, error) {
	var payload struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
		Result      struct {
			InviteLink string `json:"invite_link"`
		} `json:"result"`
	}

	if err := doTelegramJSON(ctx, botToken, "createChatInviteLink", map[string]interface{}{
		"chat_id":      chatID,
		"member_limit": memberLimit,
	}, &payload); err != nil {
		return "", err
	}
	if !payload.OK {
		return "", fmt.Errorf(payload.Description)
	}
	if payload.Result.InviteLink == "" {
		return "", fmt.Errorf("telegram returned an empty invite link")
	}
	return payload.Result.InviteLink, nil
}

func RevokeChatInviteLink(ctx context.Context, botToken string, chatID int64, inviteLink string) error {
	var payload struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}

	if err := doTelegramJSON(ctx, botToken, "revokeChatInviteLink", map[string]interface{}{
		"chat_id":     chatID,
		"invite_link": inviteLink,
	}, &payload); err != nil {
		return err
	}
	if !payload.OK {
		return fmt.Errorf(payload.Description)
	}
	return nil
}

func doTelegramJSON(ctx context.Context, botToken, method string, body interface{}, out interface{}) error {
	rawBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.telegram.org/bot"+botToken+"/"+method, bytes.NewReader(rawBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode telegram %s response: %w", method, err)
	}
	return nil
}
