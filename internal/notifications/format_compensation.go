package notifications

import (
	"errors"
	"time"
)

func compensationFields(copy copySet, facts map[string]string, location *time.Location, chinese bool) ([]cardField, error) {
	fields, err := requiredFields(copy, facts,
		pair("node", FactNode),
		pair("affectedSquads", FactAffectedSquads),
		durationPair("downtime", FactDowntimeSeconds, chinese),
		datePair("outageStarted", FactOutageStarted, location),
		datePair("recovered", FactRecovered, location),
		durationPair("compensation", FactAddedSeconds, chinese),
	)
	if err != nil {
		return nil, err
	}
	if capped := facts[FactCompensationCapped]; capped != "" {
		if capped != "true" {
			return nil, errors.New("invalid compensation capped fact")
		}
		fields = append(fields, cardField{copy.labels["capApplied"], copy.values["yes"]})
	}
	details, err := requiredFields(copy, facts,
		pair("combo", FactCombo),
		datePair("previousExpiry", FactPreviousExpiry, location),
		datePair("newExpiry", FactNewExpiry, location),
		pair("reason", FactReason),
		datePair("appliedAt", FactTime, location),
	)
	if err != nil {
		return nil, err
	}
	return append(fields, details...), nil
}

func durationPair(label, key string, chinese bool) fieldSpec {
	return fieldSpec{label: label, key: key, format: func(value string) (string, error) {
		return durationValue(value, chinese)
	}}
}
