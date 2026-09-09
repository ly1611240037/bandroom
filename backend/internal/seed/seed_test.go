package seed

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ly1611240037/bandroom/backend/internal/auth"
	"github.com/ly1611240037/bandroom/backend/internal/db"
)

func TestSeedDemoIsRepeatable(t *testing.T) {
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "seed.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.Migrate(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	if err := SeedDemo(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`UPDATE users SET password_hash = 'old-password' WHERE email = 'owner@bandroom.test'`); err != nil {
		t.Fatal(err)
	}
	if err := SeedDemo(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	if _, _, err := auth.NewService(database, "http://localhost:8080", nil).Login(context.Background(), "owner@bandroom.test", "demo123456"); err != nil {
		t.Fatalf("seed should restore demo password: %v", err)
	}

	checks := []struct {
		table string
		want  int
	}{
		{"users", 2},
		{"membership_plans", 2},
		{"rooms", 3},
		{"public_equipment", 3},
		{"fixed_equipment", 13},
	}
	for _, check := range checks {
		var got int
		if err := database.QueryRow("SELECT COUNT(*) FROM " + check.table).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != check.want {
			t.Fatalf("%s: got %d rows, want %d", check.table, got, check.want)
		}
	}
	if _, err := database.Exec(`UPDATE rooms SET description = '老板自定义资料' WHERE name = 'A 房'`); err != nil {
		t.Fatal(err)
	}
	if err := SeedDemo(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	var description string
	if err := database.QueryRow(`SELECT description FROM rooms WHERE name = 'A 房'`).Scan(&description); err != nil || description != "老板自定义资料" {
		t.Fatalf("seed must preserve custom room description: %q, %v", description, err)
	}
}
