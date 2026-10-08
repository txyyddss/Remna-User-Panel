package telegrampm

import (
	"context"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/integrations/telegram"
	"github.com/txyyddss/Remna-User-Panel/internal/providerops"
)

func messageReference(value string) (int64, int64, error) {
	chat, message, ok := strings.Cut(value, ":")
	if !ok {
		return 0, 0, ErrBlocked
	}
	chatID, err := strconv.ParseInt(chat, 10, 64)
	if err != nil {
		return 0, 0, err
	}
	messageID, err := strconv.ParseInt(message, 10, 64)
	if err != nil || chatID == 0 || messageID <= 0 {
		return 0, 0, ErrBlocked
	}
	return chatID, messageID, nil
}

func (w *Worker) relay(ctx context.Context, run execution, item providerops.Item, interrupted bool) (phaseResult, error) {
	if interrupted {
		return phaseResult{status: providerops.StatusPendingReview, code: "PM_DELIVERY_UNCERTAIN"}, nil
	}
	sourceChat, messageID, err := messageReference(item.TargetID)
	if err != nil {
		return phaseResult{}, err
	}
	var started atomic.Bool
	var topicID atomic.Int64
	result, callErr := w.Sender.CopyPMMessage(ctx, func(callCtx context.Context) (telegram.CopyMessageRequest, error) {
		conversation, err := w.current(callCtx, run)
		if err != nil {
			return telegram.CopyMessageRequest{}, err
		}
		if conversation.TopicState != "ready" || conversation.TopicID <= 1 {
			return telegram.CopyMessageRequest{}, ErrTopicUncertain
		}
		request := telegram.CopyMessageRequest{FromChatID: sourceChat, MessageID: messageID, DisableNotification: conversation.Muted}
		if run.inbound {
			if sourceChat != conversation.TelegramID {
				return request, ErrBlocked
			}
			request.ChatID = conversation.ChatID
			request.MessageThreadID = conversation.TopicID
		} else {
			if sourceChat != conversation.ChatID {
				return request, ErrBlocked
			}
			request.ChatID = conversation.TelegramID
		}
		topicID.Store(conversation.TopicID)
		started.Store(true)
		return request, nil
	})
	if run.inbound && topicMissing(callErr) && run.job.Attempts < 10 {
		if err := w.Repository.ResetMissingPMTopic(ctx, run.conversationID, run.operation.Receipt.ID, topicID.Load(), time.Now().UTC()); err != nil {
			return phaseResult{}, err
		}
		return phaseResult{}, callErr
	}
	return w.callResult(ctx, run, item, callErr, started.Load(), true, strconv.FormatInt(result, 10), "PM_DELIVERY_UNCERTAIN")
}
