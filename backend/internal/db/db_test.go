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
	if applied != 3 {
		t.Fatalf("expected three migrations, got %d", applied)
	}
}

var _ *sql.DB

func TestEveryConnectionEnforcesForeignKeys(t *testing.T) {
	ctx := context.Background()
	database, err := Open(ctx, filepath.Join(t.TempDir(), "connections.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := Migrate(ctx, database); err != nil {
		t.Fatal(err)
	}
	first, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	for _, conn := range []*sql.Conn{first, second} {
		var enabled int
		if err := conn.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&enabled); err != nil || enabled != 1 {
			t.Fatalf("foreign keys = %d, err = %v", enabled, err)
		}
		if _, err := conn.ExecContext(ctx, `INSERT INTO room_photos(room_id, url) VALUES (99999, 'test')`); err == nil {
			t.Fatal("orphan photo must be rejected on every connection")
		}
	}
}
