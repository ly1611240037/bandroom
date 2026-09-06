package admin

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ly1611240037/bandroom/backend/internal/db"
)

func TestStatsAndAudits(t *testing.T) {
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "admin.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.Migrate(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	service := NewService(database)
	stats, err := service.Stats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalBookings != 0 || stats.ActiveCards != 0 {
		t.Fatalf("unexpected empty stats: %+v", stats)
	}
	audits, err := service.Audits(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(audits) != 0 {
		t.Fatalf("expected empty audits, got %d", len(audits))
	}
}
