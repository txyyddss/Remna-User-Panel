package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

// ComboInput is the validated catalog representation persisted by an administrator.
type ComboInput struct {
	ID                string
	Name              string
	Description       string
	PriceTXBMinor     int64
	ValidityDays      int
	TrafficLimitBytes int64
	ResetStrategy     string
	Active            bool
	SquadProductIDs   []string
}

// SquadProductInput is the local merchandising data associated with a Remnawave squad.
type SquadProductInput struct {
	ID              string
	RemnaSquadUUID  string
	Name            string
	Description     string
	PriceTXBMinor   int64
	Visible         bool
	UpstreamPresent bool
}

// ImportedSquad is the upstream-owned portion of an internal squad.
type ImportedSquad struct {
	UUID string
	Name string
}

// SaveCombo creates a catalog combo when ID is empty and otherwise updates an
// existing combo. Client-selected IDs can never create records.
func (s *Store) SaveCombo(ctx context.Context, input ComboInput) (model.Combo, error) {
	creating := input.ID == ""
	if creating {
		var err error
		input.ID, err = ids.New()
		if err != nil {
			return model.Combo{}, err
		}
	}
	now := stamp(time.Now().UTC())
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Combo{}, fmt.Errorf("begin save combo: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var result sql.Result
	if creating {
		result, err = tx.ExecContext(ctx, `INSERT INTO combos(id,name,description,price_txb_minor,validity_days,traffic_limit_bytes,reset_strategy,active,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,?,?,?)`, input.ID, input.Name, input.Description, input.PriceTXBMinor, input.ValidityDays,
			input.TrafficLimitBytes, input.ResetStrategy, boolInt(input.Active), now, now)
	} else {
		result, err = tx.ExecContext(ctx, `UPDATE combos SET name=?,description=?,price_txb_minor=?,validity_days=?,traffic_limit_bytes=?,
			reset_strategy=?,active=?,updated_at=? WHERE id=?`, input.Name, input.Description, input.PriceTXBMinor, input.ValidityDays,
			input.TrafficLimitBytes, input.ResetStrategy, boolInt(input.Active), now, input.ID)
	}
	if err != nil {
		return model.Combo{}, fmt.Errorf("save combo: %w", err)
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return model.Combo{}, fmt.Errorf("inspect saved combo: %w", rowsErr)
	} else if affected != 1 {
		return model.Combo{}, ErrNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM combo_squads WHERE combo_id=?`, input.ID); err != nil {
		return model.Combo{}, fmt.Errorf("replace combo squads: %w", err)
	}
	seen := make(map[string]struct{}, len(input.SquadProductIDs))
	for _, squadID := range input.SquadProductIDs {
		if _, duplicate := seen[squadID]; duplicate {
			continue
		}
		seen[squadID] = struct{}{}
		result, err := tx.ExecContext(ctx, `INSERT INTO combo_squads(combo_id,squad_product_id)
			SELECT ?,id FROM squad_products WHERE id=? AND upstream_present=1`, input.ID, squadID)
		if err != nil {
			return model.Combo{}, fmt.Errorf("attach combo squad: %w", err)
		}
		if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
			return model.Combo{}, fmt.Errorf("inspect combo squad: %w", rowsErr)
		} else if affected != 1 {
			return model.Combo{}, ErrNotFound
		}
	}
	if err := tx.Commit(); err != nil {
		return model.Combo{}, fmt.Errorf("commit combo: %w", err)
	}
	return s.ComboByID(ctx, input.ID, false)
}

// DeleteCombo hides a combo without invalidating historical purchases.
func (s *Store) DeleteCombo(ctx context.Context, comboID string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE combos SET active=0,updated_at=? WHERE id=?`, stamp(time.Now().UTC()), comboID)
	if err != nil {
		return fmt.Errorf("hide combo: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

// SaveSquadProduct updates merchandising for an imported upstream squad.
// New upstream identities can only enter through RefreshImportedSquads.
func (s *Store) SaveSquadProduct(ctx context.Context, input SquadProductInput) (model.SquadProduct, error) {
	if input.ID == "" {
		return model.SquadProduct{}, ErrNotFound
	}
	now := stamp(time.Now().UTC())
	result, err := s.db.ExecContext(ctx, `UPDATE squad_products SET name=?,description=?,price_txb_minor=?,visible=?,updated_at=?
		WHERE id=? AND remna_squad_uuid=?`, input.Name, input.Description, input.PriceTXBMinor, boolInt(input.Visible), now,
		input.ID, input.RemnaSquadUUID)
	if err != nil {
		return model.SquadProduct{}, fmt.Errorf("save squad product: %w", err)
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return model.SquadProduct{}, fmt.Errorf("inspect saved squad product: %w", rowsErr)
	} else if affected != 1 {
		return model.SquadProduct{}, ErrNotFound
	}
	return s.SquadProductByID(ctx, input.ID)
}

// SquadProductByRemnaUUID resolves an already imported upstream squad.
func (s *Store) SquadProductByRemnaUUID(ctx context.Context, uuid string) (model.SquadProduct, error) {
	return scanSquad(s.db.QueryRowContext(ctx, squadSelect+` WHERE remna_squad_uuid=?`, uuid))
}

// ImportSquad refreshes upstream identity while preserving local description, price, and visibility.
func (s *Store) ImportSquad(ctx context.Context, remnaSquadUUID, name string) (model.SquadProduct, error) {
	now := stamp(time.Now().UTC())
	id, err := ids.New()
	if err != nil {
		return model.SquadProduct{}, err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO squad_products(id,remna_squad_uuid,name,description,price_txb_minor,visible,upstream_present,created_at,updated_at)
		VALUES(?,?,?,'',0,0,1,?,?) ON CONFLICT(remna_squad_uuid) DO UPDATE SET name=excluded.name,upstream_present=1,updated_at=excluded.updated_at`,
		id, remnaSquadUUID, name, now, now)
	if err != nil {
		return model.SquadProduct{}, fmt.Errorf("import squad: %w", err)
	}
	var productID string
	if err := s.db.QueryRowContext(ctx, `SELECT id FROM squad_products WHERE remna_squad_uuid=?`, remnaSquadUUID).Scan(&productID); err != nil {
		return model.SquadProduct{}, err
	}
	return s.SquadProductByID(ctx, productID)
}

// MarkAllSquadsMissing clears the import-presence flag before an upstream refresh.
func (s *Store) MarkAllSquadsMissing(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `UPDATE squad_products SET upstream_present=0,updated_at=?`, stamp(time.Now().UTC()))
	if err != nil {
		return fmt.Errorf("mark squads missing: %w", err)
	}
	return nil
}

// RefreshImportedSquads atomically updates presence and names while retaining all merchandising fields.
func (s *Store) RefreshImportedSquads(ctx context.Context, squads []ImportedSquad) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	now := stamp(time.Now().UTC())
	if _, err := tx.ExecContext(ctx, `UPDATE squad_products SET upstream_present=0,updated_at=?`, now); err != nil {
		return fmt.Errorf("mark squads missing: %w", err)
	}
	for _, squad := range squads {
		id, err := ids.New()
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO squad_products(id,remna_squad_uuid,name,description,price_txb_minor,visible,upstream_present,created_at,updated_at)
			VALUES(?,?,?,'',0,0,1,?,?) ON CONFLICT(remna_squad_uuid) DO UPDATE SET name=excluded.name,upstream_present=1,updated_at=excluded.updated_at`,
			id, squad.UUID, squad.Name, now, now)
		if err != nil {
			return fmt.Errorf("refresh imported squad: %w", err)
		}
	}
	return tx.Commit()
}

// SquadProductByID loads one optional product.
func (s *Store) SquadProductByID(ctx context.Context, id string) (model.SquadProduct, error) {
	return scanSquad(s.db.QueryRowContext(ctx, squadSelect+` WHERE id=?`, id))
}

// ListSquadProducts loads visible products for users or every product for administrators.
func (s *Store) ListSquadProducts(ctx context.Context, visibleOnly bool) ([]model.SquadProduct, error) {
	query := squadSelect
	if visibleOnly {
		query += ` WHERE visible=1 AND upstream_present=1`
	}
	query += ` ORDER BY name,id`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list squad products: %w", err)
	}
	defer rows.Close()
	products := make([]model.SquadProduct, 0)
	for rows.Next() {
		product, err := scanSquad(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, rows.Err()
}

const squadSelect = `SELECT id,remna_squad_uuid,name,description,price_txb_minor,visible,upstream_present,created_at,updated_at FROM squad_products`

func scanSquad(row rowScanner) (model.SquadProduct, error) {
	var product model.SquadProduct
	var visible, present int
	var created, updated string
	if err := row.Scan(&product.ID, &product.RemnaSquadUUID, &product.Name, &product.Description, &product.PriceTXBMinor,
		&visible, &present, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.SquadProduct{}, ErrNotFound
		}
		return model.SquadProduct{}, fmt.Errorf("scan squad product: %w", err)
	}
	product.Visible = visible == 1
	product.UpstreamPresent = present == 1
	product.Price = model.TXBMoney(product.PriceTXBMinor)
	var err error
	product.CreatedAt, err = parseStamp(created)
	if err != nil {
		return model.SquadProduct{}, err
	}
	product.UpdatedAt, err = parseStamp(updated)
	return product, err
}

// ComboByID returns one combo with its included squad products.
func (s *Store) ComboByID(ctx context.Context, id string, activeOnly bool) (model.Combo, error) {
	query := comboSelect + ` WHERE id=?`
	if activeOnly {
		query += ` AND active=1`
	}
	combo, err := scanCombo(s.db.QueryRowContext(ctx, query, id))
	if err != nil {
		return model.Combo{}, err
	}
	combo.IncludedSquads, err = s.comboSquads(ctx, combo.ID)
	return combo, err
}

// ListCombos returns the catalog in stable display order.
func (s *Store) ListCombos(ctx context.Context, activeOnly bool) ([]model.Combo, error) {
	query := comboSelect
	if activeOnly {
		query += ` WHERE active=1 AND NOT EXISTS (
			SELECT 1 FROM combo_squads cs
			JOIN squad_products sp ON sp.id=cs.squad_product_id
			WHERE cs.combo_id=combos.id AND sp.upstream_present=0
		)`
	}
	query += ` ORDER BY price_txb_minor,name,id`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list combos: %w", err)
	}
	defer rows.Close()
	combos := make([]model.Combo, 0)
	for rows.Next() {
		combo, err := scanCombo(rows)
		if err != nil {
			return nil, err
		}
		combos = append(combos, combo)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate combos: %w", err)
	}
	for index := range combos {
		combos[index].IncludedSquads, err = s.comboSquads(ctx, combos[index].ID)
		if err != nil {
			return nil, err
		}
	}
	return combos, nil
}

const comboSelect = `SELECT id,name,description,price_txb_minor,validity_days,traffic_limit_bytes,reset_strategy,active,created_at,updated_at FROM combos`

func scanCombo(row rowScanner) (model.Combo, error) {
	var combo model.Combo
	var active int
	var created, updated string
	if err := row.Scan(&combo.ID, &combo.Name, &combo.Description, &combo.PriceTXBMinor, &combo.ValidityDays,
		&combo.TrafficLimitBytes, &combo.ResetStrategy, &active, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Combo{}, ErrNotFound
		}
		return model.Combo{}, fmt.Errorf("scan combo: %w", err)
	}
	combo.Active = active == 1
	combo.Price = model.TXBMoney(combo.PriceTXBMinor)
	combo.TrafficLimit = fmt.Sprintf("%d", combo.TrafficLimitBytes)
	var err error
	combo.CreatedAt, err = parseStamp(created)
	if err != nil {
		return model.Combo{}, err
	}
	combo.UpdatedAt, err = parseStamp(updated)
	return combo, err
}

func (s *Store) comboSquads(ctx context.Context, comboID string) ([]model.SquadProduct, error) {
	rows, err := s.db.QueryContext(ctx, squadSelect+` JOIN combo_squads cs ON cs.squad_product_id=squad_products.id WHERE cs.combo_id=? ORDER BY name,id`, comboID)
	if err != nil {
		return nil, fmt.Errorf("list combo squads: %w", err)
	}
	defer rows.Close()
	products := make([]model.SquadProduct, 0)
	for rows.Next() {
		product, err := scanSquad(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, rows.Err()
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok || value == "" {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
