package handlers

import "strings"

func extractShortUUID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.ContainsAny(raw, "/?#") {
		return raw
	}

	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '/' || r == '?' || r == '#'
	})
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
