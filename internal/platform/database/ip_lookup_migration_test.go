package database

import (
	"database/sql"
	"testing"
)

func TestIPLookupMigrationGivesOneGiftPerLiveMember(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("sqlite", "file:ip-gift-test?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	_, err = db.Exec(`CREATE TABLE users(id TEXT PRIMARY KEY); CREATE TABLE combos(id TEXT PRIMARY KEY); CREATE TABLE provider_operations(id TEXT PRIMARY KEY);
	CREATE TABLE purchases(id TEXT PRIMARY KEY,user_id TEXT,status TEXT,valid_from TEXT,valid_until TEXT);
	INSERT INTO users VALUES('live'),('expired'),('queued');
	INSERT INTO purchases VALUES('a','live','active',datetime('now','-1 day'),datetime('now','+1 day')),
	('b','live','active',datetime('now','-2 day'),datetime('now','+2 day')),
	('c','expired','active',datetime('now','-2 day'),datetime('now','-1 day')),
	('d','queued','queued',datetime('now','-1 day'),datetime('now','+1 day'));`)
	if err != nil {
		t.Fatal(err)
	}
	migration, err := migrations.ReadFile("migrations/066_ip_lookup.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(migration)); err != nil {
		t.Fatal(err)
	}
	var count, total int
	if err := db.QueryRow(`SELECT COUNT(*),SUM(total) FROM ip_lookup_allowances`).Scan(&count, &total); err != nil {
		t.Fatal(err)
	}
	if count != 1 || total != 1 {
		t.Fatalf("granted %d rows / %d checks", count, total)
	}
}
