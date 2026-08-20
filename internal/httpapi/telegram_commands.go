package httpapi

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/activity"
	"github.com/txyyddss/Remna-User-Panel/internal/affiliates"
	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/botcommands"
	"github.com/txyyddss/Remna-User-Panel/internal/integrations/telegram"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

func (s *Server) processTelegramGroupMessage(ctx context.Context, message *telegram.Message) {
	if message == nil || message.From == nil || message.From.IsBot || message.Chat.ID == 0 {
		return
	}
	command, isCommand := botcommands.Parse(message.Text)
	groupID, configuredGroup := s.telegramGroupID(ctx)
	allowedChat := message.Chat.Type == "private" || configuredGroup && message.Chat.ID == groupID
	if isCommand {
		if allowedChat {
			s.processTelegramCommand(ctx, message, command)
		}
		return
	}
	if !configuredGroup || message.Chat.ID != groupID {
		return
	}
	now := time.Now().UTC()
	config, configErr := s.activityConfig(ctx)
	localDate := telegramGroupMessageDate(config.Timezone, configErr, now)
	if err := s.deps.Store.BufferGroupMessageFact(message.Chat.ID, message.MessageID, localDate, now); err != nil {
		s.deps.Logger.Warn("buffer raw Telegram group message", "chat_id", message.Chat.ID, "message_id", message.MessageID, "error", err)
	}
	user, err := s.deps.Store.UserByTelegramID(ctx, message.From.ID)
	if err != nil {
		return
	}
	if configErr != nil {
		s.deps.Logger.Warn("load group-message reward configuration", "telegram_id", message.From.ID, "error", configErr)
		return
	}
	if _, err := s.deps.Activity.RecordGroupMessage(ctx, user.ID, message.Chat.ID, message.MessageID, activity.GroupMessageRewardConfig{
		Timezone: config.Timezone, Threshold: config.GroupMessageThreshold, RewardMinor: config.GroupMessageRewardMinor,
	}); err != nil {
		s.deps.Logger.Warn("process Telegram group message reward", "telegram_id", message.From.ID, "message_id", message.MessageID, "error", err)
	}
}

func telegramGroupMessageDate(configuredZone string, configErr error, now time.Time) string {
	zoneName := configuredZone
	if configErr != nil || strings.TrimSpace(zoneName) == "" {
		zoneName = defaultActivityTimezone
	}
	location, err := time.LoadLocation(zoneName)
	if err != nil {
		location = time.UTC
	}
	return now.In(location).Format(time.DateOnly)
}

func (s *Server) processTelegramCommand(ctx context.Context, message *telegram.Message, command botcommands.Command) {
	copy := botcommands.Text(botcommands.LanguageFor(message.From.LanguageCode))
	if command.Name == botcommands.Deduct {
		s.processTelegramDeduction(ctx, message, command, copy)
		return
	}
	if !command.Known || len(command.Args) > 0 && command.Name != botcommands.Start {
		s.sendTelegramReply(ctx, message, botcommands.FormatUnknown(copy))
		return
	}
	if command.Name == botcommands.Start {
		if message.Chat.Type == "private" && len(command.Args) == 1 {
			inviterID, err := parsePositiveTelegramID(command.Args[0])
			if err == nil && inviterID > 0 {
				name, accepted, acceptErr := s.deps.Affiliates.AcceptReferral(ctx, message.From.ID, inviterID)
				if acceptErr != nil {
					s.deps.Logger.Warn("accept Telegram referral", "telegram_id", message.From.ID, "error", acceptErr)
				}
				if accepted {
					s.sendTelegramReply(ctx, message, affiliates.FormatReferralWelcome(affiliates.NormalizeLocale(message.From.LanguageCode), name)+"\n\n"+botcommands.FormatStart(copy))
					return
				}
			}
		}
		s.sendTelegramReply(ctx, message, botcommands.FormatStart(copy))
		return
	}
	target := message.From
	if command.AllowsReplyTarget() && message.ReplyToMessage != nil && message.ReplyToMessage.From != nil && !message.ReplyToMessage.From.IsBot {
		target = message.ReplyToMessage.From
	}
	user, err := s.deps.Store.UserByTelegramID(ctx, target.ID)
	if err != nil {
		s.sendTelegramReply(ctx, message, botcommands.FormatUnavailable(copy))
		return
	}
	var reply string
	switch command.Name {
	case botcommands.Balance:
		money, balanceErr := s.deps.Store.Balance(ctx, user.ID)
		if balanceErr == nil {
			reply = botcommands.FormatBalance(copy, money)
		}
	case botcommands.SignIn:
		config, configErr := s.activityConfig(ctx)
		if configErr == nil {
			result, checkInErr := s.deps.Activity.CheckIn(ctx, user.ID, config)
			if checkInErr == nil {
				reply = botcommands.FormatCheckIn(copy, result, s.telegramCheckInAverage(ctx))
			}
		}
	case botcommands.Sub:
		reply = s.telegramSubscriptionReply(ctx, user, copy)
	case botcommands.MyCombo:
		reply = s.telegramComboReply(ctx, user, copy)
	}
	if strings.TrimSpace(reply) == "" {
		reply = botcommands.FormatUnavailable(copy)
	}
	s.sendTelegramReply(ctx, message, reply)
}

func (s *Server) sendTelegramReply(ctx context.Context, message *telegram.Message, reply string) {
	if err := s.deps.Telegram.SendMarkdownV2Message(ctx, message.Chat.ID, message.MessageID, botcommands.Limit(reply)); err != nil {
		s.deps.Logger.Warn("send Telegram command reply", "chat_id", message.Chat.ID, "message_id", message.MessageID, "error", err)
	}
}

func (s *Server) telegramGroupID(ctx context.Context) (int64, bool) {
	value, err := s.deps.Settings.Optional(ctx, "telegram.group_chat_id")
	if err != nil {
		s.deps.Logger.Warn("load Telegram group chat for message processing", "error", err)
		return 0, false
	}
	groupID, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return groupID, err == nil && groupID != 0
}

func (s *Server) processTelegramDeduction(ctx context.Context, message *telegram.Message, command botcommands.Command, copy botcommands.Copy) {
	if len(command.Args) != 1 {
		s.sendTelegramReply(ctx, message, botcommands.FormatDeductUsage(copy))
		return
	}
	amount, err := billing.ParseTXBMajor(command.Args[0])
	if err != nil || amount <= 0 {
		s.sendTelegramReply(ctx, message, botcommands.FormatDeductUsage(copy))
		return
	}
	if message.From == nil || !s.isAdminTelegramID(message.From.ID) || message.ReplyToMessage == nil || message.ReplyToMessage.From == nil || message.ReplyToMessage.From.IsBot {
		s.sendTelegramReply(ctx, message, botcommands.FormatDeductRejected(copy))
		return
	}
	actor, err := s.deps.Store.UserByTelegramID(ctx, message.From.ID)
	if err != nil {
		s.deps.Logger.Warn("load Telegram administrator for deduction", "error", err)
		s.sendTelegramReply(ctx, message, botcommands.FormatDeductRejected(copy))
		return
	}
	target, err := s.deps.Store.UserByTelegramID(ctx, message.ReplyToMessage.From.ID)
	if err != nil {
		s.sendTelegramReply(ctx, message, botcommands.FormatDeductRejected(copy))
		return
	}
	reason := fmt.Sprintf("Telegram /deduct in chat %d on message %d", message.Chat.ID, message.ReplyToMessage.MessageID)
	if _, err := s.deps.Admin.DeductBalance(ctx, actor.ID, target.ID, amount, reason); err != nil {
		s.deps.Logger.Warn("Telegram TXB deduction rejected", "actor_id", actor.ID, "target_id", target.ID, "amount_minor", amount, "error", err)
		s.sendTelegramReply(ctx, message, botcommands.FormatDeductRejected(copy))
		return
	}
	s.sendTelegramReply(ctx, message, botcommands.FormatDeductSucceeded(copy, model.TXBMoney(amount)))
}

func telegramDeductCommand(text string) (string, bool) {
	command, ok := botcommands.Parse(text)
	if !ok || command.Name != botcommands.Deduct || len(command.Args) != 1 {
		return "", false
	}
	amount, err := billing.ParseTXBMajor(command.Args[0])
	return command.Args[0], err == nil && amount > 0
}
