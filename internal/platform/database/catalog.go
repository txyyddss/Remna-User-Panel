package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

// ComboInput is the validated catalog representation persisted by an administrator.
// SquadProductIDs contains Remnawave internal-squad UUIDs; the historical field
// name remains at the service boundary for source compatibility.
type ComboInput struct {
	ID                      string
	Name                    string
	Description             string
	PriceTXBMinor           int64
	ValidityDays            int
	TrafficLimitBytes       int64
	ResetStrategy           string
	Active                  bool
	SquadProductIDs         []string
	RolloverMinRemainingBPS int
	RolloverMaxTXBMinor     int64
}

// SquadProductInput is the local merchandising override associated with a
// Remnawave-owned internal squad.
type SquadProductInput struct {
	ID              string
	RemnaSquadUUID  string
	Name            string
	Description     string
	PriceTXBMinor   int64
	Visible         bool
	UpstreamPresent bool
}

// ImportedSquad is retained as a compatibility DTO. Upstream squads are no
// longer persisted by refresh operations.
type ImportedSquad struct {
	UUID string
	Name string
}

// SaveCombo creates or updates one stable live combo and schedules affected
// active users for upstream re-synchronization after an edit.
func (s *Store) SaveCombo(ctx context.Context, input ComboInput) (model.Combo, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	creating := input.ID == ""
	if creating {
		var err error
		input.ID, err = ids.New()
		if err != nil {
			return model.Combo{}, err
		}
	}
	squadUUIDs := uniqueSorted(input.SquadProductIDs)
	encodedSquads, err := json.Marshal(squadUUIDs)
	if err != nil {
		return model.Combo{}, fmt.Errorf("encode combo squad UUIDs: %w", err)
	}
	now := time.Now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Combo{}, fmt.Errorf("begin save combo: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var result sql.Result
	if creating {
		result, err = tx.ExecContext(ctx, `INSERT INTO combos(id,name,description,price_txb_minor,validity_days,traffic_limit_bytes,reset_strategy,active,rollover_min_remaining_bps,rollover_max_txb_minor,included_squad_uuids,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, input.ID, input.Name, input.Description, input.PriceTXBMinor, input.ValidityDays,
			input.TrafficLimitBytes, input.ResetStrategy, boolInt(input.Active), input.RolloverMinRemainingBPS, input.RolloverMaxTXBMinor,
			string(encodedSquads), stamp(now), stamp(now))
	} else {
		result, err = tx.ExecContext(ctx, `UPDATE combos SET name=?,description=?,price_txb_minor=?,validity_days=?,traffic_limit_bytes=?,reset_strategy=?,active=?,rollover_min_remaining_bps=?,rollover_max_txb_minor=?,included_squad_uuids=?,updated_at=? WHERE id=?`,
			input.Name, input.Description, input.PriceTXBMinor, input.ValidityDays, input.TrafficLimitBytes, input.ResetStrategy,
			boolInt(input.Active), input.RolloverMinRemainingBPS, input.RolloverMaxTXBMinor, string(encodedSquads), stamp(now), input.ID)
	}
	if err != nil {
		return model.Combo{}, fmt.Errorf("save combo: %w", err)
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return model.Combo{}, fmt.Errorf("inspect saved combo: %w", rowsErr)
	} else if affected != 1 {
		return model.Combo{}, ErrNotFound
	}
	if !creating {
		rows, queryErr := tx.QueryContext(ctx, `SELECT DISTINCT user_id FROM purchases WHERE combo_id=? AND status IN ('activating','active','queued')`, input.ID)
		if queryErr != nil {
			return model.Combo{}, fmt.Errorf("list combo users for resync: %w", queryErr)
		}
		userIDs := make([]string, 0)
		for rows.Next() {
			var userID string
			if scanErr := rows.Scan(&userID); scanErr != nil {
				_ = rows.Close()
				return model.Combo{}, fmt.Errorf("scan combo user for resync: %w", scanErr)
			}
			userIDs = append(userIDs, userID)
		}
		if rowsErr := rows.Err(); rowsErr != nil {
			_ = rows.Close()
			return model.Combo{}, fmt.Errorf("iterate combo users for resync: %w", rowsErr)
		}
		_ = rows.Close()
		for _, userID := range userIDs {
			payload, marshalErr := json.Marshal(map[string]string{"userId": userID})
			if marshalErr != nil {
				return model.Combo{}, marshalErr
			}
			if enqueueErr := insertOutboxTx(ctx, tx, "remna_sync_user", string(payload), now, now); enqueueErr != nil {
				return model.Combo{}, enqueueErr
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return model.Combo{}, fmt.Errorf("commit combo: %w", err)
	}
	return s.ComboByID(ctx, input.ID, false)
}

// DeleteCombo hides a combo without invalidating purchases that reference it.
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

// SaveSquadProduct upserts only non-default local merchandising. Restoring all
// defaults removes the row so unedited upstream squads consume no local space.
func (s *Store) SaveSquadProduct(ctx context.Context, input SquadProductInput) (model.SquadProduct, error) {
	uuid := strings.TrimSpace(input.RemnaSquadUUID)
	if uuid == "" {
		return model.SquadProduct{}, ErrNotFound
	}
	now := time.Now().UTC()
	if strings.TrimSpace(input.Description) == "" && input.PriceTXBMinor == 0 && !input.Visible {
		if _, err := s.db.ExecContext(ctx, `DELETE FROM squad_product_overrides WHERE remna_squad_uuid=?`, uuid); err != nil {
			return model.SquadProduct{}, fmt.Errorf("remove default squad override: %w", err)
		}
		return virtualSquad(input, now), nil
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO squad_product_overrides(remna_squad_uuid,description,price_txb_minor,visible,created_at,updated_at)
		VALUES(?,?,?,?,?,?) ON CONFLICT(remna_squad_uuid) DO UPDATE SET description=excluded.description,price_txb_minor=excluded.price_txb_minor,visible=excluded.visible,updated_at=excluded.updated_at`,
		uuid, strings.TrimSpace(input.Description), input.PriceTXBMinor, boolInt(input.Visible), stamp(now), stamp(now))
	if err != nil {
		return model.SquadProduct{}, fmt.Errorf("save squad override: %w", err)
	}
	product, err := s.SquadProductByRemnaUUID(ctx, uuid)
	if err != nil {
		return model.SquadProduct{}, err
	}
	product.Name = input.Name
	return product, nil
}

func virtualSquad(input SquadProductInput, now time.Time) model.SquadProduct {
	return model.SquadProduct{ID: input.RemnaSquadUUID, RemnaSquadUUID: input.RemnaSquadUUID, Name: input.Name,
		Description: strings.TrimSpace(input.Description), PriceTXBMinor: input.PriceTXBMinor, Price: model.TXBMoney(input.PriceTXBMinor),
		Visible: input.Visible, UpstreamPresent: true, CreatedAt: now, UpdatedAt: now}
}

// SquadProductByRemnaUUID resolves a persisted local override.
func (s *Store) SquadProductByRemnaUUID(ctx context.Context, uuid string) (model.SquadProduct, error) {
	return scanSquad(s.db.QueryRowContext(ctx, squadSelect+` WHERE remna_squad_uuid=?`, uuid))
}

// SquadProductByID treats the public product ID as the upstream UUID.
func (s *Store) SquadProductByID(ctx context.Context, id string) (model.SquadProduct, error) {
	return s.SquadProductByRemnaUUID(ctx, id)
}

// ImportSquad returns the live upstream identity overlaid with local data. It
// deliberately performs no persistence.
func (s *Store) ImportSquad(ctx context.Context, remnaSquadUUID, name string) (model.SquadProduct, error) {
	product, err := s.SquadProductByRemnaUUID(ctx, remnaSquadUUID)
	if errors.Is(err, ErrNotFound) {
		return virtualSquad(SquadProductInput{RemnaSquadUUID: remnaSquadUUID, Name: name, UpstreamPresent: true}, time.Now().UTC()), nil
	}
	if err != nil {
		return model.SquadProduct{}, err
	}
	product.Name = name
	return product, nil
}

// MarkAllSquadsMissing is retained for compatibility and intentionally does
// nothing because presence is now read directly from Remnawave.
func (s *Store) MarkAllSquadsMissing(context.Context) error { return nil }

// RefreshImportedSquads is retained for compatibility and intentionally does
// not store upstream-owned identities.
func (s *Store) RefreshImportedSquads(context.Context, []ImportedSquad) error { return nil }

// ListSquadProducts loads only local overrides.
func (s *Store) ListSquadProducts(ctx context.Context, visibleOnly bool) ([]model.SquadProduct, error) {
	query := squadSelect
	if visibleOnly {
		query += ` WHERE visible=1`
	}
	query += ` ORDER BY remna_squad_uuid`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list squad overrides: %w", err)
	}
	defer func() { _ = rows.Close() }()
	products := make([]model.SquadProduct, 0)
	for rows.Next() {
		product, scanErr := scanSquad(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		products = append(products, product)
	}
	return products, rows.Err()
}

const squadSelect = `SELECT remna_squad_uuid,description,price_txb_minor,visible,created_at,updated_at FROM squad_product_overrides`

func scanSquad(row rowScanner) (model.SquadProduct, error) {
	var product model.SquadProduct
	var visible int
	var created, updated string
	if err := row.Scan(&product.RemnaSquadUUID, &product.Description, &product.PriceTXBMinor, &visible, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.SquadProduct{}, ErrNotFound
		}
		return model.SquadProduct{}, fmt.Errorf("scan squad override: %w", err)
	}
	product.ID = product.RemnaSquadUUID
	product.Visible = visible == 1
	product.UpstreamPresent = true
	product.Price = model.TXBMoney(product.PriceTXBMinor)
	var err error
	product.CreatedAt, err = parseStamp(created)
	if err != nil {
		return model.SquadProduct{}, err
	}
	product.UpdatedAt, err = parseStamp(updated)
	return product, err
}

// ComboByID returns one live combo with UUID-only included-squad placeholders.
func (s *Store) ComboByID(ctx context.Context, id string, activeOnly bool) (model.Combo, error) {
	query := comboSelect + ` WHERE id=?`
	if activeOnly {
		query += ` AND active=1`
	}
	return scanCombo(s.db.QueryRowContext(ctx, query, id))
}

// ListCombos returns the catalog in stable display order.
func (s *Store) ListCombos(ctx context.Context, activeOnly bool) ([]model.Combo, error) {
	query := comboSelect
	if activeOnly {
		query += ` WHERE active=1`
	}
	query += ` ORDER BY price_txb_minor,name,id`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list combos: %w", err)
	}
	defer func() { _ = rows.Close() }()
	combos := make([]model.Combo, 0)
	for rows.Next() {
		combo, scanErr := scanCombo(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		combos = append(combos, combo)
	}
	return combos, rows.Err()
}

const comboSelect = `SELECT id,name,description,price_txb_minor,validity_days,traffic_limit_bytes,reset_strategy,active,rollover_min_remaining_bps,rollover_max_txb_minor,included_squad_uuids,created_at,updated_at FROM combos`

func scanCombo(row rowScanner) (model.Combo, error) {
	var combo model.Combo
	var active int
	var encodedSquads, created, updated string
	if err := row.Scan(&combo.ID, &combo.Name, &combo.Description, &combo.PriceTXBMinor, &combo.ValidityDays,
		&combo.TrafficLimitBytes, &combo.ResetStrategy, &active, &combo.RolloverMinRemainingBPS, &combo.RolloverMaxTXBMinor,
		&encodedSquads, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Combo{}, ErrNotFound
		}
		return model.Combo{}, fmt.Errorf("scan combo: %w", err)
	}
	var squadUUIDs []string
	if err := json.Unmarshal([]byte(encodedSquads), &squadUUIDs); err != nil {
		return model.Combo{}, fmt.Errorf("decode combo squad UUIDs: %w", err)
	}
	combo.IncludedSquads = make([]model.SquadProduct, 0, len(squadUUIDs))
	for _, uuid := range uniqueSorted(squadUUIDs) {
		combo.IncludedSquads = append(combo.IncludedSquads, model.SquadProduct{ID: uuid, RemnaSquadUUID: uuid, UpstreamPresent: true})
	}
	combo.Active = active == 1
	combo.Price = model.TXBMoney(combo.PriceTXBMinor)
	combo.RolloverMax = model.TXBMoney(combo.RolloverMaxTXBMinor)
	combo.TrafficLimit = fmt.Sprintf("%d", combo.TrafficLimitBytes)
	var err error
	combo.CreatedAt, err = parseStamp(created)
	if err != nil {
		return model.Combo{}, err
	}
	combo.UpdatedAt, err = parseStamp(updated)
	return combo, err
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if _, ok := seen[value]; ok || value == "" {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
