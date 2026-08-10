package telegram

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestClientDoesNotForwardBotTokenAcrossRedirect(t *testing.T) {
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
	client, err := NewClient("123:never-forward", WithBaseURL(redirect.URL))
	if err != nil {
		t.Fatal(err)
	}
	if err := client.call(context.Background(), "getMe", map[string]string{"secret": "body"}, nil); err == nil {
		t.Fatal("Telegram call followed a provider redirect")
	}
	if sinkCalls.Load() != 0 {
		t.Fatal("redirect target received the Telegram bot token or request body")
	}
}
