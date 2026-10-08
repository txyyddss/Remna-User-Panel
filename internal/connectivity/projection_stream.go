package connectivity

import "strings"

func projectStream(host map[string]any) (map[string]any, error) {
	network := stringValue(host["transport"])
	options := configObject(host["transportOptions"])
	stream := map[string]any{"network": network, "security": host["security"]}
	var settings map[string]any
	var key string
	switch network {
	case "tcp":
		key, settings = "tcpSettings", pickFields(options, "header")
	case "ws":
		key, settings = "wsSettings", pickFields(options, "path", "host", "headers", "heartbeatPeriod")
	case "httpupgrade":
		key, settings = "httpupgradeSettings", pickFields(options, "path", "host", "headers")
	case "xhttp":
		key, settings = "xhttpSettings", pickFields(options, "path", "host", "mode", "extra")
	case "grpc":
		key, settings = "grpcSettings", pickFields(options, "serviceName", "authority", "multiMode")
	case "kcp":
		key, settings = "kcpSettings", pickFields(options, "congestion")
		if value := options["clientMtu"]; value != nil {
			settings["mtu"] = value
		}
		if value := options["clientTti"]; value != nil {
			settings["tti"] = value
		}
	case "hysteria":
		if numberValue(options["version"]) != 2 || !nonemptyString(options["auth"]) {
			return nil, projectionError("INVALID_RESOLVED_CONFIG")
		}
		key, settings = "hysteriaSettings", pickFields(options, "version", "auth")
	default:
		return nil, projectionError("UNSUPPORTED_TRANSPORT")
	}
	stream[key] = settings
	security := configObject(host["securityOptions"])
	switch stringValue(host["security"]) {
	case "none":
	case "tls":
		tls := pickFields(security, "serverName", "enableSessionResumption", "fingerprint", "pinnedPeerCertSha256", "verifyPeerCertByName", "echForceQuery", "echConfigList", "echSockopt", "cipherSuites")
		if alpn := stringValue(security["alpn"]); alpn != "" {
			tls["alpn"] = strings.Split(alpn, ",")
		}
		if !nonemptyString(tls["fingerprint"]) {
			delete(tls, "fingerprint")
		}
		stream["tlsSettings"] = tls
	case "reality":
		stream["realitySettings"] = pickFields(security, "serverName", "publicKey", "mldsa65Verify", "shortId", "spiderX", "fingerprint")
	default:
		return nil, projectionError("UNSUPPORTED_SECURITY")
	}
	for key, value := range pickFields(configObject(host["streamOverrides"]), "sockopt", "finalMask") {
		if key == "finalMask" {
			key = "finalmask"
		}
		stream[key] = value
	}
	return stream, nil
}
