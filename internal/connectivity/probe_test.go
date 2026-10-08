package connectivity

import (
	"context"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestXrayProbeAuthenticatedHTTPSStatus(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusNoContent, http.StatusFound, http.StatusServiceUnavailable} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var requests atomic.Int64
			server, probe := trustedProbeTarget(t, func(writer http.ResponseWriter, request *http.Request) {
				requests.Add(1)
				writer.Header().Set("Location", "/redirected")
				writer.WriteHeader(status)
			})
			proxyAddress, authenticated := nativeVLESSFixture(t, probeUserUUID, server.URL)
			t.Setenv("HTTPS_PROXY", "http://127.0.0.1:1")
			t.Setenv("HTTP_PROXY", "http://127.0.0.1:1")
			t.Setenv("NO_PROXY", "")
			target := probeResolvedTarget(t, proxyAddress, probeUserUUID)
			// Uppercase upstream UUIDs must resolve the same forced outbound tag.
			target.HostUUID = strings.ToUpper(target.HostUUID)
			for range 2 {
				outcome := probe.Check(context.Background(), target, Config{ProbeURL: server.URL, TimeoutSeconds: 3})
				expected := "connected"
				if status >= 300 {
					expected = "failed"
				}
				if outcome.Status != expected || outcome.HTTPStatus == nil || *outcome.HTTPStatus != status {
					t.Fatalf("status check = %+v", outcome)
				}
				if outcome.LatencyMS == nil || *outcome.LatencyMS <= 0 {
					t.Fatalf("missing TTFB: %+v", outcome)
				}
				if expected == "failed" && outcome.ErrorCode != "CONNECTIVITY_UNEXPECTED_HTTP_STATUS" {
					t.Fatalf("HTTP failure code = %q", outcome.ErrorCode)
				}
			}
			if requests.Load() != 2 || authenticated.Load() != 2 {
				t.Fatalf("redirect/reuse/proxy routing: requests=%d authenticated=%d", requests.Load(), authenticated.Load())
			}
		})
	}
}

func TestXrayProbeBadCredentialsNeverFallBackToDirect(t *testing.T) {
	var requests atomic.Int64
	server, probe := trustedProbeTarget(t, func(writer http.ResponseWriter, request *http.Request) {
		requests.Add(1)
		writer.WriteHeader(http.StatusNoContent)
	})
	proxyAddress, authenticated := nativeVLESSFixture(t, probeUserUUID, server.URL)
	target := probeResolvedTarget(t, proxyAddress, "33333333-3333-4333-8333-333333333333")
	outcome := probe.Check(context.Background(), target, Config{ProbeURL: server.URL, TimeoutSeconds: 2})
	if outcome.Status != "failed" || outcome.ErrorCode != "CONNECTIVITY_PROBE_FAILED" || outcome.HTTPStatus != nil {
		t.Fatalf("bad authentication = %+v", outcome)
	}
	if requests.Load() != 0 || authenticated.Load() != 0 {
		t.Fatalf("unauthenticated probe reached target: %d requests", requests.Load())
	}
}

func TestXrayProbeRejectsUntrustedHTTPSCertificate(t *testing.T) {
	server, _ := trustedProbeTarget(t, func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	})
	proxyAddress, authenticated := nativeVLESSFixture(t, probeUserUUID, server.URL)
	outcome := NewXrayProbe().Check(context.Background(), probeResolvedTarget(t, proxyAddress, probeUserUUID), Config{ProbeURL: server.URL, TimeoutSeconds: 3})
	if outcome.Status != "failed" || outcome.ErrorCode != "CONNECTIVITY_PROBE_FAILED" || authenticated.Load() != 1 {
		t.Fatalf("untrusted certificate = %+v; authenticated=%d", outcome, authenticated.Load())
	}
}
