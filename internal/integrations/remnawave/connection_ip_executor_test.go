package remnawave

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestConnectionIPExecutorPayloads(t *testing.T) {
	t.Parallel()
	const node = "6aa6d759-20de-4b11-8c0a-8e0daee3a4ee"
	tests := []struct {
		name string
		call func(*Client) error
		want string
	}{
		{
			name: "block IPv4",
			call: func(client *Client) error { return client.BlockIP(context.Background(), "203.0.113.8", node, 259200) },
			want: `{"command":{"command":"blockIps","ips":[{"ip":"203.0.113.8","timeout":259200}]},"targetNodes":{"target":"specificNodes","nodeUuids":["` + node + `"]}}`,
		},
		{
			name: "block IPv6 canonical",
			call: func(client *Client) error { return client.BlockIP(context.Background(), "2001:0db8:0:0::1", node, 41) },
			want: `{"command":{"command":"blockIps","ips":[{"ip":"2001:db8::1","timeout":41}]},"targetNodes":{"target":"specificNodes","nodeUuids":["` + node + `"]}}`,
		},
		{
			name: "unblock",
			call: func(client *Client) error { return client.UnblockIP(context.Background(), "2001:db8::1", node) },
			want: `{"command":{"command":"unblockIps","ips":["2001:db8::1"]},"targetNodes":{"target":"specificNodes","nodeUuids":["` + node + `"]}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.Method != http.MethodPost || request.URL.Path != "/api/node-plugins/executor" {
					t.Errorf("request = %s %s", request.Method, request.URL.Path)
				}
				assertExecutorJSON(t, request, test.want)
				writer.WriteHeader(http.StatusAccepted)
			}))
			defer server.Close()
			client, err := NewClient(server.URL, "token", WithHTTPClient(server.Client()))
			if err != nil {
				t.Fatal(err)
			}
			if err := test.call(client); err != nil {
				t.Fatalf("executor call: %v", err)
			}
		})
	}
}

func assertExecutorJSON(t *testing.T, request *http.Request, wantJSON string) {
	t.Helper()
	var got, want any
	if err := json.NewDecoder(request.Body).Decode(&got); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	if err := json.Unmarshal([]byte(wantJSON), &want); err != nil {
		t.Fatalf("decode expectation: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("payload = %#v, want %#v", got, want)
	}
}

func TestConnectionIPExecutorRejectsInvalidTargets(t *testing.T) {
	t.Parallel()
	client, err := NewClient("https://example.com", "token")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.BlockIP(context.Background(), "not-an-ip", "6aa6d759-20de-4b11-8c0a-8e0daee3a4ee", 1); err == nil {
		t.Fatal("invalid IPv4/IPv6 target was accepted")
	}
	if err := client.UnblockIP(context.Background(), "203.0.113.8", "not-a-uuid"); err == nil {
		t.Fatal("invalid node UUID was accepted")
	}
}
