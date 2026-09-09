package database

import (
	"context"
	"testing"
)

func TestSquadGeocheckDefaultsAndSparseOverride(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	const uuid = "geocheck-squad"
	product, err := store.ImportSquad(ctx, uuid, "Geocheck squad")
	if err != nil || !product.GeocheckEnabled {
		t.Fatalf("unedited squad = %+v, %v", product, err)
	}
	disabled, enabled := false, true
	for _, tc := range []struct {
		name  string
		input *bool
		want  bool
		rows  int
	}{
		{"disable", &disabled, false, 1},
		{"omission preserves disabled", nil, false, 1},
		{"enable restores sparse defaults", &enabled, true, 0},
		{"omission defaults enabled", nil, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			product, err := store.SaveSquadProduct(ctx, SquadProductInput{
				RemnaSquadUUID: uuid, Name: "Geocheck squad", UpstreamPresent: true, GeocheckEnabled: tc.input,
			})
			if err != nil || product.GeocheckEnabled != tc.want {
				t.Fatalf("saved squad = %+v, %v", product, err)
			}
			actual, err := store.SquadGeocheckEnabled(ctx, uuid)
			if err != nil || actual != tc.want {
				t.Fatalf("persisted setting = %v, %v", actual, err)
			}
			var count int
			if err := store.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM squad_product_overrides WHERE remna_squad_uuid=?`, uuid).Scan(&count); err != nil || count != tc.rows {
				t.Fatalf("sparse rows = %d, %v", count, err)
			}
		})
	}
}
