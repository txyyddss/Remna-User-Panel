package remnawave

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestCreateAndUpdateUserWireContract(t *testing.T) {
	t.Parallel()
	const bearer = "secret-bearer"
	requests := make(chan remnaRequest, 2)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() { _ = request.Body.Close() }()
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		requests <- remnaRequest{method: request.Method, path: request.URL.Path, authorization: request.Header.Get("Authorization"), body: body}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"response":` + userJSON(9, "ada", 42) + `}`))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, bearer, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	user, err := client.CreateUser(context.Background(), CreateUserRequest{
		Username: "ada", ExpireAt: time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC), TelegramID: 42,
	})
	if err != nil || user.ID != 9 {
		t.Fatalf("CreateUser() = %#v, %v", user, err)
	}
	request := <-requests
	if request.method != http.MethodPost || request.path != "/api/users" || request.authorization != "Bearer "+bearer {
		t.Fatalf("create request = %#v", request)
	}
	if request.body["status"] != "ACTIVE" || request.body["trafficLimitStrategy"] != "NO_RESET" {
		t.Fatalf("create defaults = %#v", request.body)
	}
	if squads, ok := request.body["activeInternalSquads"].([]any); !ok || len(squads) != 0 {
		t.Fatalf("activeInternalSquads = %#v", request.body["activeInternalSquads"])
	}
	if value, found := request.body["externalSquadUuid"]; !found || value != nil {
		t.Fatalf("externalSquadUuid = %#v (found=%v)", value, found)
	}

	emptySquads := []string{}
	limit := int64(10_000)
	strategy := TrafficMonthly
	_, err = client.UpdateUser(context.Background(), UpdateUserRequest{
		ID: 9, TrafficLimitBytes: &limit, TrafficLimitStrategy: &strategy,
		ActiveInternalSquads: &emptySquads, ClearExternalSquad: true,
	})
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}
	request = <-requests
	if request.method != http.MethodPatch || request.path != "/api/users" || request.body["id"] != float64(9) {
		t.Fatalf("update request = %#v", request)
	}
	if value, found := request.body["externalSquadUuid"]; !found || value != nil {
		t.Fatalf("update externalSquadUuid = %#v (found=%v)", value, found)
	}
}

func TestClientEndpointMapping(t *testing.T) {
	t.Parallel()
	statsStart := time.Date(2026, time.August, 1, 14, 0, 0, 0, time.FixedZone("test", 8*60*60))
	statsEnd := time.Date(2026, time.August, 7, 14, 0, 0, 0, time.FixedZone("test", 8*60*60))
	tests := []struct {
		name     string
		method   string
		path     string
		query    url.Values
		response string
		call     func(*Client) error
	}{
		{
			name: "get by username", method: http.MethodGet, path: "/api/users/by-username/ada", response: `{"response":` + userJSON(9, "ada", 42) + `}`,
			call: func(client *Client) error {
				_, err := client.GetUserByUsername(context.Background(), "ada")
				return err
			},
		},
		{
			name: "resolve", method: http.MethodPost, path: "/api/users/resolve", response: `{"response":{"id":9,"username":"ada","shortUuid":"short"}}`,
			call: func(client *Client) error {
				_, err := client.ResolveUser(context.Background(), UserSelector{Username: "ada"})
				return err
			},
		},
		{
			name: "revoke", method: http.MethodPost, path: "/api/users/9/actions/revoke", response: `{"response":` + userJSON(9, "ada", 42) + `}`,
			call: func(client *Client) error {
				_, err := client.RevokeSubscription(context.Background(), 9, false)
				return err
			},
		},
		{
			name: "reset", method: http.MethodPost, path: "/api/users/9/actions/reset-traffic", response: `{"response":` + userJSON(9, "ada", 42) + `}`,
			call: func(client *Client) error { _, err := client.ResetTraffic(context.Background(), 9); return err },
		},
		{
			name: "subscription", method: http.MethodGet, path: "/api/subscriptions/by-id/9", response: `{"response":{"isFound":true,"subscriptionUrl":"https://secret.example/sub","links":[],"ssConfLinks":{},"user":{"username":"ada"}}}`,
			call: func(client *Client) error { _, err := client.GetSubscription(context.Background(), 9); return err },
		},
		{
			name: "stats", method: http.MethodGet, path: "/api/bandwidth-stats/users/9",
			query:    url.Values{"start": {"2026-08-01"}, "end": {"2026-08-07"}, "topNodesLimit": {"5"}},
			response: `{"response":{"categories":[],"sparklineData":[],"topNodes":[],"series":[]}}`,
			call: func(client *Client) error {
				_, err := client.GetUserStats(context.Background(), 9, statsStart, statsEnd, 5)
				return err
			},
		},
		{
			name: "squads", method: http.MethodGet, path: "/api/internal-squads", response: `{"response":{"total":1,"internalSquads":[{"uuid":"squad","name":"Free","viewPosition":1,"info":{"membersCount":2,"inboundsCount":3}}]}}`,
			call: func(client *Client) error { _, err := client.ListInternalSquads(context.Background()); return err },
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.Method != test.method || request.URL.Path != test.path {
					t.Errorf("request = %s %s, want %s %s", request.Method, request.URL.Path, test.method, test.path)
				}
				if test.query != nil && request.URL.Query().Encode() != test.query.Encode() {
					t.Errorf("query = %s, want %s", request.URL.Query().Encode(), test.query.Encode())
				}
				writer.Header().Set("Content-Type", "application/json")
				_, _ = writer.Write([]byte(test.response))
			}))
			defer server.Close()
			client, err := NewClient(server.URL, "token", WithHTTPClient(server.Client()))
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}
			if err := test.call(client); err != nil {
				t.Fatalf("call() error = %v", err)
			}
		})
	}
}

func TestInternalSquadNodeWireContract(t *testing.T) {
	t.Parallel()

	const squadUUID = "11111111-1111-4111-8111-111111111111"
	const nodeUUID = "22222222-2222-4222-8222-222222222222"
	const inboundUUID = "33333333-3333-4333-8333-333333333333"
	requestNumber := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestNumber++
		writer.Header().Set("Content-Type", "application/json")
		switch requestNumber {
		case 1:
			if request.Method != http.MethodGet || request.URL.Path != "/api/nodes" {
				t.Errorf("nodes request = %s %s", request.Method, request.URL.Path)
			}
			_, _ = writer.Write([]byte(`{"response":[{"uuid":"` + nodeUUID + `","name":"Singapore","countryCode":"SG","consumptionMultiplier":1.5,"isDisabled":false,"configProfile":{"activeInbounds":[{"uuid":"` + inboundUUID + `"}]}}]}`))
		case 2:
			if request.Method != http.MethodGet || request.URL.Path != "/api/internal-squads/"+squadUUID+"/accessible-nodes" {
				t.Errorf("accessible nodes request = %s %s", request.Method, request.URL.Path)
			}
			_, _ = writer.Write([]byte(`{"response":{"squadUuid":"` + squadUUID + `","accessibleNodes":[{"uuid":"` + nodeUUID + `","nodeName":"Singapore","countryCode":"SG","activeInbounds":["` + inboundUUID + `"]}]}}`))
		case 3:
			if request.Method != http.MethodPatch || request.URL.Path != "/api/internal-squads" {
				t.Errorf("update squad request = %s %s", request.Method, request.URL.Path)
			}
			var body struct {
				UUID     string   `json:"uuid"`
				Inbounds []string `json:"inbounds"`
			}
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Errorf("decode update body: %v", err)
			}
			if body.UUID != squadUUID || len(body.Inbounds) != 1 || body.Inbounds[0] != inboundUUID {
				t.Errorf("update squad body = %#v", body)
			}
			_, _ = writer.Write([]byte(`{"response":{"uuid":"` + squadUUID + `","name":"Squad"}}`))
		default:
			t.Errorf("unexpected request %d", requestNumber)
		}
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "token", WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	nodes, err := client.ListNodes(context.Background())
	if err != nil || len(nodes) != 1 || nodes[0].ConfigProfile.ActiveInbounds[0].UUID != inboundUUID {
		t.Fatalf("ListNodes() = (%+v, %v)", nodes, err)
	}
	accessible, err := client.InternalSquadAccessibleNodes(context.Background(), squadUUID)
	if err != nil || len(accessible) != 1 || accessible[0].UUID != nodeUUID {
		t.Fatalf("InternalSquadAccessibleNodes() = (%+v, %v)", accessible, err)
	}
	if _, err := client.UpdateInternalSquadInbounds(context.Background(), squadUUID, []string{inboundUUID}); err != nil {
		t.Fatalf("UpdateInternalSquadInbounds() error = %v", err)
	}
}

func TestAPIErrorCode(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(`{"timestamp":"2026-08-07T00:00:00Z","path":"/api/users","message":"User username already exists","errorCode":"A019"}`))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "token", WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	_, err = client.CreateUser(context.Background(), CreateUserRequest{Username: "ada", ExpireAt: time.Now().Add(time.Hour), TelegramID: 42})
	if !IsErrorCode(err, "A019") {
		t.Fatalf("CreateUser() error = %v", err)
	}
	var apiError *APIError
	if !errors.As(err, &apiError) || apiError.HTTPStatus != http.StatusBadRequest {
		t.Fatalf("CreateUser() API error = %#v", apiError)
	}
}

func TestFindUserByUsernameNotFound(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte(`{"message":"User not found","errorCode":"A025"}`))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "token", WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	user, exists, err := client.FindUserByUsername(context.Background(), "missing")
	if err != nil || exists || user != nil {
		t.Fatalf("FindUserByUsername() = %#v, %v, %v", user, exists, err)
	}
}

func TestFindUserByTelegramID(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/users/stream" || request.URL.Query().Get("size") != "1000" || request.URL.Query().Get("telegramId") != "42" {
			t.Errorf("request URL = %s", request.URL.String())
		}
		_, _ = writer.Write([]byte(`{"response":{"users":[` + userJSON(2, "two", 42) + `],"nextCursor":null,"hasMore":false}}`))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "token", WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	user, err := client.FindUserByTelegramID(context.Background(), 42)
	if err != nil || user == nil || user.ID != 2 {
		t.Fatalf("FindUserByTelegramID() = %#v, %v", user, err)
	}
}

func TestUpdateUserValidation(t *testing.T) {
	t.Parallel()
	tests := []UpdateUserRequest{
		{},
		{ID: 1, Username: "ada", Description: stringPointer("x")},
		{ID: 1},
	}
	for i, input := range tests {
		if _, err := updatePayload(input); err == nil {
			t.Errorf("case %d: updatePayload() error = nil", i)
		}
	}
}

func stringPointer(value string) *string {
	return &value
}

type remnaRequest struct {
	method        string
	path          string
	authorization string
	body          map[string]any
}
