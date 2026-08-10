package emby

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var embyGUIDPattern = regexp.MustCompile(`^(?:[0-9A-Fa-f]{32}|[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12})$`)

func parseBaseURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return nil, errors.New("URL must use http or https")
	}
	if parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, errors.New("URL must be absolute and contain no credentials, query, or fragment")
	}
	return parsed, nil
}

func validateID(id string) error {
	if !embyGUIDPattern.MatchString(id) {
		return errors.New("invalid Emby user Id")
	}
	return nil
}
