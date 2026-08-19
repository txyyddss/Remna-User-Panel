package remnawave

import (
	"net/http"
	"strings"
)

// Remnawave nodes allow a 32 MiB Geocheck command result. Leave room for the
// backend response envelope while retaining the stricter default everywhere else.
const maxGeocheckResponseBytes = 40 << 20

func responseByteLimit(method, endpoint string) int {
	if method == http.MethodGet && strings.HasPrefix(endpoint, "/api/connections/geocheck/") {
		return maxGeocheckResponseBytes
	}
	return maxResponseBytes
}
