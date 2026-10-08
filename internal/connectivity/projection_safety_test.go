package connectivity

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProjectionRejectsUnsafeOrUnsupportedConfigurations(t *testing.T) {
	for _, item := range []struct {
		name, code string
		change     func(map[string]any)
	}{
		{"wrong host identity", "HOST_IDENTITY_MISMATCH", func(host map[string]any) {
			configObject(host["metadata"])["uuid"] = "33333333-3333-4333-8333-333333333333"
		}},
		{"unknown protocol", "UNSUPPORTED_PROTOCOL", func(host map[string]any) { host["protocol"] = "freedom" }},
		{"legacy transport", "UNSUPPORTED_TRANSPORT", func(host map[string]any) { host["transport"] = "quic" }},
		{"unknown security", "UNSUPPORTED_SECURITY", func(host map[string]any) { host["security"] = "insecure" }},
		{"bad port", "INVALID_RESOLVED_CONFIG", func(host map[string]any) { host["port"] = 70000 }},
		{"empty credentials", "INVALID_RESOLVED_CONFIG", func(host map[string]any) { host["protocolOptions"] = map[string]any{} }},
		{"custom template", "UNSUPPORTED_TEMPLATE", func(host map[string]any) {
			configObject(host["clientOverrides"])["xrayJsonTemplate"] = map[string]any{"routing": map[string]any{}}
		}},
		{"chained socket", "UNSUPPORTED_CHAIN", func(host map[string]any) {
			host["streamOverrides"] = map[string]any{"sockopt": map[string]any{"dialerProxy": "other"}}
		}},
		{"case-insensitive chain", "UNSUPPORTED_CHAIN", func(host map[string]any) {
			host["streamOverrides"] = map[string]any{"sockopt": map[string]any{"DialerProxy": "other"}}
		}},
		{"mapped direct fallback", "UNSUPPORTED_PROTOCOL", func(host map[string]any) {
			probeMapper(host, map[string]any{"op": "set", "to": "protocol", "value": "freedom"})
		}},
		{"mapped chain", "UNSUPPORTED_CHAIN", func(host map[string]any) {
			probeMapper(host, map[string]any{"op": "set", "to": "proxySettings.tag", "value": "other"})
		}},
		{"local certificate file", "UNSUPPORTED_TEMPLATE", func(host map[string]any) {
			probeMapper(host, map[string]any{"op": "set", "to": "streamSettings.tlsSettings.certificates.0.certificateFile", "value": "/tmp/secret.pem"})
		}},
		{"invalid operation", "INVALID_MAPPER", func(host map[string]any) {
			probeMapper(host, map[string]any{"op": "merge", "to": "settings", "value": "test-secret"})
		}},
	} {
		t.Run(item.name, func(t *testing.T) {
			target := mutateProbeTarget(t, probeResolvedTarget(t, "example.com:443", probeUserUUID), item.change)
			_, err := projectTarget(target)
			if err == nil || ErrorCode(err) != "CONNECTIVITY_"+item.code {
				t.Fatalf("configuration rejection = %v", err)
			}
			if strings.Contains(err.Error(), "test-secret") || strings.Contains(err.Error(), "/tmp/secret.pem") {
				t.Fatal("configuration error exposed its input")
			}
		})
	}
}

func TestProjectionRejectsMalformedOrTrailingJSON(t *testing.T) {
	target := probeResolvedTarget(t, "example.com:443", probeUserUUID)
	for _, raw := range []json.RawMessage{nil, []byte(`null`), []byte(`[]`), []byte(`{"metadata":`), append(append([]byte{}, target.Resolved...), []byte(`{}`)...)} {
		target.Resolved = raw
		_, err := projectTarget(target)
		if err == nil || ErrorCode(err) != "CONNECTIVITY_INVALID_RESOLVED_CONFIG" {
			t.Fatalf("invalid document rejection = %v", err)
		}
	}
}

func probeMapper(host map[string]any, operations ...any) {
	configObject(host["clientOverrides"])["mapper"] = map[string]any{"xrayJson": operations}
}
