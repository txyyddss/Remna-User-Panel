package connectivity

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestXrayProbeTimeoutAndCancellation(t *testing.T) {
	for _, cancelParent := range []bool{false, true} {
		name := "timeout"
		if cancelParent {
			name = "parent cancellation"
		}
		t.Run(name, func(t *testing.T) {
			started := make(chan struct{}, 1)
			server, probe := trustedProbeTarget(t, func(writer http.ResponseWriter, request *http.Request) {
				started <- struct{}{}
				<-request.Context().Done()
			})
			proxyAddress, _ := nativeVLESSFixture(t, probeUserUUID, server.URL)
			target := probeResolvedTarget(t, proxyAddress, probeUserUUID)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			result := make(chan Outcome, 1)
			go func() {
				result <- probe.Check(ctx, target, Config{ProbeURL: server.URL, TimeoutSeconds: 1})
			}()
			select {
			case <-started:
			case <-time.After(3 * time.Second):
				t.Fatal("authenticated probe did not reach TLS target")
			}
			if cancelParent {
				cancel()
			}
			select {
			case outcome := <-result:
				expectedStatus, expectedCode := "failed", "CONNECTIVITY_PROBE_TIMEOUT"
				if cancelParent {
					expectedStatus, expectedCode = "interrupted", "CONNECTIVITY_INTERRUPTED"
				}
				if outcome.Status != expectedStatus || outcome.ErrorCode != expectedCode {
					t.Fatalf("cancellation = %+v", outcome)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("probe did not stop within its deadline")
			}
		})
	}
}

func TestXrayProbeLifecycleWaitIsCancellable(t *testing.T) {
	xrayLifecycle <- struct{}{}
	defer func() { <-xrayLifecycle }()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	outcome := NewXrayProbe().Check(ctx, Target{}, Config{ProbeURL: "https://example.com", TimeoutSeconds: 1})
	if outcome.Status != "interrupted" || outcome.ErrorCode != "CONNECTIVITY_INTERRUPTED" {
		t.Fatalf("cancelled lifecycle wait = %+v", outcome)
	}
}

func TestXrayProbeRejectsHTTPAndOversizedTimeout(t *testing.T) {
	target := probeResolvedTarget(t, "127.0.0.1:443", probeUserUUID)
	for _, config := range []Config{{ProbeURL: "http://example.com", TimeoutSeconds: 1}, {ProbeURL: "https://example.com", TimeoutSeconds: 61}} {
		outcome := NewXrayProbe().Check(context.Background(), target, config)
		if outcome.Status != "error" || outcome.ErrorCode != "CONNECTIVITY_INVALID_RESOLVED_CONFIG" {
			t.Fatalf("invalid probe options = %+v", outcome)
		}
	}
}
