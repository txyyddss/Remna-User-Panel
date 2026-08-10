package remnawave

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUserOperationsRejectMismatchedResponseIdentity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		response string
		call     func(*Client) error
	}{
		{name: "get by id", response: userJSON(10, "ada", 42), call: func(client *Client) error {
			_, err := client.GetUserByID(context.Background(), 9)
			return err
		}},
		{name: "get by username", response: userJSON(9, "other", 42), call: func(client *Client) error {
			_, err := client.GetUserByUsername(context.Background(), "ada")
			return err
		}},
		{name: "create", response: userJSON(9, "other", 42), call: func(client *Client) error {
			_, err := client.CreateUser(context.Background(), CreateUserRequest{Username: "ada", TelegramID: 42, ExpireAt: time.Now().Add(time.Hour)})
			return err
		}},
		{name: "update", response: userJSON(10, "ada", 42), call: func(client *Client) error {
			status := UserStatusDisabled
			_, err := client.UpdateUser(context.Background(), UpdateUserRequest{ID: 9, Status: &status})
			return err
		}},
		{name: "revoke", response: userJSON(10, "ada", 42), call: func(client *Client) error {
			_, err := client.RevokeSubscription(context.Background(), 9, false)
			return err
		}},
		{name: "reset", response: userJSON(10, "ada", 42), call: func(client *Client) error {
			_, err := client.ResetTraffic(context.Background(), 9)
			return err
		}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				_, _ = writer.Write([]byte(`{"response":` + test.response + `}`))
			}))
			defer server.Close()
			client, err := NewClient(server.URL, "token", WithHTTPClient(server.Client()))
			if err != nil {
				t.Fatal(err)
			}
			if err := test.call(client); err == nil {
				t.Fatal("user operation accepted a mismatched response identity")
			}
		})
	}
}

func TestResolveUserRejectsMismatchedSelector(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		selector UserSelector
	}{
		{name: "id", selector: UserSelector{ID: 9}},
		{name: "short UUID", selector: UserSelector{ShortUUID: "requested"}},
		{name: "username", selector: UserSelector{Username: "requested"}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				_, _ = writer.Write([]byte(`{"response":{"id":10,"shortUuid":"other","username":"other"}}`))
			}))
			defer server.Close()
			client, err := NewClient(server.URL, "token", WithHTTPClient(server.Client()))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := client.ResolveUser(context.Background(), test.selector); err == nil {
				t.Fatal("ResolveUser() accepted a mismatched response selector")
			}
		})
	}
}
