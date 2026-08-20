package botcommands

import (
	"strings"

	"github.com/txyyddss/Remna-User-Panel/internal/activity"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

// FormatBalance renders a self-only TXB balance reply without a title.
func FormatBalance(copy Copy, money model.Money) string {
	return Limit(escapeMarkdownV2(copy.BalanceLabel + ": " + money.Display))
}

// FormatCheckIn renders an idempotent daily check-in result against an optional
// cached historical average in TXB minor units.
func FormatCheckIn(copy Copy, result activity.DailyCheckIn, averageMinor *int64) string {
	tone := copy.SignInNeutral
	if result.AlreadyClaimed {
		tone = copy.SignInAlready
	} else if averageMinor != nil {
		switch {
		case result.RewardMinor > *averageMinor:
			tone = copy.SignInAbove
		case result.RewardMinor < *averageMinor:
			tone = copy.SignInBelow
		default:
			tone = copy.SignInEqual
		}
	}
	lines := []string{
		escapeMarkdownV2(tone),
		"🎁 *" + escapeMarkdownV2(copy.SignInReward) + ":* " + escapeMarkdownV2("+"+model.TXBMoney(result.RewardMinor).Display),
		"💰 *" + escapeMarkdownV2(copy.BalanceLabel) + ":* " + escapeMarkdownV2(model.TXBMoney(result.BalanceAfterMinor).Display),
	}
	return Limit(strings.Join(lines, "\n"))
}

// FormatStart renders the bot entry-point sentence without a title.
func FormatStart(copy Copy) string {
	return Limit(escapeMarkdownV2(copy.Start))
}

// FormatWelcome renders a safe MarkdownV2 text mention for a joining member.
func FormatWelcome(copy Copy, telegramID int64, displayName string) string {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = copy.MemberLabel
	}
	return Limit(escapeMarkdownV2(copy.WelcomePrefix) + safeMention(telegramID, displayName) + escapeMarkdownV2(copy.WelcomeSuffix))
}
