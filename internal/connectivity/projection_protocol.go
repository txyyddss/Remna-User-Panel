package connectivity

import (
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
)

func projectProtocol(host map[string]any) (map[string]any, error) {
	options := configObject(host["protocolOptions"])
	server := map[string]any{"address": host["address"], "port": host["port"]}
	if !nonemptyString(server["address"]) || !validProxyPort(server["port"]) || options == nil {
		return nil, projectionError("INVALID_RESOLVED_CONFIG")
	}
	switch stringValue(host["protocol"]) {
	case "vless":
		id, err := uuid.Parse(stringValue(options["id"]))
		if err != nil || id == uuid.Nil {
			return nil, projectionError("INVALID_RESOLVED_CONFIG")
		}
		user := pickFields(options, "id", "encryption", "flow")
		if !nonemptyString(user["encryption"]) {
			user["encryption"] = "none"
		}
		server["users"] = []any{user}
		return map[string]any{"vnext": []any{server}}, nil
	case "trojan":
		if !nonemptyString(options["password"]) {
			return nil, projectionError("INVALID_RESOLVED_CONFIG")
		}
		server["password"] = options["password"]
		return map[string]any{"servers": []any{server}}, nil
	case "shadowsocks":
		if !nonemptyString(options["password"]) || !nonemptyString(options["method"]) {
			return nil, projectionError("INVALID_RESOLVED_CONFIG")
		}
		for key, value := range pickFields(options, "password", "method", "uot") {
			server[key] = value
		}
		if version := options["uotVersion"]; version != nil {
			server["UoTVersion"] = version
		}
		return map[string]any{"servers": []any{server}}, nil
	case "hysteria":
		if numberValue(options["version"]) != 2 || stringValue(host["transport"]) != "hysteria" {
			return nil, projectionError("UNSUPPORTED_PROTOCOL")
		}
		server["version"] = 2
		return server, nil
	default:
		return nil, projectionError("UNSUPPORTED_PROTOCOL")
	}
}

func numberValue(value any) int64 {
	switch number := value.(type) {
	case json.Number:
		result, err := number.Int64()
		if err == nil {
			return result
		}
	case int:
		return int64(number)
	case int64:
		return number
	case float64:
		if number == float64(int64(number)) {
			return int64(number)
		}
	case string:
		result, err := strconv.ParseInt(number, 10, 64)
		if err == nil {
			return result
		}
	}
	return -1
}

func validProxyPort(value any) bool {
	port := numberValue(value)
	return port > 0 && port <= 65535
}
