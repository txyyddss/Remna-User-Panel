package validation

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestText(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		value     string
		multiline bool
		wantErr   error
	}{
		{name: "unicode password", value: " 密碼 päss 🧩 ", multiline: true},
		{name: "multiline CSV", value: "code,name\nA1,测试", multiline: true},
		{name: "single line rejects newline", value: "one\ntwo", wantErr: ErrControlCharacter},
		{name: "NUL rejected", value: "one\x00two", multiline: true, wantErr: ErrControlCharacter},
		{name: "invalid UTF-8", value: string([]byte{0xff}), wantErr: ErrInvalidUTF8},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := Text("value", test.value, 128, test.multiline)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Text() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestRequest(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		mutate    func(*testing.T) *http.Request
		wantError bool
	}{
		{
			name: "printable encoded query",
			mutate: func(t *testing.T) *http.Request {
				t.Helper()
				request := httptest.NewRequest("GET", "/api/v1/onboarding/content?locale=zh-CN&next=https%3A%2F%2Fexample.test%2Fa%3Fb%3D1", nil)
				request.Header.Set("X-Client", "Telegram WebView")
				return request
			},
		},
		{
			name: "query control",
			mutate: func(t *testing.T) *http.Request {
				t.Helper()
				return httptest.NewRequest("GET", "/?value=%00", nil)
			},
			wantError: true,
		},
		{
			name: "path control",
			mutate: func(t *testing.T) *http.Request {
				t.Helper()
				request := httptest.NewRequest("GET", "/", nil)
				request.URL.Path = "/api/\x00value"
				return request
			},
			wantError: true,
		},
		{
			name: "header control",
			mutate: func(t *testing.T) *http.Request {
				t.Helper()
				request := httptest.NewRequest("GET", "/", nil)
				request.Header["X-Client"] = []string{"bad\x00value"}
				return request
			},
			wantError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := Request(test.mutate(t))
			if (err != nil) != test.wantError {
				t.Fatalf("Request() error = %v, wantError %t", err, test.wantError)
			}
		})
	}
}

func TestJSONDocument(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		body    []byte
		wantErr bool
	}{
		{name: "password and initData", body: []byte(`{"password":" 密碼 päss 🧩 ","initData":"query_id=A%2FB&user=%7B%22id%22%3A1%7D"}`)},
		{name: "multiline and CSV", body: []byte(`{"text":"line one\nline two","csv":"code,name\nA1,测试"}`)},
		{name: "NUL escape", body: []byte(`{"text":"bad\u0000value"}`), wantErr: true},
		{name: "invalid UTF-8", body: []byte{'{', '"', 'x', '"', ':', '"', 0xff, '"', '}'}, wantErr: true},
		{name: "multiple documents", body: []byte(`{} {}`), wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := JSONDocument(test.body)
			if (err != nil) != test.wantErr {
				t.Fatalf("JSONDocument() error = %v, wantErr %t", err, test.wantErr)
			}
		})
	}
}
