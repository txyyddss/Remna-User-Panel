package validation

import (
	"fmt"
	"net/http"
	"net/url"
)

const (
	maxPathBytes        = 4 << 10
	maxRawQueryBytes    = 16 << 10
	maxQueryElement     = 8 << 10
	maxHeaderValueBytes = 16 << 10
)

// Request validates the request line, decoded query fields, and headers.
// Body formats are validated by their format-specific decoders.
func Request(request *http.Request) error {
	if request == nil || request.URL == nil {
		return &FieldError{Field: "request", Err: ErrPattern}
	}
	if !methodPattern.MatchString(request.Method) {
		return &FieldError{Field: "method", Err: ErrPattern}
	}
	escapedPath := request.URL.EscapedPath()
	if escapedPath == "" {
		escapedPath = "/"
	}
	if err := Text("path", escapedPath, maxPathBytes, false); err != nil {
		return err
	}
	if err := Text("decoded path", request.URL.Path, maxPathBytes, false); err != nil {
		return err
	}
	if !escapedPathPattern.MatchString(escapedPath) || !decodedPathPattern.MatchString(request.URL.Path) {
		return &FieldError{Field: "path", Err: ErrPattern}
	}
	if err := validateQuery(request.URL.RawQuery); err != nil {
		return err
	}
	return validateHeaders(request.Header)
}

func validateQuery(raw string) error {
	if err := Text("query", raw, maxRawQueryBytes, false); err != nil {
		return err
	}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return &FieldError{Field: "query", Err: fmt.Errorf("%w: %v", ErrPattern, err)}
	}
	for key, items := range values {
		if err := Text("query name", key, maxQueryElement, false); err != nil {
			return err
		}
		for _, item := range items {
			if err := Text("query value", item, maxQueryElement, false); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateHeaders(headers http.Header) error {
	for name, values := range headers {
		if !headerNamePattern.MatchString(name) {
			return &FieldError{Field: "header name", Err: ErrPattern}
		}
		for _, value := range values {
			if err := Text("header value", value, maxHeaderValueBytes, false); err != nil {
				return err
			}
		}
	}
	return nil
}
