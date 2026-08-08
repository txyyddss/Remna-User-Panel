package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	maximumRestoreTables  = 500
	maximumRestoreColumns = 500
)

// PrepareRestoreSnapshot applies this binary's pending migrations to a staged
// SQLite snapshot and verifies that its resulting application schema matches a
// freshly migrated database. It must only be used on an offline staged copy.
func PrepareRestoreSnapshot(ctx context.Context, path string) (returnedErr error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("restore snapshot path is empty")
	}
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("resolve restore snapshot: %w", err)
	}
	dsn := "file:" + filepath.ToSlash(absolute) + "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(DELETE)&_pragma=synchronous(FULL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("open staged restore snapshot: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	closed := false
	defer func() {
		if !closed {
			closeErr := db.Close()
			if returnedErr == nil && closeErr != nil {
				returnedErr = fmt.Errorf("close staged restore snapshot: %w", closeErr)
			}
		}
	}()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping staged restore snapshot: %w", err)
	}
	if err := migrate(ctx, db); err != nil {
		return fmt.Errorf("migrate staged restore snapshot: %w", err)
	}
	if err := validateRestoreDatabase(ctx, db); err != nil {
		return err
	}

	// Close before syncing so DELETE-journal commits and connection cleanup are
	// fully reflected in the staged database file.
	if err := db.Close(); err != nil {
		return fmt.Errorf("close staged restore snapshot: %w", err)
	}
	closed = true
	file, err := os.OpenFile(absolute, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open staged restore snapshot for sync: %w", err)
	}
	syncErr := file.Sync()
	closeErr := file.Close()
	if syncErr != nil {
		return fmt.Errorf("sync staged restore snapshot: %w", syncErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close synced restore snapshot: %w", closeErr)
	}
	return nil
}

func validateRestoreDatabase(ctx context.Context, candidate *sql.DB) error {
	var integrity string
	if err := candidate.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&integrity); err != nil {
		return fmt.Errorf("verify migrated restore integrity: %w", err)
	}
	if integrity != "ok" {
		return fmt.Errorf("migrated restore integrity check returned %q", integrity)
	}
	foreignRows, err := candidate.QueryContext(ctx, `PRAGMA foreign_key_check`)
	if err != nil {
		return fmt.Errorf("verify migrated restore foreign keys: %w", err)
	}
	if foreignRows.Next() {
		_ = foreignRows.Close()
		return errors.New("migrated restore contains foreign-key violations")
	}
	if err := foreignRows.Err(); err != nil {
		_ = foreignRows.Close()
		return fmt.Errorf("read migrated restore foreign keys: %w", err)
	}
	if err := foreignRows.Close(); err != nil {
		return fmt.Errorf("close migrated restore foreign-key check: %w", err)
	}

	expected, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return fmt.Errorf("open expected restore schema: %w", err)
	}
	expected.SetMaxOpenConns(1)
	defer func() { _ = expected.Close() }()
	if _, err := expected.ExecContext(ctx, `PRAGMA foreign_keys=ON`); err != nil {
		return fmt.Errorf("configure expected restore schema: %w", err)
	}
	if err := migrate(ctx, expected); err != nil {
		return fmt.Errorf("build expected restore schema: %w", err)
	}
	want, err := readRestoreSchemaShape(ctx, expected)
	if err != nil {
		return fmt.Errorf("read expected restore schema: %w", err)
	}
	got, err := readRestoreSchemaShape(ctx, candidate)
	if err != nil {
		return fmt.Errorf("read staged restore schema: %w", err)
	}
	if !reflect.DeepEqual(got, want) {
		return errors.New("migrated restore schema does not match this application build")
	}
	return nil
}

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

func readRestoreTableShape(ctx context.Context, db *sql.DB, tableName string) (restoreTableShape, error) {
	table := restoreTableShape{Name: tableName, Columns: make([]restoreColumnShape, 0), ForeignKeys: make([]restoreForeignKeyShape, 0), Indexes: make([]restoreIndexShape, 0)}
	columnRows, err := db.QueryContext(ctx, `SELECT cid,name,type,"notnull",dflt_value,pk,hidden FROM pragma_table_xinfo(?) ORDER BY cid`, tableName)
	if err != nil {
		return table, err
	}
	for columnRows.Next() {
		var column restoreColumnShape
		var defaultValue sql.NullString
		if err := columnRows.Scan(&column.CID, &column.Name, &column.Type, &column.NotNull, &defaultValue, &column.PrimaryKey, &column.Hidden); err != nil {
			_ = columnRows.Close()
			return table, err
		}
		if len(table.Columns) >= maximumRestoreColumns {
			_ = columnRows.Close()
			return table, errors.New("table contains too many columns")
		}
		column.Default, column.HasDefault = defaultValue.String, defaultValue.Valid
		table.Columns = append(table.Columns, column)
	}
	if err := columnRows.Err(); err != nil {
		_ = columnRows.Close()
		return table, err
	}
	if err := columnRows.Close(); err != nil {
		return table, err
	}

	foreignRows, err := db.QueryContext(ctx, `SELECT id,seq,"table","from","to",on_update,on_delete,match FROM pragma_foreign_key_list(?) ORDER BY id,seq`, tableName)
	if err != nil {
		return table, err
	}
	for foreignRows.Next() {
		var foreignKey restoreForeignKeyShape
		if err := foreignRows.Scan(&foreignKey.ID, &foreignKey.Sequence, &foreignKey.Table, &foreignKey.From, &foreignKey.To,
			&foreignKey.OnUpdate, &foreignKey.OnDelete, &foreignKey.Match); err != nil {
			_ = foreignRows.Close()
			return table, err
		}
		table.ForeignKeys = append(table.ForeignKeys, foreignKey)
	}
	if err := foreignRows.Err(); err != nil {
		_ = foreignRows.Close()
		return table, err
	}
	if err := foreignRows.Close(); err != nil {
		return table, err
	}

	indexRows, err := db.QueryContext(ctx, `SELECT name,"unique",origin,partial FROM pragma_index_list(?) ORDER BY name`, tableName)
	if err != nil {
		return table, err
	}
	for indexRows.Next() {
		var index restoreIndexShape
		if err := indexRows.Scan(&index.Name, &index.Unique, &index.Origin, &index.Partial); err != nil {
			_ = indexRows.Close()
			return table, err
		}
		table.Indexes = append(table.Indexes, index)
	}
	if err := indexRows.Err(); err != nil {
		_ = indexRows.Close()
		return table, err
	}
	if err := indexRows.Close(); err != nil {
		return table, err
	}
	for indexPosition := range table.Indexes {
		index := &table.Indexes[indexPosition]
		var definition sql.NullString
		if err := db.QueryRowContext(ctx, `SELECT sql FROM sqlite_schema WHERE type='index' AND name=?`, index.Name).Scan(&definition); err != nil {
			return table, err
		}
		if definition.Valid {
			index.SQL = normalizeSchemaSQL(definition.String)
		}
		columnRows, err := db.QueryContext(ctx, `SELECT seqno,cid,name,"desc",coll,"key" FROM pragma_index_xinfo(?) ORDER BY seqno`, index.Name)
		if err != nil {
			return table, err
		}
		for columnRows.Next() {
			var column restoreIndexColumnShape
			var name, collate sql.NullString
			if err := columnRows.Scan(&column.Sequence, &column.CID, &name, &column.Desc, &collate, &column.Key); err != nil {
				_ = columnRows.Close()
				return table, err
			}
			column.Name, column.HasName, column.Collate = name.String, name.Valid, collate.String
			index.Columns = append(index.Columns, column)
		}
		if err := columnRows.Err(); err != nil {
			_ = columnRows.Close()
			return table, err
		}
		if err := columnRows.Close(); err != nil {
			return table, err
		}
	}
	return table, nil
}

func normalizeSchemaSQL(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
