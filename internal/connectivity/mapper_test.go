package connectivity

import (
	"reflect"
	"testing"
)

func TestXrayMapperOrderedCopySetUnsetAndIsolation(t *testing.T) {
	fromInbound := map[string]any{"key": "original"}
	host := map[string]any{
		"metadata":        map[string]any{"rawInbound": map[string]any{"source": fromInbound}},
		"protocolOptions": map[string]any{"id": "resolved-user"},
		"clientOverrides": map[string]any{"mapper": map[string]any{"xrayJson": []any{
			map[string]any{"op": "copy", "from": "source", "to": "settings.copied"},
			map[string]any{"op": "set", "value": "changed", "to": "settings.copied.key"},
			map[string]any{"op": "copy", "from": "$host.protocolOptions.id", "to": "settings.users[0].id"},
			map[string]any{"op": "unset", "to": "settings.old"},
			map[string]any{"op": "set", "value": true, "to": "settings.flag"},
		}}},
	}
	outbound := map[string]any{"settings": map[string]any{"old": true}}
	result, err := applyXrayMapper(outbound, host)
	if err != nil {
		t.Fatal(err)
	}
	settings := configObject(result["settings"])
	if configObject(settings["copied"])["key"] != "changed" || singleConfigObject(settings["users"])["id"] != "resolved-user" || settings["old"] != nil || settings["flag"] != true {
		t.Fatal("ordered mapper semantics changed")
	}
	if fromInbound["key"] != "original" || !reflect.DeepEqual(outbound, map[string]any{"settings": map[string]any{"old": true}}) {
		t.Fatal("mapper mutated source data")
	}
}

func TestXrayMapperSkipsMissingDisallowedAndBlockedPaths(t *testing.T) {
	host := map[string]any{"metadata": map[string]any{"uuid": probeHostUUID}, "clientOverrides": map[string]any{"mapper": map[string]any{"xrayJson": []any{
		map[string]any{"op": "copy", "from": "$host.metadata.uuid", "to": "forbidden"},
		map[string]any{"op": "copy", "from": "missing.source", "to": "missing"},
		map[string]any{"op": "set", "to": "primitive.child", "value": true},
		map[string]any{"op": "unset", "to": "array.0"},
	}}}}
	result, err := applyXrayMapper(map[string]any{"primitive": false, "array": []any{"first", "second"}}, host)
	if err != nil {
		t.Fatal(err)
	}
	if result["forbidden"] != nil || result["missing"] != nil || result["primitive"] != false || !reflect.DeepEqual(result["array"], []any{nil, "second"}) {
		t.Fatal("mapper did not retain upstream skip and array-hole behavior")
	}
}

func TestXrayMapperRejectsUnboundedOrAmbiguousPaths(t *testing.T) {
	for _, path := range []string{"", "settings..users", "settings.users[999999]", "settings.__proto__.key", "settings.users[-1]"} {
		t.Run(path, func(t *testing.T) {
			if _, err := mapperSegments(path); err == nil || ErrorCode(err) != "CONNECTIVITY_INVALID_MAPPER" {
				t.Fatal("unsafe mapper path accepted")
			}
		})
	}
}
