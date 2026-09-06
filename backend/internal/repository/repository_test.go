package repository

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ly1611240037/bandroom/backend/internal/db"
)

func TestRoomAndEquipmentCRUD(t *testing.T) {
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.Migrate(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	repo := New(database)

	capacity := int64(6)
	roomID, err := repo.CreateRoom(context.Background(), "A 房", "基础排练房", &capacity)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateRoomStatus(context.Background(), roomID, "maintenance"); err != nil {
		t.Fatal(err)
	}
	room, err := repo.GetRoom(context.Background(), roomID)
	if err != nil || room.Status != "maintenance" || room.Name != "A 房" {
		t.Fatalf("unexpected room: %+v, err=%v", room, err)
	}

	equipmentID, err := repo.CreatePublicEquipment(context.Background(), "麦克风", 4)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdatePublicEquipmentStatus(context.Background(), equipmentID, "maintenance"); err != nil {
		t.Fatal(err)
	}
	equipment, err := repo.GetPublicEquipment(context.Background(), equipmentID)
	if err != nil || equipment.Quantity != 4 || equipment.Status != "maintenance" {
		t.Fatalf("unexpected equipment: %+v, err=%v", equipment, err)
	}
}
