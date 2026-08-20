package notifications

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	jobpayload "github.com/txyyddss/Remna-User-Panel/internal/outbox"
	"github.com/txyyddss/Remna-User-Panel/internal/telegramformat"
)

type cardField struct{ label, value string }

// Format renders one validated notification as bounded MarkdownV2.
func Format(payload jobpayload.UserNotification, location *time.Location) (string, error) {
	copy := copyFor(payload.Locale)
	title := copy.titles[payload.Kind]
	if title == "" {
		return "", errors.New("notification title is unavailable")
	}
	fields, err := notificationFields(payload, copy, location)
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(title, " ", 2)
	lines := []string{parts[0] + " *" + telegramformat.Escape(parts[1]) + "*"}
	for _, field := range fields {
		lines = append(lines, "*"+telegramformat.Escape(field.label)+":* "+telegramformat.Escape(field.value))
	}
	return telegramformat.Limit(strings.Join(lines, "\n")), nil
}

func notificationFields(payload jobpayload.UserNotification, copy copySet, location *time.Location) ([]cardField, error) {
	facts := payload.Facts
	switch payload.Kind {
	case jobpayload.UserEventExpiration:
		return requiredFields(copy, facts, pair("combo", FactCombo), datePair("expired", FactExpired, location))
	case jobpayload.UserEventExpiryReminder:
		return requiredFields(copy, facts, pair("combo", FactCombo), datePair("expires", FactExpires, location),
			fixedPair("autoRenewal", "off", copy), fixedPair("queuedCombo", "none", copy))
	case jobpayload.UserEventQueuedActivation:
		return activationFields(copy, facts, location)
	case jobpayload.UserEventAutoRenewal:
		return renewalFields(copy, facts, location)
	case jobpayload.UserEventTrafficThreshold:
		return trafficFields(copy, facts)
	case jobpayload.UserEventGroupReward:
		return requiredFields(copy, facts, pair("messages", FactMessages), moneyPair("reward", FactReward),
			moneyPair("balance", FactBalance), datePair("time", FactTime, location))
	case jobpayload.UserEventAdminExtension, jobpayload.UserEventAdminUpdate:
		return adminFields(copy, facts, location, payload.Kind == jobpayload.UserEventAdminExtension, payload.Locale == "zh-CN")
	default:
		return nil, errors.New("unsupported notification format")
	}
}

func activationFields(copy copySet, facts map[string]string, location *time.Location) ([]cardField, error) {
	fields, err := requiredFields(copy, facts, pair("combo", FactCombo), bytesPair("traffic", FactTrafficLimit),
		translatedPair("reset", FactReset, copy), datePair("validUntil", FactValidUntil, location))
	if err != nil {
		return nil, err
	}
	return insertOptional(fields, len(fields)-1, copy, facts, "addOns", FactAddOns), nil
}

func renewalFields(copy copySet, facts map[string]string, location *time.Location) ([]cardField, error) {
	used, err := byteRatio(facts, FactUsed, FactAllocated)
	if err != nil {
		return nil, err
	}
	rollover := copy.values["unavailable"]
	if facts[FactRolloverStatus] != "unavailable" {
		rollover, err = money(facts[FactRollover])
		if err != nil {
			return nil, err
		}
	}
	fields, err := requiredFields(copy, facts, pair("combo", FactCombo), moneyPair("renewalDebit", FactRenewalDebit),
		literalPair("usedAllocated", used), bytesPair("eligible", FactEligible), literalPair("rollover", rollover),
		moneyPair("balance", FactBalance), datePair("validUntil", FactValidUntil, location))
	return fields, err
}

func trafficFields(copy copySet, facts map[string]string) ([]cardField, error) {
	usage, err := byteRatio(facts, FactUsed, FactTrafficLimit)
	if err != nil {
		return nil, err
	}
	used, _ := strconv.ParseInt(facts[FactUsed], 10, 64)
	limit, _ := strconv.ParseInt(facts[FactTrafficLimit], 10, 64)
	usage += fmt.Sprintf(" (%.1f%%)", float64(used)*100/float64(limit))
	return requiredFields(copy, facts, pair("combo", FactCombo), literalPair("traffic", usage),
		bytesPair("remaining", FactRemaining), translatedPair("reset", FactReset, copy))
}
