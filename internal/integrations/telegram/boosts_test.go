package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUserChatBoosts(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request memberRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if r.URL.Path != "/bot123:test/getUserChatBoosts" || request.ChatID != "-100123456" || request.UserID != 42 {
			t.Errorf("unexpected request: %s %+v", r.URL.Path, request)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"result":{"boosts":[{"boost_id":"first","add_date":100,"expiration_date":300}]}}`))
	}))
	defer server.Close()
	client, err := NewClient("123:test", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.GetUserChatBoosts(context.Background(), "-100123456", 42)
	if err != nil || len(result.Boosts) != 1 || result.Boosts[0].BoostID != "first" || result.Boosts[0].ExpirationDate != 300 {
		t.Fatalf("GetUserChatBoosts() = (%+v, %v)", result, err)
	}
	if _, err := client.GetUserChatBoosts(context.Background(), "", 42); err == nil {
		t.Fatal("empty chat accepted")
	}
	if _, err := client.GetUserChatBoosts(context.Background(), "-100123456", 0); err == nil {
		t.Fatal("invalid user accepted")
	}
}

func TestMessageDecodesSenderBoostCount(t *testing.T) {
	var message Message
	if err := json.Unmarshal([]byte(`{"message_id":1,"sender_boost_count":3,"chat":{"id":-100123456,"type":"supergroup"}}`), &message); err != nil {
		t.Fatal(err)
	}
	if message.SenderBoostCount != 3 {
		t.Fatalf("sender count = %d", message.SenderBoostCount)
	}
}
