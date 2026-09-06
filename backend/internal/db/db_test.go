package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestMigrateCreatesSchemaAndIsRepeatable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bandroom.db")
	database, err := Open(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	if err := Migrate(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(context.Background(), database); err != nil {
		t.Fatal(err)
	}

	for _, table := range []string{"users", "membership_cards", "rooms", "public_equipment", "bookings", "audit_logs"} {
		var name string
		if err := database.QueryRow("SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&name); err != nil {
			t.Fatalf("table %s was not created: %v", table, err)
		}
	}
	var applied int
	if err := database.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&applied); err != nil {
		t.Fatal(err)
	}
	if applied != 1 {
		t.Fatalf("expected one migration, got %d", applied)
	}
}

var _ *sql.DB
