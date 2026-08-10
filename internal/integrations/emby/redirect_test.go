package emby

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestClientDoesNotForwardCredentialsAcrossRedirect(t *testing.T) {
	t.Parallel()
	const userID = "11111111-1111-4111-8111-111111111111"
	var sinkCalls atomic.Int32
	sink := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		sinkCalls.Add(1)
	}))
	defer sink.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Location", sink.URL)
		writer.WriteHeader(http.StatusTemporaryRedirect)
	}))
	defer redirect.Close()
	client, err := NewClient(redirect.URL, "server-api-token")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.SetPassword(context.Background(), userID, nil, []byte("never-forward")); err == nil {
		t.Fatal("SetPassword() followed a provider redirect")
	}
	if sinkCalls.Load() != 0 {
		t.Fatal("redirect target received Emby credentials or password body")
	}
}
