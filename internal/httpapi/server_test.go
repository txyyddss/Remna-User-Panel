package httpapi

import (
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestDecodeJSONAcceptsOneStrictDocument(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{name: "valid object", body: `{"name":"Ada"}`},
		{name: "empty body", body: ``, wantErr: true},
		{name: "unknown field", body: `{"name":"Ada","role":"admin"}`, wantErr: true},
		{name: "second document", body: `{"name":"Ada"} {"name":"Grace"}`, wantErr: true},
		{name: "trailing invalid bytes", body: `{"name":"Ada"} trailing`, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest("POST", "/", strings.NewReader(test.body))
			response := httptest.NewRecorder()
			var input struct {
				Name string `json:"name"`
			}
			err := decodeJSON(response, request, &input)
			if test.wantErr && err == nil {
				t.Fatal("decodeJSON() unexpectedly succeeded")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("decodeJSON(): %v", err)
			}
		})
	}
}

func TestSPAHandlerServesAssetsAndFallsBackToIndex(t *testing.T) {
	t.Parallel()

	content := fstest.MapFS{
		"index.html":    {Data: []byte("<main>TX Carpool</main>")},
		"assets/app.js": {Data: []byte("window.txc = true")},
	}
	handler := spaHandler(content)

	tests := []struct {
		name        string
		path        string
		wantBody    string
		contentType string
	}{
		{name: "static asset", path: "/assets/app.js", wantBody: "window.txc = true", contentType: "text/javascript"},
		{name: "client route", path: "/catalog/checkout", wantBody: "<main>TX Carpool</main>", contentType: "text/html"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest("GET", test.path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != 200 {
				t.Fatalf("status = %d, want 200", response.Code)
			}
			if got := strings.TrimSpace(response.Body.String()); got != test.wantBody {
				t.Fatalf("body = %q, want %q", got, test.wantBody)
			}
			if got := response.Header().Get("Content-Type"); !strings.Contains(got, test.contentType) {
				t.Fatalf("Content-Type = %q, want it to contain %q", got, test.contentType)
			}
		})
	}
}
