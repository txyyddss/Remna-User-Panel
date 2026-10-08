package admin

import (
	"context"
	"errors"
	"strings"
)

func validateTurnstileKey(value string) error {
	if len(value) > 256 || strings.ContainsAny(value, " \t\r\n") {
		return errors.New("invalid Turnstile site key")
	}
	return nil
}

func (s *SettingsService) validateTurnstileSettings(ctx context.Context, key, value string) error {
	if key == "captcha.turnstile.enabled" && value == "true" {
		for _, required := range []string{"captcha.turnstile.site_key", "captcha.turnstile.secret_key"} {
			stored, err := s.Optional(ctx, required)
			if err != nil {
				return err
			}
			if stored == "" {
				return errors.New("both Turnstile keys are required before enabling CAPTCHA")
			}
		}
	}
	if key == "captcha.turnstile.site_key" && value == "" {
		enabled, err := s.Optional(ctx, "captcha.turnstile.enabled")
		if err != nil {
			return err
		}
		if enabled == "true" {
			return errors.New("disable CAPTCHA before clearing its site key")
		}
	}
	return nil
}
