package httpapi

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/txyyddss/Remna-User-Panel/internal/activity"
	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/integrations/telegram"
)

func (s *Server) processTelegramGroupMessage(ctx context.Context, message *telegram.Message) {
	if message == nil || message.From == nil || message.From.IsBot || message.Chat.ID == 0 {
		return
	}
	groupIDValue, err := s.deps.Settings.Optional(ctx, "telegram.group_chat_id")
	if err != nil {
		s.deps.Logger.Warn("load Telegram group chat for message processing", "error", err)
		return
	}
	groupID, err := strconv.ParseInt(groupIDValue, 10, 64)
	if err != nil || groupID == 0 || message.Chat.ID != groupID {
		return
	}
	if amountText, ok := telegramDeductCommand(message.Text); ok {
		s.processTelegramDeduction(ctx, message, amountText)
		return
	}
	user, err := s.deps.Store.UserByTelegramID(ctx, message.From.ID)
	if err != nil {
		return
	}
	config, err := s.activityConfig(ctx)
	if err != nil {
		s.deps.Logger.Warn("load group-message reward configuration", "telegram_id", message.From.ID, "error", err)
		return
	}
	if _, err := s.deps.Activity.RecordGroupMessage(ctx, user.ID, message.Chat.ID, message.MessageID, activity.GroupMessageRewardConfig{
		Timezone: config.Timezone, Threshold: config.GroupMessageThreshold, RewardMinor: config.GroupMessageRewardMinor,
	}); err != nil {
		s.deps.Logger.Warn("process Telegram group message reward", "telegram_id", message.From.ID, "message_id", message.MessageID, "error", err)
	}
}

func (s *Server) processTelegramDeduction(ctx context.Context, message *telegram.Message, amountText string) {
	if message.From == nil || message.From.ID != s.deps.AdminTelegramID || message.ReplyToMessage == nil || message.ReplyToMessage.From == nil || message.ReplyToMessage.From.IsBot {
		return
	}
	amount, err := billing.ParseTXBMajor(amountText)
	if err != nil || amount <= 0 {
		return
	}
	actor, err := s.deps.Store.UserByTelegramID(ctx, s.deps.AdminTelegramID)
	if err != nil {
		s.deps.Logger.Warn("load Telegram administrator for deduction", "error", err)
		return
	}
	target, err := s.deps.Store.UserByTelegramID(ctx, message.ReplyToMessage.From.ID)
	if err != nil {
		return
	}
	reason := fmt.Sprintf("Telegram /deduct in chat %d on message %d", message.Chat.ID, message.ReplyToMessage.MessageID)
	if _, err := s.deps.Admin.DeductBalance(ctx, actor.ID, target.ID, amount, reason); err != nil {
		s.deps.Logger.Warn("Telegram TXB deduction rejected", "actor_id", actor.ID, "target_id", target.ID, "amount_minor", amount, "error", err)
	}
}

func telegramDeductCommand(text string) (string, bool) {
	fields := strings.Fields(text)
	if len(fields) != 2 {
		return "", false
	}
	command := strings.ToLower(fields[0])
	if at := strings.IndexByte(command, '@'); at >= 0 {
		command = command[:at]
	}
	if command != "/deduct" {
		return "", false
	}
	amount, err := billing.ParseTXBMajor(fields[1])
	if err != nil || amount <= 0 {
		return "", false
	}
	return fields[1], true
}
