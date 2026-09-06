package rooms

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ly1611240037/bandroom/backend/internal/db"
)

func roomService(t *testing.T) *Service {
	t.Helper()
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "rooms.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.Migrate(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO users(role, name, email, phone, password_hash) VALUES ('customer', '顾客', 'customer@example.com', '13800000000', 'hash')`); err != nil {
		t.Fatal(err)
	}
	return NewService(database)
}

func TestRoomEquipmentAndScheduleManagement(t *testing.T) {
	service := roomService(t)
	ctx := context.Background()
	capacity := int64(6)
	room, err := service.CreateRoom(ctx, "排练房 A", "标准房", &capacity)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.AddPhoto(ctx, room.ID, "https://example.test/room-a.jpg", 0); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddFixedEquipment(ctx, room.ID, "架子鼓", "固定设备", 1); err != nil {
		t.Fatal(err)
	}
	room, err = service.GetRoom(ctx, room.ID)
	if err != nil || len(room.Photos) != 1 || len(room.FixedEquipment) != 1 {
		t.Fatalf("room details incomplete: %+v, %v", room, err)
	}
	if err := service.UpdateRoom(ctx, room.ID, room.Name, room.Description, room.Capacity, "maintenance"); err != nil {
		t.Fatal(err)
	}
	if err := service.CanBookRoom(ctx, room.ID); err != ErrUnavailable {
		t.Fatalf("maintenance room should be unavailable: %v", err)
	}
	item, err := service.CreateEquipment(ctx, "麦克风", "公共麦克风", 8, 0)
	if err != nil || item.Quantity != 8 {
		t.Fatalf("equipment create failed: %+v, %v", item, err)
	}
	if err := service.UpdateEquipment(ctx, item.ID, item.Name, item.Description, 7, 0, "maintenance"); err != nil {
		t.Fatal(err)
	}
	if err := service.CanSelectEquipment(ctx, item.ID, 1); err == nil {
		t.Fatal("maintenance equipment should not be selectable")
	}
	hours, err := service.ListBusinessHours(ctx)
	if err != nil || len(hours) != 7 || hours[0].OpensAt != "08:00" || hours[0].ClosesAt != "23:00" {
		t.Fatalf("unexpected business hours: %+v, %v", hours, err)
	}
	buffer, err := service.GetCleaningBuffer(ctx)
	if err != nil || buffer != 30 {
		t.Fatalf("unexpected cleaning buffer: %d, %v", buffer, err)
	}
	if err := service.SetBusinessHour(ctx, BusinessHour{Weekday: 0, OpensAt: "09:00", ClosesAt: "22:00", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	issue, err := service.CreateIssueReport(ctx, 1, &room.ID, &item.ID, "麦克风没有声音")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.UpdateIssueReport(ctx, issue.ID, "in_progress", "maintenance"); err != nil {
		t.Fatal(err)
	}
	updated, err := service.GetEquipment(ctx, item.ID)
	if err != nil || updated.Status != "maintenance" {
		t.Fatalf("issue should update equipment status: %+v, %v", updated, err)
	}
}
