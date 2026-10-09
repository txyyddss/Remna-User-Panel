package remnawave

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRawSubscriptionProtectedContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/subscriptions/by-short-uuid/monitor-key/raw" ||
			r.Header.Get("Authorization") != "Bearer fixture-token" || r.URL.RawQuery != "" {
			t.Errorf("unexpected raw subscription request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"response":{"user":{"id":7,"shortUuid":"monitor-key","status":"ACTIVE"},"convertedUserInfo":{"hwidCheckup":{"subscriptionAllowed":false}},"resolvedProxyConfigs":[]}}`)
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "fixture-token")
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.GetRawSubscription(context.Background(), "monitor-key")
	if err != nil {
		t.Fatal(err)
	}
	if result.ResolvedProxyConfigs == nil || result.User.ID != 7 || result.ConvertedUserInfo.HWIDCheckup.SubscriptionAllowed {
		t.Fatal("raw identity, empty collection or HWID restriction was lost")
	}
}

func TestRawSubscriptionRejectsMismatchedIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"response":{"user":{"id":7,"shortUuid":"different","status":"ACTIVE"},"resolvedProxyConfigs":[]}}`)
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "fixture-token")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetRawSubscription(context.Background(), "monitor-key"); err == nil {
		t.Fatal("mismatched subscription identity was accepted")
	}
}
