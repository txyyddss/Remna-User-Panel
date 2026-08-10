package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type restoreSchemaShape struct {
	Tables  []restoreTableShape
	Objects []restoreObjectShape
}

type restoreTableShape struct {
	Name        string
	Columns     []restoreColumnShape
	ForeignKeys []restoreForeignKeyShape
	Indexes     []restoreIndexShape
}

type restoreColumnShape struct {
	CID        int
	Name       string
	Type       string
	NotNull    int
	Default    string
	HasDefault bool
	PrimaryKey int
	Hidden     int
}

type restoreForeignKeyShape struct {
	ID       int
	Sequence int
	Table    string
	From     string
	To       string
	OnUpdate string
	OnDelete string
	Match    string
}

type restoreIndexShape struct {
	Name    string
	Unique  int
	Origin  string
	Partial int
	SQL     string
	Columns []restoreIndexColumnShape
}

type restoreIndexColumnShape struct {
	Sequence int
	CID      int
	Name     string
	HasName  bool
	Desc     int
	Collate  string
	Key      int
}

type restoreObjectShape struct {
	Type  string
	Name  string
	Table string
	SQL   string
}

func readRestoreSchemaShape(ctx context.Context, db *sql.DB) (restoreSchemaShape, error) {
	rows, err := db.QueryContext(ctx, `SELECT name FROM sqlite_schema WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return restoreSchemaShape{}, err
	}
	tableNames := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			_ = rows.Close()
			return restoreSchemaShape{}, err
		}
		if len(tableNames) >= maximumRestoreTables {
			_ = rows.Close()
			return restoreSchemaShape{}, errors.New("restore schema contains too many tables")
		}
		tableNames = append(tableNames, name)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return restoreSchemaShape{}, err
	}
	if err := rows.Close(); err != nil {
		return restoreSchemaShape{}, err
	}

	shape := restoreSchemaShape{Tables: make([]restoreTableShape, 0, len(tableNames)), Objects: make([]restoreObjectShape, 0)}
	for _, name := range tableNames {
		table, err := readRestoreTableShape(ctx, db, name)
		if err != nil {
			return restoreSchemaShape{}, fmt.Errorf("inspect table %q: %w", name, err)
		}
		shape.Tables = append(shape.Tables, table)
	}
	objectRows, err := db.QueryContext(ctx, `SELECT type,name,tbl_name,sql FROM sqlite_schema WHERE type IN ('trigger','view') ORDER BY type,name`)
	if err != nil {
		return restoreSchemaShape{}, err
	}
	defer func() { _ = objectRows.Close() }()
	for objectRows.Next() {
		var object restoreObjectShape
		if err := objectRows.Scan(&object.Type, &object.Name, &object.Table, &object.SQL); err != nil {
			return restoreSchemaShape{}, err
		}
		object.SQL = normalizeSchemaSQL(object.SQL)
		shape.Objects = append(shape.Objects, object)
	}
	return shape, objectRows.Err()
}
