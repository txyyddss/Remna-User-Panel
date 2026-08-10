package remnawave

import (
	"errors"
	"strings"
)

func requireUserID(user *User, expected int64) error {
	if user == nil || user.ID != expected {
		return errors.New("Remnawave user response identity does not match request")
	}
	return nil
}

func requireUsername(user *User, expected string) error {
	if user == nil || user.Username != expected {
		return errors.New("Remnawave user response identity does not match request")
	}
	return nil
}

func requireCreatedUserIdentity(user *User, input CreateUserRequest) error {
	if err := requireUsername(user, input.Username); err != nil {
		return err
	}
	if user.TelegramID == nil || *user.TelegramID != input.TelegramID {
		return errors.New("Remnawave user response identity does not match request")
	}
	return nil
}

func requireUpdatedUserIdentity(user *User, input UpdateUserRequest) error {
	if input.ID > 0 {
		return requireUserID(user, input.ID)
	}
	return requireUsername(user, input.Username)
}

func requireResolvedUserIdentity(user *UserSelector, requested UserSelector) error {
	if user == nil || user.ID <= 0 || strings.TrimSpace(user.ShortUUID) == "" || strings.TrimSpace(user.Username) == "" {
		return errors.New("Remnawave resolved user identity is invalid")
	}
	if requested.ID > 0 && user.ID != requested.ID ||
		requested.ShortUUID != "" && user.ShortUUID != requested.ShortUUID ||
		requested.Username != "" && user.Username != requested.Username {
		return errors.New("Remnawave resolved user identity does not match request")
	}
	return nil
}
