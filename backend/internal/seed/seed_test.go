package seed

import (
	"context"
	"path/filepath"
	"testing"

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
	if err := SeedDemo(context.Background(), database); err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		table string
		want  int
	}{
		{"users", 2},
		{"membership_plans", 2},
		{"rooms", 3},
		{"public_equipment", 3},
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
}
