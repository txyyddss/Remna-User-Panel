package bepusdt

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestClientDoesNotForwardSignedBodyAcrossRedirect(t *testing.T) {
	t.Parallel()
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
	client, err := NewClient(redirect.URL, "never-forward-token")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.CancelTransaction(context.Background(), "trade-1"); err == nil {
		t.Fatal("CancelTransaction() followed a provider redirect")
	}
	if sinkCalls.Load() != 0 {
		t.Fatal("redirect target received the BEPusdt signed request body")
	}
}
