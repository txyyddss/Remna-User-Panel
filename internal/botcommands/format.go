package botcommands

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/txyyddss/Remna-User-Panel/internal/activity"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

const MessageLimit = 4096

// FormatBalance renders a self-only TXB balance reply.
func FormatBalance(copy Copy, money model.Money) string {
	return Limit(fmt.Sprintf("%s: %s", copy.BalanceLabel, money.Display))
}

// FormatCheckIn renders an idempotent daily check-in result.
func FormatCheckIn(copy Copy, result activity.DailyCheckIn) string {
	if result.AlreadyClaimed {
		return Limit(fmt.Sprintf("%s\n%s: %s", copy.SignInAlready, copy.BalanceLabel, model.TXBMoney(result.BalanceAfterMinor).Display))
	}
	return Limit(fmt.Sprintf("%s: %s\n%s: %s", copy.SignInReward, model.TXBMoney(result.RewardMinor).Display,
		copy.BalanceLabel, model.TXBMoney(result.BalanceAfterMinor).Display))
}

// FormatSubscription renders current usage and the provider's top-node totals.
func FormatSubscription(copy Copy, dashboard model.Dashboard, now time.Time) string {
	if dashboard.ActivePurchase == nil || dashboard.Statistics == nil {
		return copy.NoSubscription
	}
	used := parseBytes(dashboard.Statistics.UsedTrafficBytes)
	limit := parseBytes(dashboard.Statistics.TrafficLimitBytes)
	percent := percentage(used, limit)
	days := int(math.Ceil(dashboard.ActivePurchase.ValidUntil.Sub(now).Hours() / 24))
	if days < 0 {
		days = 0
	}
	lines := []string{
		fmt.Sprintf("%s %s %.1f%%", progressBar(percent), copy.UsageLabel, percent),
		fmt.Sprintf("%s / %s", formatBytes(used), formatBytes(limit)),
		fmt.Sprintf("%s: %d %s", copy.RemainingLabel, days, copy.DaysLabel),
	}
	if len(dashboard.Statistics.TopNodes) > 0 {
		lines = append(lines, "", copy.NodesLabel+":")
		for _, node := range dashboard.Statistics.TopNodes {
			value := parseBytes(node.TotalBytes)
			share := percentage(value, used)
			lines = append(lines, fmt.Sprintf("- %s (%s): %s (%.1f%%)", node.Name, strings.ToUpper(node.CountryCode), formatBytes(value), share))
		}
	}
	return Limit(strings.Join(lines, "\n"))
}

// FormatCombo renders the active purchase with localized rollover state.
func FormatCombo(copy Copy, purchase *model.Purchase, squadNames []string, rollover string) string {
	if purchase == nil {
		return copy.NoSubscription
	}
	if len(squadNames) == 0 {
		squadNames = append([]string(nil), purchase.SquadUUIDs...)
	}
	lines := []string{
		fmt.Sprintf("%s: %s", copy.ComboLabel, purchase.ComboName),
		fmt.Sprintf("%s: %s", copy.SquadsLabel, strings.Join(squadNames, ", ")),
		fmt.Sprintf("%s: %s", copy.TrafficLabel, formatBytes(purchase.TrafficLimitBytes)),
		fmt.Sprintf("%s: %s", copy.ResetLabel, purchase.ResetStrategy),
		fmt.Sprintf("%s: %s", copy.RolloverLabel, rollover),
	}
	return Limit(strings.Join(lines, "\n"))
}

// Limit keeps replies below Telegram's 4096-code-point limit.
func Limit(value string) string {
	if utf8.RuneCountInString(value) <= MessageLimit {
		return value
	}
	runes := []rune(value)
	return strings.TrimSpace(string(runes[:MessageLimit-3])) + "..."
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
