package botcommands

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/activity"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

const MessageLimit = 4096

type cardField struct {
	label string
	value string
}

// FormatBalance renders a self-only TXB balance reply.
func FormatBalance(copy Copy, money model.Money) string {
	return formatCard(copy.BalanceTitle, cardField{copy.BalanceLabel, money.Display})
}

// FormatCheckIn renders an idempotent daily check-in result.
func FormatCheckIn(copy Copy, result activity.DailyCheckIn) string {
	if result.AlreadyClaimed {
		return formatCard(copy.CheckInTitle,
			cardField{copy.StatusLabel, copy.SignInAlready},
			cardField{copy.BalanceLabel, model.TXBMoney(result.BalanceAfterMinor).Display},
		)
	}
	return formatCard(copy.CheckInTitle,
		cardField{copy.SignInReward, model.TXBMoney(result.RewardMinor).Display},
		cardField{copy.BalanceLabel, model.TXBMoney(result.BalanceAfterMinor).Display},
	)
}

// FormatSubscription renders current usage and the provider's top-node totals.
func FormatSubscription(copy Copy, dashboard model.Dashboard, now time.Time) string {
	if dashboard.ActivePurchase == nil || dashboard.Statistics == nil {
		return FormatNoSubscription(copy)
	}
	used := parseBytes(dashboard.Statistics.UsedTrafficBytes)
	limit := parseBytes(dashboard.Statistics.TrafficLimitBytes)
	percent := percentage(used, limit)
	days := int(math.Ceil(dashboard.ActivePurchase.ValidUntil.Sub(now).Hours() / 24))
	if days < 0 {
		days = 0
	}
	fields := []cardField{
		{copy.UsageLabel, fmt.Sprintf("%s %.1f%%", progressBar(percent), percent)},
		{copy.TrafficLabel, fmt.Sprintf("%s / %s", formatBytes(used), formatBytes(limit))},
		{copy.RemainingLabel, fmt.Sprintf("%d %s", days, copy.DaysLabel)},
	}
	var nodes []string
	if len(dashboard.Statistics.TopNodes) > 0 {
		for _, node := range dashboard.Statistics.TopNodes {
			value := parseBytes(node.TotalBytes)
			share := percentage(value, used)
			nodes = append(nodes, fmt.Sprintf("%s (%s): %s (%.1f%%)", node.Name, strings.ToUpper(node.CountryCode), formatBytes(value), share))
		}
	}
	return formatCardWithItems(copy.SubscriptionTitle, fields, copy.NodesLabel, nodes)
}

// FormatCombo renders the active purchase with localized rollover state.
func FormatCombo(copy Copy, purchase *model.Purchase, squadNames []string, rollover string) string {
	if purchase == nil {
		return FormatNoSubscription(copy)
	}
	if len(squadNames) == 0 {
		squadNames = append([]string(nil), purchase.SquadUUIDs...)
	}
	return formatCard(copy.ComboTitle,
		cardField{copy.ComboLabel, purchase.ComboName},
		cardField{copy.SquadsLabel, strings.Join(squadNames, ", ")},
		cardField{copy.TrafficLabel, formatBytes(purchase.TrafficLimitBytes)},
		cardField{copy.ResetLabel, purchase.ResetStrategy},
		cardField{copy.RolloverLabel, rollover},
	)
}

// FormatStart renders the bot entry-point reply.
func FormatStart(copy Copy) string {
	return formatCard(copy.StartTitle, cardField{copy.StatusLabel, copy.Start})
}

// FormatUnknown renders a safe unknown-command response.
func FormatUnknown(copy Copy) string {
	return formatCard(copy.UnknownTitle, cardField{copy.StatusLabel, copy.Unknown})
}

// FormatUnavailable renders a temporary command-data failure.
func FormatUnavailable(copy Copy) string {
	return formatCard(copy.UnavailableTitle, cardField{copy.StatusLabel, copy.Unavailable})
}

// FormatNoSubscription renders an empty subscription state.
func FormatNoSubscription(copy Copy) string {
	return formatCard(copy.NoSubscriptionTitle, cardField{copy.StatusLabel, copy.NoSubscription})
}

// FormatDeductUsage renders safe administrator command usage guidance.
func FormatDeductUsage(copy Copy) string {
	return formatCard(copy.DeductTitle, cardField{copy.StatusLabel, copy.DeductUsage})
}

// FormatDeductRejected renders a deduction rejection without exposing details.
func FormatDeductRejected(copy Copy) string {
	return formatCard(copy.DeductTitle, cardField{copy.StatusLabel, copy.DeductRejected})
}

// FormatDeductSucceeded renders a successful balance deduction.
func FormatDeductSucceeded(copy Copy, amount model.Money) string {
	return formatCard(copy.DeductTitle,
		cardField{copy.StatusLabel, copy.DeductSucceeded},
		cardField{copy.AmountLabel, amount.Display},
	)
}

func formatCard(title string, fields ...cardField) string {
	return formatCardWithItems(title, fields, "", nil)
}

func formatCardWithItems(title string, fields []cardField, itemLabel string, items []string) string {
	lines := []string{"*" + escapeMarkdownV2(title) + "*"}
	for _, field := range fields {
		lines = append(lines, "*"+escapeMarkdownV2(field.label)+":* "+escapeMarkdownV2(field.value))
	}
	if len(items) > 0 {
		lines = append(lines, "", "*"+escapeMarkdownV2(itemLabel)+":*")
		for _, item := range items {
			lines = append(lines, escapeMarkdownV2(item))
		}
	}
	return Limit(strings.Join(lines, "\n"))
}

// Limit keeps MarkdownV2 replies below Telegram's 4096-code-point limit.
func Limit(value string) string {
	return limitMarkdownV2(value)
}

func parseBytes(value string) int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || parsed < 0 {
		return 0
	}
	return parsed
}

func formatBytes(value int64) string {
	const unit = int64(1024)
	if value < unit {
		return fmt.Sprintf("%d B", value)
	}
	divisor, suffix := unit, "KiB"
	for _, candidate := range []string{"MiB", "GiB", "TiB", "PiB"} {
		if value < divisor*unit {
			break
		}
		divisor *= unit
		suffix = candidate
	}
	return fmt.Sprintf("%.2f %s", float64(value)/float64(divisor), suffix)
}

func percentage(value, total int64) float64 {
	if value <= 0 || total <= 0 {
		return 0
	}
	return math.Min(100, float64(value)*100/float64(total))
}

func progressBar(percent float64) string {
	filled := int(math.Round(math.Min(100, math.Max(0, percent)) / 10))
	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", 10-filled) + "]"
}
