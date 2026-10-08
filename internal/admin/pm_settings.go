package admin

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

func validatePMGroup(value string) error {
	if value == "" {
		return nil
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id >= 0 || !strings.HasPrefix(value, "-100") || len(value) <= 4 {
		return errors.New("PM destination must be a forum supergroup ID")
	}
	return nil
}

func (s *SettingsService) SetPMForumValidator(validator func(context.Context, int64) error) {
	s.pmForumValidator = validator
}

func (s *SettingsService) validatePMSettings(ctx context.Context, key, value string) error {
	if key != "telegram.pm.enabled" && key != "telegram.pm.group_chat_id" {
		return nil
	}
	if key == "telegram.pm.enabled" && value != "true" {
		return nil
	}
	group := value
	if key == "telegram.pm.enabled" {
		var err error
		group, err = s.Optional(ctx, "telegram.pm.group_chat_id")
		if err != nil {
			return err
		}
	} else {
		enabled, err := s.Optional(ctx, "telegram.pm.enabled")
		if err != nil {
			return err
		}
		if enabled != "true" {
			return nil
		}
	}
	if group == "" {
		return errors.New("a PM forum destination is required before enabling")
	}
	id, err := strconv.ParseInt(group, 10, 64)
	if err != nil {
		return err
	}
	if s.pmForumValidator == nil {
		return errors.New("PM forum validation unavailable")
	}
	return s.pmForumValidator(ctx, id)
}
