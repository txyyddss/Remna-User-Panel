package telegram

import "context"

// ChatBoost is the identity and expiration portion of a Telegram chat boost.
type ChatBoost struct {
	BoostID        string          `json:"boost_id"`
	AddDate        int64           `json:"add_date"`
	ExpirationDate int64           `json:"expiration_date"`
	Source         ChatBoostSource `json:"source"`
}

// ChatBoostSource identifies the booster when Telegram supplies their identity.
// Giveaway boosts may omit User; no upstream identity is guessed in that case.
type ChatBoostSource struct {
	Source string `json:"source"`
	User   *User  `json:"user,omitempty"`
}

// ChatBoostUpdated is an added or changed boost in an administrator-owned chat.
type ChatBoostUpdated struct {
	Chat  Chat      `json:"chat"`
	Boost ChatBoost `json:"boost"`
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
