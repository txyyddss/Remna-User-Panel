package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) booleanCall(ctx context.Context, method string, payload any) error {
	if request, ok := payload.(memberRequest); ok {
		if err := validateMemberRequest(request.ChatID, request.UserID); err != nil {
			return err
		}
	}
	var result bool
	if err := c.call(ctx, method, payload, &result); err != nil {
		return err
	}
	if !result {
		return fmt.Errorf("telegram %s returned false", method)
	}
	return nil
}

func (c *Client) call(ctx context.Context, method string, payload, result any) error {
	if ctx == nil {
		return fmt.Errorf("telegram %s: context is nil", method)
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("telegram %s encode request: %w", method, err)
	}
	endpoint := *c.baseURL
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/bot" + c.token + "/" + method
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram %s create request: %w", method, err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		var urlError *url.Error
		if errors.As(err, &urlError) {
			return fmt.Errorf("telegram %s request: %w", method, urlError.Err)
		}
		return fmt.Errorf("telegram %s request: %w", method, err)
	}
	defer func() { _ = response.Body.Close() }()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return fmt.Errorf("telegram %s read response: %w", method, err)
	}
	if len(responseBody) > maxResponseBytes {
		return fmt.Errorf("telegram %s response exceeds %d bytes", method, maxResponseBytes)
	}

	var envelope struct {
		OK          bool            `json:"ok"`
		Result      json.RawMessage `json:"result"`
		ErrorCode   int             `json:"error_code"`
		Description string          `json:"description"`
		Parameters  struct {
			RetryAfter int `json:"retry_after"`
		} `json:"parameters"`
	}
	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		return fmt.Errorf("telegram %s decode response (http=%d): %w", method, response.StatusCode, err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || !envelope.OK {
		return &APIError{Method: method, HTTPStatus: response.StatusCode, ErrorCode: envelope.ErrorCode, Description: envelope.Description, RetryAfter: envelope.Parameters.RetryAfter}
	}
	if result == nil {
		return nil
	}
	if len(envelope.Result) == 0 || string(envelope.Result) == "null" {
		return fmt.Errorf("telegram %s response has no result", method)
	}
	if err := json.Unmarshal(envelope.Result, result); err != nil {
		return fmt.Errorf("telegram %s decode result: %w", method, err)
	}
	return nil
}

func parseBaseURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return nil, errors.New("base URL must use http or https")
	}
	if u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("base URL must be an absolute URL without credentials, query, or fragment")
	}
	return u, nil
}

func validateHTTPSURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if u.Scheme != "https" || u.Host == "" || u.User != nil {
		return errors.New("URL must be an absolute HTTPS URL without credentials")
	}
	return nil
}

func validWebhookSecret(secret string) bool {
	if len(secret) == 0 || len(secret) > 256 {
		return false
	}
	for _, character := range []byte(secret) {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
}

func validBotToken(token string) bool {
	if token == "" {
		return false
	}
	colon := strings.IndexByte(token, ':')
	if colon < 1 || colon == len(token)-1 {
		return false
	}
	for index, character := range []byte(token) {
		if index < colon {
			if character < '0' || character > '9' {
				return false
			}
			continue
		}
		if index == colon {
			continue
		}
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
}
