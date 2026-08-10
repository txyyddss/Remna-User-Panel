package database

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

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
