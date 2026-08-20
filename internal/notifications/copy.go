package notifications

type copySet struct {
	titles map[string]string
	labels map[string]string
	values map[string]string
}

var englishCopy = copySet{
	titles: map[string]string{
		"expiration": "🛑 Subscription expired", "expiry_reminder": "⏳ Expires in 2 days",
		"queued_activation": "🚀 Queued combo activated", "auto_renewal": "♻️ Auto-renewed",
		"traffic_threshold": "⚠️ Traffic above 90%", "group_reward": "🎁 Group reward received",
		"admin_extension": "🎁 Extended by admin", "admin_update": "🛠 Updated by admin",
	},
	labels: map[string]string{
		"combo": "Combo", "expires": "Expires", "expired": "Expired", "autoRenewal": "Auto-renewal",
		"queuedCombo": "Queued combo", "traffic": "Traffic", "reset": "Reset", "addOns": "Add-ons",
		"validUntil": "Valid until", "renewalDebit": "Renewal debit", "usedAllocated": "Used / Allocated",
		"eligible": "Eligible unused", "rollover": "Rollover", "balance": "Balance", "remaining": "Remaining",
		"messages": "Messages", "reward": "Reward", "time": "Time", "change": "Change", "amount": "Amount",
		"credited": "Credited", "cancelledCombos": "Cancelled combos", "added": "Added",
		"previousExpiry": "Previous expiry", "newExpiry": "New expiry", "previousCombo": "Previous combo",
		"newCombo": "New combo", "validFrom": "Valid from", "status": "Status", "squads": "Squads", "reason": "Reason",
	},
	values: map[string]string{
		"off": "Off", "none": "None", "unavailable": "Unavailable", "DAY": "Daily", "WEEK": "Weekly",
		"MONTH": "Monthly", "MONTH_ROLLING": "Monthly rolling", "NO_RESET": "No reset",
		"balance_adjustment": "TXB adjustment", "balance_deduction": "TXB deduction", "payment_refund": "Payment refund",
		"courtesy_credit": "Courtesy credit", "entitlement_refund": "Entitlement refund",
		"entitlement_cancel": "Entitlement cancellation", "combo_replacement": "Combo replacement", "entitlement_edit": "Entitlement edit",
		"activating": "Activating", "active": "Active", "queued": "Queued", "expired": "Expired", "cancelled": "Cancelled", "failed": "Failed",
	},
}

var chineseCopy = copySet{
	titles: map[string]string{
		"expiration": "🛑 订阅已到期", "expiry_reminder": "⏳ 订阅将在 2 天后到期",
		"queued_activation": "🚀 排队套餐已启用", "auto_renewal": "♻️ 自动续费成功",
		"traffic_threshold": "⚠️ 流量已超过 90%", "group_reward": "🎁 群聊奖励到账",
		"admin_extension": "🎁 管理员已延长订阅", "admin_update": "🛠 管理员已更新账户",
	},
	labels: map[string]string{
		"combo": "套餐", "expires": "到期时间", "expired": "到期时间", "autoRenewal": "自动续费",
		"queuedCombo": "排队套餐", "traffic": "流量", "reset": "重置", "addOns": "附加项",
		"validUntil": "有效期至", "renewalDebit": "续费扣款", "usedAllocated": "已用 / 总量",
		"eligible": "可结转未用量", "rollover": "结转返还", "balance": "余额", "remaining": "剩余",
		"messages": "消息数", "reward": "奖励", "time": "时间", "change": "变更", "amount": "金额",
		"credited": "返还", "cancelledCombos": "已取消套餐", "added": "延长",
		"previousExpiry": "原到期时间", "newExpiry": "新到期时间", "previousCombo": "原套餐",
		"newCombo": "新套餐", "validFrom": "生效时间", "status": "状态", "squads": "节点组", "reason": "原因",
	},
	values: map[string]string{
		"off": "关闭", "none": "无", "unavailable": "不可用", "DAY": "每日", "WEEK": "每周",
		"MONTH": "每月", "MONTH_ROLLING": "滚动月", "NO_RESET": "不重置",
		"balance_adjustment": "TXB 调整", "balance_deduction": "TXB 扣除", "payment_refund": "支付退款",
		"courtesy_credit": "补偿金", "entitlement_refund": "套餐退款",
		"entitlement_cancel": "取消套餐", "combo_replacement": "更换套餐", "entitlement_edit": "编辑订阅",
		"activating": "激活中", "active": "使用中", "queued": "排队中", "expired": "已到期", "cancelled": "已取消", "failed": "失败",
	},
}

func copyFor(locale string) copySet {
	if locale == "zh-CN" {
		return chineseCopy
	}
	return englishCopy
}
