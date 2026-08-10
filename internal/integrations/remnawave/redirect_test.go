package remnawave

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestClientDoesNotForwardBearerAcrossRedirect(t *testing.T) {
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
	client, err := NewClient(redirect.URL, "secret-bearer")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetUserByID(context.Background(), 9); err == nil {
		t.Fatal("GetUserByID() followed a provider redirect")
	}
	if sinkCalls.Load() != 0 {
		t.Fatal("redirect target received the Remnawave bearer token")
	}
}
