package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAttributedPMTextWirePreservesEntitiesAndReply(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Error(err)
		}
		if _, ok := input["parse_mode"]; ok {
			t.Error("formatting reinterpreted")
		}
		var text string
		_ = json.Unmarshal(input["text"], &text)
		var entities []MessageEntity
		_ = json.Unmarshal(input["entities"], &entities)
		if text != "Hi 👋\n\nBy @casey_ops" || len(entities) != 1 || entities[0].Offset != 3 || entities[0].Length != 2 {
			t.Error("text/entity loss")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":73,"chat":{"id":42}}}`))
	}))
	defer server.Close()
	client, err := NewClient("123:token", WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatal(err)
	}
	id, err := client.SendPMText(context.Background(), PMTextRequest{ChatID: 42, Text: "Hi 👋\n\nBy @casey_ops", Entities: []MessageEntity{{Type: "bold", Offset: 3, Length: 2}}, ReplyParameters: &PMReplyParameters{MessageID: 62}, DisableNotification: true})
	if err != nil || id != 73 {
		t.Fatalf("id=%d err=%v", id, err)
	}
	if UTF16Length("👋") != 2 {
		t.Fatal("UTF-16 offsets broken")
	}
}
