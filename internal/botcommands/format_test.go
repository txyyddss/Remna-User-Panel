package botcommands

import (
	"strings"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/activity"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

func TestFormattedRepliesUseMarkdownV2StatusCards(t *testing.T) {
	t.Parallel()
	for _, language := range []Language{English, Chinese} {
		language := language
		t.Run(string(language), func(t *testing.T) {
			t.Parallel()
			copy := Text(language)
			replies := []string{
				FormatStart(copy),
				FormatUnknown(copy),
				FormatUnavailable(copy),
				FormatNoSubscription(copy),
				FormatBalance(copy, model.TXBMoney(1_250)),
				FormatCheckIn(copy, activity.DailyCheckIn{RewardMinor: 50, BalanceAfterMinor: 1_300}),
				FormatCheckIn(copy, activity.DailyCheckIn{AlreadyClaimed: true, BalanceAfterMinor: 1_300}),
				FormatSubscription(copy, markdownDashboard(), time.Now().UTC()),
				FormatCombo(copy, markdownPurchase(), []string{markdownValue}, copy.RolloverWill),
				FormatDeductUsage(copy),
				FormatDeductRejected(copy),
				FormatDeductSucceeded(copy, model.TXBMoney(500)),
			}
			for _, reply := range replies {
				if !strings.HasPrefix(reply, "*") || !strings.Contains(reply, "\n*") {
					t.Fatalf("reply is not a status card: %q", reply)
				}
				if len([]rune(reply)) > MessageLimit {
					t.Fatalf("reply exceeds message limit: %d", len([]rune(reply)))
				}
			}
		})
	}
}

func TestFormattersEscapeMarkdownV2DynamicValues(t *testing.T) {
	t.Parallel()
	copy := Text(English)
	replies := []string{
		FormatCombo(copy, markdownPurchase(), []string{markdownValue}, copy.RolloverWill),
		FormatSubscription(copy, markdownDashboard(), time.Now().UTC()),
	}
	for _, reply := range replies {
		for _, escaped := range []string{`\_`, `\*`, `\[`, `\]`, `\(`, `\)`, `\~`, "\\`", `\>`, `\#`, `\+`, `\-`, `\=`, `\|`, `\{`, `\}`, `\.`, `\!`, `\\`} {
			if !strings.Contains(reply, escaped) {
				t.Fatalf("reply does not escape %q: %q", escaped, reply)
			}
		}
	}
}

func TestLimitMarkdownV2KeepsEscapesComplete(t *testing.T) {
	t.Parallel()
	result := Limit(strings.Repeat("a", MessageLimit-7) + "\\" + strings.Repeat("x", 10))
	withoutEllipsis := strings.TrimSuffix(result, markdownV2Ellipsis)
	if len([]rune(result)) > MessageLimit || !strings.HasSuffix(result, markdownV2Ellipsis) {
		t.Fatalf("limited reply = %q", result)
	}
	if trailingBackslashes([]rune(withoutEllipsis))%2 != 0 {
		t.Fatalf("limited reply ends with an incomplete escape: %q", result)
	}
}

const markdownValue = "edge_*[]()~`>#+-=|{}.!\\"

func markdownDashboard() model.Dashboard {
	return model.Dashboard{
		ActivePurchase: markdownPurchase(),
		Statistics: &model.Statistics{
			UsedTrafficBytes:  "1024",
			TrafficLimitBytes: "2048",
			TopNodes: []model.TopNode{{
				Name: markdownValue, CountryCode: "us", TotalBytes: "1024",
			}},
		},
	}
}

func markdownPurchase() *model.Purchase {
	return &model.Purchase{ComboName: markdownValue, TrafficLimitBytes: 1024, ResetStrategy: "MONTHLY"}
}
