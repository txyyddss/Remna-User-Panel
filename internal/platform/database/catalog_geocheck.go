package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// SquadGeocheckEnabled reads the sparse display override; unedited squads default on.
func (s *Store) SquadGeocheckEnabled(ctx context.Context, uuid string) (bool, error) {
	var disabled bool
	err := s.db.QueryRowContext(ctx, `SELECT geocheck_disabled FROM squad_product_overrides WHERE remna_squad_uuid=?`, uuid).Scan(&disabled)
	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("read squad geocheck setting: %w", err)
	}
	return !disabled, nil
}
