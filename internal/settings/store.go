// Package settings reads and writes runtime app settings stored in PostgreSQL.
package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store wraps app_settings access.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore creates a settings store.
func NewStore(pool *pgxpool.Pool) Store {
	return Store{pool: pool}
}

// Get returns a raw JSON setting and whether it exists.
func (s Store) Get(ctx context.Context, key string) (json.RawMessage, bool, error) {
	if s.pool == nil {
		return nil, false, nil
	}
	var raw json.RawMessage
	if err := s.pool.QueryRow(ctx, "SELECT value FROM app_settings WHERE key=$1", key).Scan(&raw); err != nil {
		if err == pgx.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, err
	}
	return raw, true, nil
}

// String returns a string setting or fallback.
func (s Store) String(ctx context.Context, key string, fallback string) string {
	raw, ok, err := s.Get(ctx, key)
	if err != nil || !ok {
		return fallback
	}
	var value string
	if json.Unmarshal(raw, &value) == nil {
		return strings.TrimSpace(value)
	}
	return strings.TrimSpace(string(raw))
}

// Bool returns a bool setting or fallback.
func (s Store) Bool(ctx context.Context, key string, fallback bool) bool {
	raw, ok, err := s.Get(ctx, key)
	if err != nil || !ok {
		return fallback
	}
	var value bool
	if json.Unmarshal(raw, &value) == nil {
		return value
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		switch strings.ToLower(strings.TrimSpace(text)) {
		case "1", "true", "yes", "y", "on":
			return true
		case "0", "false", "no", "n", "off":
			return false
		}
	}
	return fallback
}

// Int returns an int setting or fallback.
func (s Store) Int(ctx context.Context, key string, fallback int) int {
	raw, ok, err := s.Get(ctx, key)
	if err != nil || !ok {
		return fallback
	}
	var number float64
	if json.Unmarshal(raw, &number) == nil {
		return int(number)
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		if value, err := strconv.Atoi(strings.TrimSpace(text)); err == nil {
			return value
		}
	}
	return fallback
}

// Float returns a float64 setting or fallback.
func (s Store) Float(ctx context.Context, key string, fallback float64) float64 {
	raw, ok, err := s.Get(ctx, key)
	if err != nil || !ok {
		return fallback
	}
	var number float64
	if json.Unmarshal(raw, &number) == nil {
		return number
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		if value, err := strconv.ParseFloat(strings.TrimSpace(text), 64); err == nil {
			return value
		}
	}
	return fallback
}

// Upsert writes a JSON setting.
func (s Store) Upsert(ctx context.Context, key string, value any) error {
	if s.pool == nil {
		return fmt.Errorf("settings store is not configured")
	}
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
INSERT INTO app_settings (key, value, updated_at)
VALUES ($1, $2, $3)
ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value, updated_at=EXCLUDED.updated_at`,
		key, body, time.Now())
	return err
}

// Delete removes a setting override.
func (s Store) Delete(ctx context.Context, key string) error {
	if s.pool == nil {
		return fmt.Errorf("settings store is not configured")
	}
	_, err := s.pool.Exec(ctx, "DELETE FROM app_settings WHERE key=$1", key)
	return err
}
