package booking

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/ly1611240037/bandroom/backend/internal/db"
)

func bookingService(t *testing.T) *Service {
	t.Helper()
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "booking.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.Migrate(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 2; i++ {
		if _, err := database.Exec(`INSERT INTO users(role, name, email, phone, password_hash, email_verified_at) VALUES ('customer', ?, ?, '13800000000', 'hash', CURRENT_TIMESTAMP)`, "顾客", "customer"+string(rune('0'+i))+"@example.com"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := database.Exec(`INSERT INTO membership_plans(plan_type, name, price_cents, included_uses) VALUES ('monthly', '月卡', 1, NULL)`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO membership_cards(user_id, plan_id, starts_at, ends_at, status, paid_amount_cents) VALUES (1, 1, '2026-09-01T00:00:00Z', '2026-10-01T00:00:00Z', 'active', 1), (2, 1, '2026-09-01T00:00:00Z', '2026-10-01T00:00:00Z', 'active', 1)`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO rooms(name, description, status) VALUES ('A 房', '', 'available'), ('B 房', '', 'available')`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO public_equipment(name, quantity, status) VALUES ('麦克风', 1, 'available')`); err != nil {
		t.Fatal(err)
	}
	return NewService(database)
}

func TestAvailabilityAndTransactionalBookingConflicts(t *testing.T) {
	service := bookingService(t)
	ctx := context.Background()
	slots, err := service.Availability(ctx, 1, "2026-09-07")
	if err != nil {
		t.Fatal(err)
	}
	if len(slots) == 0 {
		t.Fatal("expected available slots")
	}
	first, err := service.Create(ctx, 1, 1, "测试乐队", "13800000000", "2026-09-07T09:00:00+08:00", "2026-09-07T10:00:00+08:00", "第一次排练", []EquipmentRequest{{EquipmentID: 1, Quantity: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if first.OccupiedUntil != "2026-09-07T10:30:00+08:00" {
		t.Fatalf("cleaning buffer was not stored: %+v", first)
	}
	if _, err := service.Create(ctx, 2, 1, "另一支乐队", "13800000001", "2026-09-07T10:00:00+08:00", "2026-09-07T11:00:00+08:00", "", nil); err != ErrConflict {
		t.Fatalf("expected room conflict, got %v", err)
	}
	if _, err := service.Create(ctx, 2, 2, "另一支乐队", "13800000001", "2026-09-07T09:00:00+08:00", "2026-09-07T10:00:00+08:00", "", []EquipmentRequest{{EquipmentID: 1, Quantity: 1}}); err == nil {
		t.Fatal("expected equipment inventory conflict")
	}
	if _, err := service.Create(ctx, 2, 2, "另一支乐队", "13800000001", "2026-09-07T11:00:00+08:00", "2026-09-07T12:00:00+08:00", "", nil); err != nil {
		t.Fatal(err)
	}
}

func TestBookingDurationAndFutureBookingRules(t *testing.T) {
	service := bookingService(t)
	ctx := context.Background()
	if _, err := service.Create(ctx, 1, 1, "测试乐队", "13800000000", "2026-09-07T09:15:00+08:00", "2026-09-07T10:00:00+08:00", "", nil); err == nil {
		t.Fatal("quarter-hour booking should be rejected")
	}
	if _, err := service.Create(ctx, 1, 1, "测试乐队", "13800000000", "2026-09-07T09:00:00+08:00", "2026-09-07T09:30:00+08:00", "", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(ctx, 1, 2, "测试乐队", "13800000000", "2026-09-07T11:00:00+08:00", "2026-09-07T11:30:00+08:00", "", nil); err == nil {
		t.Fatal("customer should have only one future booking")
	}
}

func TestCancellationRestoresCountAndCompletesBookings(t *testing.T) {
	service := bookingService(t)
	ctx := context.Background()
	if _, err := service.db.Exec(`INSERT INTO membership_plans(plan_type, name, price_cents, included_uses) VALUES ('count', '次数卡', 1, 2)`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.db.Exec(`INSERT INTO membership_cards(user_id, plan_id, starts_at, status, remaining_uses, paid_amount_cents) VALUES (1, 2, '2026-09-01T00:00:00Z', 'active', 2, 1)`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.db.Exec(`UPDATE membership_cards SET status = 'disabled' WHERE id = 1`); err != nil {
		t.Fatal(err)
	}
	item, err := service.Create(ctx, 1, 1, "次数乐队", "13800000000", "2026-09-08T09:00:00+08:00", "2026-09-08T09:30:00+08:00", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	var remaining int
	if err := service.db.QueryRow(`SELECT remaining_uses FROM membership_cards WHERE id = 3`).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 1 {
		t.Fatalf("expected count deduction, got %d", remaining)
	}
	if err := service.CancelByCustomer(ctx, 1, item.ID, "临时有事"); err != nil {
		t.Fatal(err)
	}
	if err := service.db.QueryRow(`SELECT remaining_uses FROM membership_cards WHERE id = 3`).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 2 {
		t.Fatalf("expected count restoration, got %d", remaining)
	}
	if _, err := service.Create(ctx, 1, 1, "完成测试", "13800000000", "2026-09-08T10:00:00+08:00", "2026-09-08T10:30:00+08:00", "", nil); err != nil {
		t.Fatal(err)
	}
	count, err := service.CompleteDue(ctx, time.Date(2026, 9, 8, 3, 0, 0, 0, time.UTC))
	if err != nil || count != 1 {
		t.Fatalf("expected one completed booking, count=%d err=%v", count, err)
	}
}
