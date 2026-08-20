package httpapi

import "strconv"

func parsePositiveTelegramID(value string) (int64, error) {
	if value == "" {
		return 0, strconv.ErrSyntax
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return 0, strconv.ErrSyntax
		}
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, strconv.ErrSyntax
	}
	return parsed, nil
}
