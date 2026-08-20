package database

import (
	"database/sql"
	"log/slog"
)

// SetLogger enables structured persistence lifecycle logs.
func (s *Store) SetLogger(logger *slog.Logger) { s.logger = logger }

// DB exposes the connection pool for health checks and online backup.
func (s *Store) DB() *sql.DB { return s.db }
