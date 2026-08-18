package httpapi

import (
	"net/http"
	"net/url"
	"strings"
)

const streamingBackupUploadPath = "/api/v1/admin/backups/upload"

func isStreamingBackupUpload(r *http.Request) bool {
	return r.Method == http.MethodPost && r.URL.Path == streamingBackupUploadPath
}

func (s *Server) authorizeStreamingBackupUpload(w http.ResponseWriter, r *http.Request) bool {
	origin, err := url.Parse(strings.TrimSpace(r.Header.Get("Origin")))
	if err != nil || !sameOrigin(origin, s.deps.PublicURL) ||
		!strings.EqualFold(strings.TrimSpace(r.Header.Get("Sec-Fetch-Site")), "same-origin") {
		s.writeError(w, r, http.StatusForbidden, "BACKUP_UPLOAD_ORIGIN_REQUIRED",
			"Backup uploads must originate from this administration application.")
		return false
	}
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" || len(key) > 128 {
		s.writeError(w, r, http.StatusBadRequest, "IDEMPOTENCY_KEY_REQUIRED",
			"Send an Idempotency-Key header containing 1 to 128 characters.")
		return false
	}
	return true
}

func sameOrigin(actual, expected *url.URL) bool {
	if actual == nil || expected == nil || actual.IsAbs() == false || actual.User != nil ||
		actual.Path != "" || actual.RawQuery != "" || actual.Fragment != "" {
		return false
	}
	return strings.EqualFold(actual.Scheme, expected.Scheme) &&
		strings.EqualFold(actual.Hostname(), expected.Hostname()) &&
		effectivePort(actual) == effectivePort(expected)
}

func effectivePort(value *url.URL) string {
	if port := value.Port(); port != "" {
		return port
	}
	if strings.EqualFold(value.Scheme, "https") {
		return "443"
	}
	if strings.EqualFold(value.Scheme, "http") {
		return "80"
	}
	return ""
}
