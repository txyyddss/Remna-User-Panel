package telegram

import "context"

// ChatBoost is the identity and expiration portion of a Telegram chat boost.
type ChatBoost struct {
	BoostID        string `json:"boost_id"`
	AddDate        int64  `json:"add_date"`
	ExpirationDate int64  `json:"expiration_date"`
}

// UserChatBoosts contains boosts added to the requested chat by one user.
type UserChatBoosts struct {
	Boosts []ChatBoost `json:"boosts"`
}

// GetUserChatBoosts requires the bot to be an administrator in the target chat.
func (c *Client) GetUserChatBoosts(ctx context.Context, chatID string, userID int64) (UserChatBoosts, error) {
	if err := validateMemberRequest(chatID, userID); err != nil {
		return UserChatBoosts{}, err
	}
	var result UserChatBoosts
	err := c.call(ctx, "getUserChatBoosts", memberRequest{ChatID: chatID, UserID: userID}, &result)
	return result, err
}
