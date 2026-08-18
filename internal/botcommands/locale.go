package botcommands

import "strings"

// Language is a supported bot reply locale.
type Language string

const (
	English Language = "en"
	Chinese Language = "zh-CN"
)

// Description is one localized Bot API command description.
type Description struct {
	Name        Name
	Description string
}

// Copy owns every user-visible bot string.
type Copy struct {
	Descriptions        []Description
	Start               string
	Unknown             string
	Unavailable         string
	NoSubscription      string
	BalanceLabel        string
	UsageLabel          string
	RemainingLabel      string
	DaysLabel           string
	NodesLabel          string
	ComboLabel          string
	SquadsLabel         string
	TrafficLabel        string
	ResetLabel          string
	RolloverLabel       string
	RolloverWill        string
	RolloverWillNot     string
	RolloverCannot      string
	RolloverUnavailable string
	SignInReward        string
	SignInAlready       string
}

// LanguageFor maps all Telegram Chinese variants to Simplified Chinese.
func LanguageFor(code string) Language {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(code)), "zh") {
		return Chinese
	}
	return English
}

// Text returns the immutable copy set for a language.
func Text(language Language) Copy {
	if language == Chinese {
		return chineseCopy()
	}
	return englishCopy()
}

func englishCopy() Copy {
	return Copy{
		Descriptions: []Description{{Sub, "Show subscription usage"}, {Balance, "Show your TXB balance"}, {SignIn, "Claim today's check-in"}, {Start, "Open TX Carpool"}, {MyCombo, "Show combo and rollover details"}},
		Start:        "Press Open TX Carpool to use the app.", Unknown: "This command is not available.", Unavailable: "The requested information is temporarily unavailable.",
		NoSubscription: "No active subscription.", BalanceLabel: "Balance", UsageLabel: "Usage", RemainingLabel: "Remaining", DaysLabel: "days", NodesLabel: "Nodes",
		ComboLabel: "Combo", SquadsLabel: "Squads", TrafficLabel: "Traffic", ResetLabel: "Reset", RolloverLabel: "Rollover",
		RolloverWill: "will receive", RolloverWillNot: "will not receive", RolloverCannot: "cannot receive", RolloverUnavailable: "unavailable",
		SignInReward: "Check-in reward", SignInAlready: "Today's check-in was already recorded",
	}
}

func chineseCopy() Copy {
	return Copy{
		Descriptions: []Description{{Sub, "查看订阅用量"}, {Balance, "查看 TXB 余额"}, {SignIn, "领取今日签到奖励"}, {Start, "打开 TX Carpool"}, {MyCombo, "查看套餐与结转状态"}},
		Start:        "请按“打开 TX Carpool”使用应用。", Unknown: "此命令不可用。", Unavailable: "暂时无法获取所需信息。",
		NoSubscription: "当前没有生效中的订阅。", BalanceLabel: "余额", UsageLabel: "用量", RemainingLabel: "剩余", DaysLabel: "天", NodesLabel: "节点",
		ComboLabel: "套餐", SquadsLabel: "线路组", TrafficLabel: "流量", ResetLabel: "重置", RolloverLabel: "结转",
		RolloverWill: "预计可获得", RolloverWillNot: "预计无法获得", RolloverCannot: "不可获得", RolloverUnavailable: "暂不可用",
		SignInReward: "签到奖励", SignInAlready: "今日签到奖励已领取",
	}
}
