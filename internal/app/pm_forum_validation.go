package app

import (
	"context"
	"errors"
	"strconv"
	"time"
)

func validatePMForum(client *queuedTelegram) func(context.Context, int64) error {
	return func(ctx context.Context, chatID int64) error {
		ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		chat, err := client.GetChat(ctx, chatID)
		if err != nil {
			return err
		}
		if chat.ID != chatID || chat.Type != "supergroup" || !chat.IsForum {
			return errors.New("PM destination is not a forum supergroup")
		}
		bot, err := client.GetMe(ctx)
		if err != nil {
			return err
		}
		member, err := client.GetChatMember(ctx, strconv.FormatInt(chatID, 10), bot.ID)
		if err != nil {
			return err
		}
		if member == nil || member.User.ID != bot.ID || (member.Status != "creator" && (member.Status != "administrator" || !member.CanManageTopics)) {
			return errors.New("PM bot requires forum administrator and manage-topics permissions")
		}
		return nil
	}
}
