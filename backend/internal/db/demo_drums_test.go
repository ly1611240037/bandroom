package db

import (
	"context"
	"path/filepath"
	"testing"
)

func TestLegacyDemoDrumsRepair(t *testing.T) {
	ctx := context.Background()
	database, err := Open(ctx, filepath.Join(t.TempDir(), "legacy.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := Migrate(ctx, database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`DELETE FROM schema_migrations WHERE version = 4; INSERT INTO rooms(name) VALUES ('B 房'), ('自定义房')`); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 13; i++ {
		if _, err := database.Exec(`INSERT INTO fixed_equipment(room_id,name,description,quantity) VALUES (1,'架子鼓','房间固定架子鼓',1)`); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := database.Exec(`INSERT INTO fixed_equipment(room_id,name,description,quantity,status) VALUES (1,'架子鼓','房间固定架子鼓',2,'available'),(1,'架子鼓','定制配置',1,'available'),(1,'架子鼓','房间固定架子鼓',1,'maintenance'),(2,'架子鼓','房间固定架子鼓',1,'available')`); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := Migrate(ctx, database); err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM fixed_equipment`).Scan(&count); err != nil || count != 5 {
		t.Fatalf("repair should retain one legacy row and four exceptions: %d, %v", count, err)
	}
	var description string
	if err := database.QueryRow(`SELECT description FROM fixed_equipment WHERE id=1`).Scan(&description); err != nil {
		t.Fatal(err)
	}
	if description != "一套五鼓配置：底鼓 1、军鼓 1、悬挂通鼓 2、落地通鼓 1；另含踩镲 1 对、吊镲 1、叮叮镲 1、支架与鼓凳，鼓棒自备" {
		t.Fatalf("missing kit components: %s", description)
	}
}
