package remnawave

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestUserResponseRequiresReferenceFields(t *testing.T) {
	t.Parallel()
	var user User
	if err := json.Unmarshal([]byte(`{"id":9,"username":"ada"}`), &user); err == nil {
		t.Fatal("User.UnmarshalJSON() accepted a partial upstream response")
	}
	if err := json.Unmarshal([]byte(userJSON(9, "ada", 42)), &user); err != nil {
		t.Fatalf("User.UnmarshalJSON() rejected reference fixture: %v", err)
	}
}

func TestUserResponseValidatesDiscardedReferenceFieldTypes(t *testing.T) {
	t.Parallel()
	valid := userJSON(9, "ada", 42)
	tests := []struct {
		name        string
		oldFragment string
		newFragment string
	}{
		{name: "email", oldFragment: `"email":null`, newFragment: `"email":{}`},
		{name: "tag", oldFragment: `"tag":null`, newFragment: `"tag":7`},
		{name: "hardware limit", oldFragment: `"hwidDeviceLimit":null`, newFragment: `"hwidDeviceLimit":"7"`},
		{name: "threshold", oldFragment: `"lastTriggeredThreshold":0`, newFragment: `"lastTriggeredThreshold":0.5`},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var user User
			malformed := strings.Replace(valid, test.oldFragment, test.newFragment, 1)
			if err := json.Unmarshal([]byte(malformed), &user); err == nil {
				t.Fatal("User.UnmarshalJSON() accepted a discarded field with the wrong type")
			}
		})
	}
}

func userJSON(id int64, username string, telegramID int64) string {
	return fmt.Sprintf(`{"id":%d,"shortUuid":"short","username":%q,"status":"ACTIVE","trafficLimitBytes":0,"trafficLimitStrategy":"NO_RESET","expireAt":"2099-12-31T23:59:59Z","telegramId":%d,"email":null,"description":null,"tag":null,"hwidDeviceLimit":null,"externalSquadUuid":null,"trojanPassword":"trojan-secret","vlessUuid":"11111111-1111-4111-8111-111111111111","ssPassword":"ss-secret","lastTriggeredThreshold":0,"subRevokedAt":null,"lastTrafficResetAt":null,"createdAt":"2026-08-01T00:00:00Z","updatedAt":"2026-08-01T00:00:00Z","subscriptionUrl":"https://secret.example/sub","activeInternalSquads":[],"userTraffic":{"usedTrafficBytes":0,"lifetimeUsedTrafficBytes":0,"onlineAt":null,"firstConnectedAt":null,"lastConnectedNodeUuid":null}}`, id, username, telegramID)
}
