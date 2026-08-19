package remnawave

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

const geocheckNodeUUID = "373f14bc-089a-4c3a-91c3-3421e7c83367"

func TestNodeGeocheckWireContract(t *testing.T) {
	t.Parallel()
	requests := make(chan *http.Request, 2)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests <- request
		writer.Header().Set("Content-Type", "application/json")
		if request.Method == http.MethodPost {
			var body map[string]any
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Errorf("decode request: %v", err)
			}
			if len(body) != 0 {
				t.Errorf("request body = %#v, want empty object", body)
			}
			_, _ = writer.Write([]byte(`{"response":{"jobId":"job-1"}}`))
			return
		}
		_, _ = writer.Write([]byte(`{"response":{"isCompleted":true,"isFailed":false,"result":{"success":true,"nodeUuid":"` + geocheckNodeUUID + `","image":{"format":"svg","media_type":"image/svg+xml","encoding":"base64","data":"PHN2Zy8+"},"rawReport":{"ignored":true}}}}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "token", WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	jobID, err := client.RequestNodeGeocheck(context.Background(), geocheckNodeUUID)
	if err != nil || jobID != "job-1" {
		t.Fatalf("RequestNodeGeocheck() = %q, %v", jobID, err)
	}
	result, err := client.NodeGeocheckResult(context.Background(), jobID)
	if err != nil || !result.Completed || !result.Success || result.Image == nil || result.Image.Data != "PHN2Zy8+" {
		t.Fatalf("NodeGeocheckResult() = %#v, %v", result, err)
	}
	first, second := <-requests, <-requests
	if first.Method != http.MethodPost || first.URL.Path != "/api/connections/geocheck/"+geocheckNodeUUID {
		t.Fatalf("request = %s %s", first.Method, first.URL.Path)
	}
	if second.Method != http.MethodGet || second.URL.Path != "/api/connections/geocheck/job-1" {
		t.Fatalf("request = %s %s", second.Method, second.URL.Path)
	}
}

func TestNodeGeocheckRejectsEmptyJobIDAndMarksUnsuccessfulCompletionFailed(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{"response":{"isCompleted":true,"isFailed":false,"result":{"success":false,"nodeUuid":"` + geocheckNodeUUID + `","image":null}}}`))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "token", WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if _, err := client.NodeGeocheckResult(context.Background(), " "); err == nil {
		t.Fatal("NodeGeocheckResult() error = nil")
	}
	result, err := client.NodeGeocheckResult(context.Background(), "job-1")
	if err != nil || !result.Completed || !result.Failed {
		t.Fatalf("NodeGeocheckResult() = %#v, %v", result, err)
	}
}
