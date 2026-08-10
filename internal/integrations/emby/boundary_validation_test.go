package emby

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserCollectionsRejectMalformedReturnedIDs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		response string
		call     func(*Client) error
	}{
		{name: "list", response: `[{"Name":"ada","Id":"not-a-guid","Policy":{}}]`, call: func(client *Client) error {
			_, err := client.ListUsers(context.Background())
			return err
		}},
		{name: "create", response: `{"Name":"ada","Id":"not-a-guid","Policy":{}}`, call: func(client *Client) error {
			_, err := client.CreateUser(context.Background(), "ada")
			return err
		}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				_, _ = writer.Write([]byte(test.response))
			}))
			defer server.Close()
			client, err := NewClient(server.URL, "token", WithHTTPClient(server.Client()))
			if err != nil {
				t.Fatal(err)
			}
			if err := test.call(client); err == nil {
				t.Fatal("Emby operation accepted a malformed returned Id")
			}
		})
	}
}
