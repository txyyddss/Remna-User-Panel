package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendMarkdownV2MessageDecodesMessageResult(t *testing.T) {
	t.Parallel()
	const token = "123:token"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() { _ = request.Body.Close() }()
		if request.URL.Path != "/bot"+token+"/sendMessage" {
			t.Errorf("path = %s", request.URL.Path)
		}
		var body struct {
			Text            string `json:"text"`
			ParseMode       string `json:"parse_mode"`
			ReplyParameters *struct {
				MessageID                int64 `json:"message_id"`
				AllowSendingWithoutReply bool  `json:"allow_sending_without_reply"`
			} `json:"reply_parameters"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if body.Text != "*paid*" || body.ParseMode != "MarkdownV2" {
			t.Errorf("request = %+v", body)
		}
		if body.ReplyParameters == nil || body.ReplyParameters.MessageID != 9 || !body.ReplyParameters.AllowSendingWithoutReply {
			t.Errorf("reply parameters = %+v", body.ReplyParameters)
		}
		_, _ = writer.Write([]byte(`{"ok":true,"result":{"message_id":7}}`))
	}))
	defer server.Close()

	client, err := NewClient(token, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if err := client.SendMarkdownV2Message(context.Background(), -100, 9, "*paid*"); err != nil {
		t.Fatalf("SendMarkdownV2Message() error = %v", err)
	}
}
