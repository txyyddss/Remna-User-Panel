package abuse

import "strings"

// ValidOutboundTag reports whether value is a safe Xray outbound tag.
func ValidOutboundTag(value string) bool {
	if len(value) == 0 || len(value) > MaxOutboundTagLength || (value[0] < 'a' || value[0] > 'z') && (value[0] < 'A' || value[0] > 'Z') && (value[0] < '0' || value[0] > '9') {
		return false
	}
	for _, character := range value {
		if (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') && (character < '0' || character > '9') && character != '.' && character != '_' && character != '-' {
			return false
		}
	}
	return true
}

// NormalizeOutboundTags validates, trims, and deduplicates configured Xray outbound tags.
func NormalizeOutboundTags(values []string) ([]string, bool) {
	if len(values) == 0 || len(values) > MaxOutboundTags {
		return nil, false
	}
	tags := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if !ValidOutboundTag(value) {
			return nil, false
		}
		key := strings.ToLower(value)
		if _, duplicate := seen[key]; duplicate {
			return nil, false
		}
		seen[key] = struct{}{}
		tags = append(tags, value)
	}
	return tags, true
}
