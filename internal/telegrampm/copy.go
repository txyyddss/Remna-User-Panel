package telegrampm

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/txyyddss/Remna-User-Panel/internal/integrations/telegram"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/telegramformat"
)

func chinese(locale string) bool { return strings.HasPrefix(strings.ToLower(locale), "zh") }

func NoticeText(item model.PMNotice) string {
	if chinese(item.Locale) {
		switch item.Reason {
		case "unsupported":
			return "Telegram 无法转发这类消息，请改用文字或常规媒体。"
		case "entry":
			return "请先打开 TX Carpool 面板并完成首次访问验证，再直接发送消息。"
		case "blocked":
			return "你的私信已被管理员暂停。"
		default:
			return "私信功能暂未开放，请稍后再试。"
		}
	}
	switch item.Reason {
	case "unsupported":
		return "Telegram cannot relay this message type. Please send text or ordinary media instead."
	case "entry":
		return "Open the TX Carpool panel and complete entry verification before sending a message directly here."
	case "blocked":
		return "Private messaging for your account has been blocked by an administrator."
	default:
		return "Private messaging is currently unavailable. Please try again later."
	}
}

func callbackText(locale, reason string) string {
	if chinese(locale) {
		if reason == "updated" {
			return "私信设置已更新。"
		}
		return "仅面板管理员可使用当前话题的资料卡。"
	}
	if reason == "updated" {
		return "Private messaging settings updated."
	}
	return "Only panel admins can use the current topic profile card."
}

func TopicName(item model.PMConversation) string {
	name := strings.TrimSpace(item.FirstName + " " + item.LastName)
	if item.Username != "" {
		name = "@" + item.Username
	}
	name += " · " + strconv.FormatInt(item.TelegramID, 10)
	if utf8.RuneCountInString(name) > 128 {
		name = string([]rune(name)[:128])
	}
	return name
}

func Profile(item model.PMConversation) telegram.TopicProfileRequest {
	name := strings.TrimSpace(item.FirstName + " " + item.LastName)
	if item.Username != "" {
		name += " (@" + item.Username + ")"
	}
	blockAction, muteAction, blockLabel, muteLabel := "block", "mute", "Block", "Mute"
	state := "Messages enabled"
	if item.Blocked {
		blockAction, blockLabel, state = "unblock", "Unblock", "Messages blocked"
	}
	if item.Muted {
		muteAction, muteLabel = "unmute", "Unmute"
		state += " · Silent relay"
	}
	if chinese(item.Locale) {
		blockLabel, muteLabel, state = "屏蔽", "静音", "允许私信"
		if item.Blocked {
			blockLabel, state = "取消屏蔽", "已屏蔽私信"
		}
		if item.Muted {
			muteLabel = "取消静音"
			state += " · 静默转发"
		}
	}
	body := telegramformat.Escape(name) + "\n" + telegramformat.Escape("Telegram ID: "+strconv.FormatInt(item.TelegramID, 10)) + "\n" + telegramformat.Escape(state)
	return telegram.TopicProfileRequest{ChatID: item.ChatID, MessageThreadID: item.TopicID, MessageID: item.ProfileMessageID, Text: body,
		ReplyMarkup: telegram.InlineKeyboardMarkup{InlineKeyboard: [][]telegram.InlineKeyboardButton{{{Text: blockLabel, CallbackData: "pm:" + blockAction + ":" + item.ID}, {Text: muteLabel, CallbackData: "pm:" + muteAction + ":" + item.ID}}}}}
}
