package remnawave

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserConnectionsTreatsCompletedUnsuccessfulResultAsFailed(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/api/connections/by-user/scan-1" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		_, _ = writer.Write([]byte(`{"response":{"isCompleted":true,"isFailed":false,"progress":{"percent":100},"result":{"success":false,"nodes":[]}}}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "token", WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	scan, err := client.UserConnections(context.Background(), "scan-1")
	if err != nil || !scan.Completed || !scan.Failed {
		t.Fatalf("UserConnections() = (%+v, %v)", scan, err)
	}
}

func TestUserConnectionsPreservesCompletedNodesWhenResultSuccessIsFalse(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"response":{"isCompleted":true,"isFailed":false,"progress":{"percent":100},"result":{"success":false,"nodes":[{"nodeUuid":"6aa6d759-20de-4b11-8c0a-8e0daee3a4ee","nodeName":"Tokyo","countryCode":"JP","ips":[{"ip":"203.0.113.1","lastSeen":"2026-08-19T00:00:00Z"}]}]}}}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "token", WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	scan, err := client.UserConnections(context.Background(), "scan-1")
	if err != nil || !scan.Completed || scan.Failed || len(scan.Nodes) != 1 {
		t.Fatalf("UserConnections() = (%+v, %v)", scan, err)
	}
}

func TestUserConnectionsNormalizesIPv6Observations(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"response":{"isCompleted":true,"isFailed":false,"progress":{"percent":100},"result":{"success":true,"nodes":[{"nodeUuid":"6aa6d759-20de-4b11-8c0a-8e0daee3a4ee","nodeName":"Tokyo","countryCode":"JP","ips":[{"ip":"2602:fbf1:b002::1009","lastSeen":"2026-08-19T00:00:00Z"},{"ip":"[2001:db8::7]:443","lastSeen":"2026-08-19T00:00:00Z"}] }]}}}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "token", WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	scan, err := client.UserConnections(context.Background(), "scan-1")
	if err != nil || len(scan.Nodes) != 1 || len(scan.Nodes[0].IPs) != 2 {
		t.Fatalf("UserConnections() = (%+v, %v)", scan, err)
	}
	if scan.Nodes[0].IPs[0].IP != "2602:fbf1:b002::1009" || scan.Nodes[0].IPs[1].IP != "2001:db8::7" {
		t.Fatalf("normalized IPs = %+v", scan.Nodes[0].IPs)
	}
}
