package questionnaires

import (
	"strings"
)

func delimiterLabel(delimiter rune) string {
	switch delimiter {
	case ';':
		return "semicolon"
	case '\t':
		return "tab"
	default:
		return "comma"
	}
}

func trimRecord(record []string) []string {
	result := make([]string, len(record))
	for index, value := range record {
		result[index] = strings.TrimSpace(value)
	}
	return result
}

func hasEmptyHeader(headers []string) bool {
	for _, header := range headers {
		if header == "" {
			return true
		}
	}
	return false
}
