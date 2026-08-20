package notifications

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func adminFields(copy copySet, facts map[string]string, location *time.Location, extension, chinese bool) ([]cardField, error) {
	fields := make([]cardField, 0, 12)
	if extension {
		added, err := durationValue(facts[FactAddedSeconds], chinese)
		if err != nil {
			return nil, err
		}
		fields = append(fields, cardField{copy.labels["added"], added})
		fields, err = appendRequiredDates(fields, copy, facts, location)
		if err != nil {
			return nil, err
		}
	} else {
		change := copy.values[facts[FactChange]]
		if change == "" {
			return nil, errors.New("admin notification change is missing")
		}
		fields = append(fields, cardField{copy.labels["change"], change})
	}
	var err error
	if fields, err = appendAdminAmounts(fields, copy, facts); err != nil {
		return nil, err
	}
	if fields, err = appendAdminChanges(fields, copy, facts, location, extension); err != nil {
		return nil, err
	}
	if reason := strings.TrimSpace(facts[FactReason]); reason != "" {
		fields = append(fields, cardField{copy.labels["reason"], reason})
	} else {
		return nil, errors.New("admin notification reason is missing")
	}
	when, err := dateValue(facts[FactTime], location)
	if err != nil {
		return nil, err
	}
	return append(fields, cardField{copy.labels["time"], when}), nil
}

func appendRequiredDates(fields []cardField, copy copySet, facts map[string]string, location *time.Location) ([]cardField, error) {
	previous, err := dateValue(facts[FactPreviousExpiry], location)
	if err != nil {
		return nil, err
	}
	next, err := dateValue(facts[FactNewExpiry], location)
	if err != nil {
		return nil, err
	}
	return append(fields, cardField{copy.labels["previousExpiry"], previous}, cardField{copy.labels["newExpiry"], next}), nil
}

func appendAdminAmounts(fields []cardField, copy copySet, facts map[string]string) ([]cardField, error) {
	for _, item := range []struct{ key, label string }{
		{FactAmount, "amount"}, {FactCredited, "credited"}, {FactBalance, "balance"},
	} {
		if facts[item.key] == "" {
			continue
		}
		value, err := money(facts[item.key])
		if err != nil {
			return nil, err
		}
		fields = append(fields, cardField{copy.labels[item.label], value})
	}
	for _, item := range []struct{ key, label string }{{FactCombo, "combo"}, {FactCancelledCombos, "cancelledCombos"}, {FactAddOns, "addOns"}} {
		if value := strings.TrimSpace(facts[item.key]); value != "" {
			fields = append(fields, cardField{copy.labels[item.label], value})
		}
	}
	return fields, nil
}

func appendAdminChanges(fields []cardField, copy copySet, facts map[string]string, location *time.Location, extension bool) ([]cardField, error) {
	items := []struct{ before, after, label, kind string }{
		{FactPreviousCombo, FactNewCombo, "combo", "plain"},
		{FactPreviousExpiry, FactNewExpiry, "validUntil", "date"},
		{FactPreviousValidFrom, FactNewValidFrom, "validFrom", "date"},
		{FactPreviousStatus, FactNewStatus, "status", "translated"},
		{FactPreviousTraffic, FactNewTraffic, "traffic", "bytes"},
		{FactPreviousReset, FactNewReset, "reset", "translated"},
		{FactPreviousSquads, FactNewSquads, "squads", "plain"},
	}
	for _, item := range items {
		if extension && item.before == FactPreviousExpiry {
			continue
		}
		before, after := strings.TrimSpace(facts[item.before]), strings.TrimSpace(facts[item.after])
		if before == "" || after == "" || before == after {
			continue
		}
		var err error
		before, after, err = changeValues(before, after, item.kind, copy, location)
		if err != nil {
			return nil, err
		}
		fields = append(fields, cardField{copy.labels[item.label], before + " → " + after})
	}
	return fields, nil
}

func changeValues(before, after, kind string, copy copySet, location *time.Location) (string, string, error) {
	var err error
	switch kind {
	case "date":
		before, err = dateValue(before, location)
		if err == nil {
			after, err = dateValue(after, location)
		}
	case "bytes":
		before, err = bytesValue(before)
		if err == nil {
			after, err = bytesValue(after)
		}
	case "translated":
		if value := copy.values[before]; value != "" {
			before = value
		}
		if value := copy.values[after]; value != "" {
			after = value
		}
	}
	return before, after, err
}

func durationValue(raw string, chinese bool) (string, error) {
	seconds, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || seconds <= 0 {
		return "", errors.New("invalid extension duration")
	}
	days, hours, minutes := seconds/86400, seconds%86400/3600, seconds%3600/60
	parts := make([]string, 0, 3)
	units := []string{" days", " hours", " minutes"}
	if chinese {
		units = []string{" 天", " 小时", " 分钟"}
	}
	for index, value := range []int64{days, hours, minutes} {
		if value > 0 {
			unit := units[index]
			if !chinese && value == 1 {
				unit = strings.TrimSuffix(unit, "s")
			}
			parts = append(parts, strconv.FormatInt(value, 10)+unit)
		}
	}
	if len(parts) == 0 {
		parts = append(parts, strconv.FormatInt(seconds, 10)+map[bool]string{true: " 秒", false: " seconds"}[chinese])
	}
	return strings.Join(parts, " "), nil
}
