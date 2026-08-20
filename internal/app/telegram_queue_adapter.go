package app

import (
	"context"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/integrations/telegram"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/upstreamqueue"
)

type queuedTelegram struct {
	client *telegram.Client
	queue  *upstreamqueue.Queue
}

func (a *queuedTelegram) GetMe(ctx context.Context) (telegram.User, error) {
	return upstreamqueue.Do(ctx, a.queue, a.client.GetMe)
}
func (a *queuedTelegram) SendMarkdownV2Message(ctx context.Context, chatID, replyID int64, body string) error {
	return upstreamqueue.Execute(ctx, a.queue, func(callCtx context.Context) error {
		return a.client.SendMarkdownV2Message(callCtx, chatID, replyID, body)
	})
}
func (a *queuedTelegram) CreateJoinRequestInvite(ctx context.Context, chatID, name string, expires time.Time) (*telegram.ChatInviteLink, error) {
	return upstreamqueue.Do(ctx, a.queue, func(callCtx context.Context) (*telegram.ChatInviteLink, error) {
		return a.client.CreateJoinRequestInvite(callCtx, chatID, name, expires)
	})
}
func (a *queuedTelegram) GetChatMember(ctx context.Context, chatID string, userID int64) (*telegram.ChatMember, error) {
	return upstreamqueue.Do(ctx, a.queue, func(callCtx context.Context) (*telegram.ChatMember, error) {
		return a.client.GetChatMember(callCtx, chatID, userID)
	})
}
func (a *queuedTelegram) ApproveJoinRequest(ctx context.Context, chatID string, userID int64) error {
	return upstreamqueue.Execute(ctx, a.queue, func(callCtx context.Context) error { return a.client.ApproveJoinRequest(callCtx, chatID, userID) })
}
func (a *queuedTelegram) RevokeInviteLink(ctx context.Context, chatID, link string) (*telegram.ChatInviteLink, error) {
	return upstreamqueue.Do(ctx, a.queue, func(callCtx context.Context) (*telegram.ChatInviteLink, error) {
		return a.client.RevokeInviteLink(callCtx, chatID, link)
	})
}
func (a *queuedTelegram) SetWebhook(ctx context.Context, config telegram.WebhookConfig) error {
	return upstreamqueue.Execute(ctx, a.queue, func(callCtx context.Context) error { return a.client.SetWebhook(callCtx, config) })
}
func (a *queuedTelegram) SetChatMenuButton(ctx context.Context, text, link string) error {
	return upstreamqueue.Execute(ctx, a.queue, func(callCtx context.Context) error { return a.client.SetChatMenuButton(callCtx, text, link) })
}
func (a *queuedTelegram) SetMyCommands(ctx context.Context, commands []telegram.BotCommand, scope telegram.BotCommandScope, language string) error {
	return upstreamqueue.Execute(ctx, a.queue, func(callCtx context.Context) error { return a.client.SetMyCommands(callCtx, commands, scope, language) })
}
