package remnawave

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateInternalSquadRejectsMismatchedResponseIdentity(t *testing.T) {
	t.Parallel()
	const requested = "11111111-1111-4111-8111-111111111111"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"response":{"uuid":"22222222-2222-4222-8222-222222222222","name":"other"}}`))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "token", WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.UpdateInternalSquadInbounds(context.Background(), requested, nil); err == nil {
		t.Fatal("UpdateInternalSquadInbounds() accepted a mismatched response UUID")
	}
}
