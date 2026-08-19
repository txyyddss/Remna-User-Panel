package remnawave

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserConnectionsTreatsCompletedUnsuccessfulResultAsFailed(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/api/connections/by-user/scan-1" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		_, _ = writer.Write([]byte(`{"response":{"isCompleted":true,"isFailed":false,"progress":{"percent":100},"result":{"success":false,"nodes":[]}}}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "token", WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	scan, err := client.UserConnections(context.Background(), "scan-1")
	if err != nil || !scan.Completed || !scan.Failed {
		t.Fatalf("UserConnections() = (%+v, %v)", scan, err)
	}
}
