package backup

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

func (s *Service) mainDatabasePath(ctx context.Context) (string, error) {
	rows, err := s.db.QueryContext(ctx, `PRAGMA database_list`)
	if err != nil {
		return "", fmt.Errorf("locate live database: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var sequence int
		var name, path string
		if err := rows.Scan(&sequence, &name, &path); err != nil {
			return "", fmt.Errorf("scan database location: %w", err)
		}
		if name == "main" {
			if strings.TrimSpace(path) == "" {
				return "", errors.New("live database does not have a filesystem path")
			}
			absolute, err := filepath.Abs(path)
			if err != nil {
				return "", fmt.Errorf("resolve live database path: %w", err)
			}
			return filepath.Clean(absolute), nil
		}
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("iterate database locations: %w", err)
	}
	return "", errors.New("main SQLite database was not found")
}

