package affiliates

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	jobpayload "github.com/txyyddss/Remna-User-Panel/internal/outbox"
)

type NotificationSender interface {
	SendMarkdownV2Message(context.Context, int64, int64, string) error
}

type NotificationWorker struct{ sender NotificationSender }

func NewNotificationWorker(sender NotificationSender) *NotificationWorker {
	return &NotificationWorker{sender: sender}
}

func (w *NotificationWorker) HandleOutbox(ctx context.Context, job model.OutboxJob) error {
	switch job.Kind {
	case jobpayload.AffiliateSuccessKind:
		payload, err := jobpayload.DecodeAffiliateSuccess(job)
		if err != nil {
			return err
		}
		return w.sender.SendMarkdownV2Message(ctx, payload.ChatID, 0, formatSuccess(payload))
	case jobpayload.AffiliateTierUpgradeKind:
		payload, err := jobpayload.DecodeAffiliateTierUpgrade(job)
		if err != nil {
			return err
		}
		return w.sender.SendMarkdownV2Message(ctx, payload.ChatID, 0, formatUpgrade(payload))
	default:
		return fmt.Errorf("unsupported affiliate notification kind: %s", job.Kind)
	}
}

func formatSuccess(payload jobpayload.AffiliateSuccess) string {
	inviteeName := payload.InviteeName
	if strings.TrimSpace(inviteeName) == "" {
		inviteeName = "TX member"
		if payload.Locale == LocaleChinese {
			inviteeName = "TX 用户"
		}
	}
	if payload.Locale == LocaleChinese {
		return "🎉 *好友首充成功*\n\n好友：" + md(inviteeName) + "\n时间：" + md(payload.SettledAt) +
			"\n返佣：*" + md(formatTXB(payload.CommissionMinor)) + " TXB*\n结算等级：" + md(payload.TierName)
	}
	return "🎉 *Referral successful*\n\nMember: " + md(inviteeName) + "\nTime: " + md(payload.SettledAt) +
		"\nCommission: *" + md(formatTXB(payload.CommissionMinor)) + " TXB*\nSettlement tier: " + md(payload.TierName)
}

func formatUpgrade(payload jobpayload.AffiliateTierUpgrade) string {
	if payload.Locale == LocaleChinese {
		return "🚀 *等级提升*\n\n恭喜你升级至 *" + md(payload.TierName) + "*\n升级奖励：" + md(payload.RewardDescription)
	}
	return "🚀 *Tier upgraded*\n\nYou reached *" + md(payload.TierName) + "*\nUpgrade reward: " + md(payload.RewardDescription)
}

func FormatReferralWelcome(locale, inviter string) string {
	if strings.TrimSpace(inviter) == "" {
		inviter = "TX member"
		if locale == LocaleChinese {
			inviter = "TX 用户"
		}
	}
	if locale == LocaleChinese {
		return "👋 你由 *" + md(inviter) + "* 邀请加入 TX Carpool。"
	}
	return "👋 You were invited to TX Carpool by *" + md(inviter) + "*\\."
}

func formatTXB(minor int64) string {
	return strconv.FormatInt(minor/100, 10) + "." + fmt.Sprintf("%02d", minor%100)
}

func md(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(", "\\(", ")", "\\)",
		"~", "\\~", "`", "\\`", ">", "\\>", "#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|", "\\|",
		"{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!")
	return replacer.Replace(value)
}
