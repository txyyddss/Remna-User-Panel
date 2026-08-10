package outbox

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

// TargetID extracts one required identifier from a canonical typed job payload.
func TargetID(job model.OutboxJob, field string) (string, error) {
	var payload map[string]string
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		return "", fmt.Errorf("decode %s job payload: %w", job.Kind, err)
	}
	value := strings.TrimSpace(payload[field])
	if value == "" {
		return "", errors.New("outbox payload is missing " + field)
	}
	return value, nil
}
