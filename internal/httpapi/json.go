package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxJSONBodyBytes = 1 << 20

func decodeJSONBody(r *http.Request, target any) error {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(strings.ToLower(contentType), "application/json") {
		return fmt.Errorf("content_type_must_be_json")
	}
	decoder := json.NewDecoder(io.LimitReader(r.Body, maxJSONBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("json_body_must_contain_one_object")
	}
	return nil
}
