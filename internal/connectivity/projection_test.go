package connectivity

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestProjectTargetHasExactlyOneAuthenticatedOutbound(t *testing.T) {
	target := probeResolvedTarget(t, "example.com:443", probeUserUUID)
	data, err := projectTarget(target)
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	if len(config["inbounds"].([]any)) != 0 || len(config["outbounds"].([]any)) != 1 {
		t.Fatalf("unexpected probe listeners or fallback: %s", data)
	}
	outbound := singleConfigObject(config["outbounds"])
	if outbound["protocol"] != "vless" || config["routing"] != nil || config["dns"] != nil {
		t.Fatal("probe introduced a different outbound, routing, or DNS service")
	}
}

func TestProjectDocumentedProtocols(t *testing.T) {
	for _, item := range []struct {
		protocol  string
		options   map[string]any
		transport string
	}{
		{"vless", map[string]any{"id": probeUserUUID, "encryption": "none", "flow": "xtls-rprx-vision"}, "tcp"},
		{"trojan", map[string]any{"password": "test-password"}, "tcp"},
		{"shadowsocks", map[string]any{"password": "test-password", "method": "aes-128-gcm", "uot": true, "uotVersion": 2}, "tcp"},
		{"hysteria", map[string]any{"version": 2}, "hysteria"},
	} {
		t.Run(item.protocol, func(t *testing.T) {
			target := mutateProbeTarget(t, probeResolvedTarget(t, "example.com:443", probeUserUUID), func(host map[string]any) {
				host["protocol"], host["protocolOptions"], host["transport"] = item.protocol, item.options, item.transport
				if item.transport == "hysteria" {
					host["transportOptions"] = map[string]any{"version": 2, "auth": "test-auth"}
				}
			})
			outbound, err := projectOutbound(target)
			if err != nil || outbound["protocol"] != item.protocol {
				t.Fatalf("protocol projection failed: %v", err)
			}
			if item.protocol == "shadowsocks" {
				server := singleConfigObject(configObject(outbound["settings"])["servers"])
				if server["uot"] != true || numberValue(server["UoTVersion"]) != 2 {
					t.Fatal("Shadowsocks UoT settings were lost")
				}
			}
		})
	}
}

func TestProjectPreservesTransportAndSecurityOverrides(t *testing.T) {
	target := mutateProbeTarget(t, probeResolvedTarget(t, "example.com:443", probeUserUUID), func(host map[string]any) {
		host["transport"] = "ws"
		host["transportOptions"] = map[string]any{"path": "/tunnel", "host": "edge.example.com", "headers": map[string]any{"X-Test": "value", "dialerProxy": "header-only"}, "heartbeatPeriod": 30}
		host["security"] = "tls"
		host["securityOptions"] = map[string]any{"serverName": "edge.example.com", "alpn": "h2,http/1.1", "echConfigList": "fixture-ech", "pinnedPeerCertSha256": "fixture-pin"}
		host["streamOverrides"] = map[string]any{"sockopt": map[string]any{"tcpFastOpen": true}, "finalMask": map[string]any{"udp": []any{map[string]any{"type": "salamander", "settings": map[string]any{"password": "fixture-mask"}}}}}
	})
	outbound, err := projectOutbound(target)
	if err != nil {
		t.Fatal(err)
	}
	stream := configObject(outbound["streamSettings"])
	ws, tls := configObject(stream["wsSettings"]), configObject(stream["tlsSettings"])
	if ws["path"] != "/tunnel" || ws["host"] != "edge.example.com" || numberValue(ws["heartbeatPeriod"]) != 30 || configObject(ws["headers"])["X-Test"] != "value" {
		t.Fatal("WebSocket options were lost")
	}
	if configObject(ws["headers"])["dialerProxy"] != "header-only" {
		t.Fatal("HTTP header data was treated as a proxy chain")
	}
	if !reflect.DeepEqual(tls["alpn"], []string{"h2", "http/1.1"}) || tls["echConfigList"] != "fixture-ech" || tls["pinnedPeerCertSha256"] != "fixture-pin" {
		t.Fatal("TLS options were lost")
	}
	if configObject(stream["sockopt"])["tcpFastOpen"] != true || stream["finalmask"] == nil || stream["finalMask"] != nil {
		t.Fatal("stream overrides were lost or incorrectly named")
	}
}

func TestProjectTransportSpecificFields(t *testing.T) {
	for _, item := range []struct {
		transport, settings string
		options             map[string]any
		field               string
		value               any
	}{
		{"tcp", "tcpSettings", map[string]any{"header": map[string]any{"type": "http"}}, "header", map[string]any{"type": "http"}},
		{"grpc", "grpcSettings", map[string]any{"serviceName": "tunnel", "authority": "edge.example.com", "multiMode": true}, "authority", "edge.example.com"},
		{"httpupgrade", "httpupgradeSettings", map[string]any{"path": "/up", "host": "edge.example.com"}, "path", "/up"},
		{"xhttp", "xhttpSettings", map[string]any{"path": "/x", "mode": "stream-one", "extra": map[string]any{"noGRPCHeader": true}}, "extra", map[string]any{"noGRPCHeader": true}},
		{"kcp", "kcpSettings", map[string]any{"clientMtu": 1350, "clientTti": 20}, "mtu", json.Number("1350")},
	} {
		t.Run(item.transport, func(t *testing.T) {
			target := mutateProbeTarget(t, probeResolvedTarget(t, "example.com:443", probeUserUUID), func(host map[string]any) {
				host["transport"], host["transportOptions"] = item.transport, item.options
			})
			outbound, err := projectOutbound(target)
			if err != nil {
				t.Fatal(err)
			}
			settings := configObject(configObject(outbound["streamSettings"])[item.settings])
			if !reflect.DeepEqual(settings[item.field], item.value) {
				t.Fatalf("transport field %s was changed", item.field)
			}
		})
	}
}
