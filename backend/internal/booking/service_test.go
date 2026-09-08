package booking

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
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
	service := NewService(database)
	service.now = func() time.Time { return time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC) }
	return service
}

func TestAvailabilityRespectsStatusAndCleaningBuffer(t *testing.T) {
	s := bookingService(t)
	ctx := context.Background()
	_, err := s.Create(ctx, 1, 1, "乐队", "13800000000", "2026-09-07T09:00:00+08:00", "2026-09-07T10:00:00+08:00", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	slots, err := s.Availability(ctx, 1, "2026-09-07")
	if err != nil {
		t.Fatal(err)
	}
	foundBoundary := false
	for _, slot := range slots {
		start, _ := time.Parse(time.RFC3339, slot.StartsAt)
		end, _ := time.Parse(time.RFC3339, slot.EndsAt)
		busyStart, _ := time.Parse(time.RFC3339, "2026-09-07T09:00:00+08:00")
		busyEnd := busyStart.Add(90 * time.Minute)
		if start.Before(busyEnd) && end.Add(30*time.Minute).After(busyStart) {
			t.Fatalf("overlapping slot: %+v", slot)
		}
		if slot.StartsAt == "2026-09-07T10:30:00+08:00" {
			foundBoundary = true
		}
	}
	if !foundBoundary {
		t.Fatal("cleaning end boundary should be bookable")
	}
	if _, err := s.db.Exec(`UPDATE rooms SET status = 'maintenance' WHERE id = 1`); err != nil {
		t.Fatal(err)
	}
	slots, err = s.Availability(ctx, 1, "2026-09-07")
	if err != nil || len(slots) != 0 {
		t.Fatalf("maintenance slots = %v, err = %v", slots, err)
	}
}

func TestBookingNormalizesTimezoneAndSearchesVenueDate(t *testing.T) {
	s := bookingService(t)
	ctx := context.Background()
	if _, _, err := s.validateInterval("2026-09-07T08:00:00+05:45", "2026-09-07T09:00:00+05:45"); err == nil {
		t.Fatal("half-hour alignment must be checked in venue timezone")
	}
	// 01:00 UTC is 09:00 at the venue, inside its opening hours.
	item, err := s.Create(ctx, 1, 1, "UTC 乐队", "13800000000", "2026-09-07T01:00:00Z", "2026-09-07T02:00:00Z", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if item.StartsAt != "2026-09-07T09:00:00+08:00" {
		t.Fatalf("unexpected normalized time %s", item.StartsAt)
	}
	// Owner overrides may include early morning bookings; date filters use venue time.
	if _, err := s.db.Exec(`UPDATE bookings SET starts_at = '2026-09-07T00:30:00+08:00' WHERE id = ?`, item.ID); err != nil {
		t.Fatal(err)
	}
	items, err := s.Search(ctx, "2026-09-07", "", "")
	if err != nil || len(items) != 1 {
		t.Fatalf("venue date results = %v, err = %v", items, err)
	}
}

func TestConcurrentBookingsCannotDoubleBook(t *testing.T) {
	s := bookingService(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for userID := int64(1); userID <= 2; userID++ {
		wg.Go(func() {
			<-start
			_, err := s.Create(ctx, userID, 1, "并发乐队", "13800000000", "2026-09-07T09:00:00+08:00", "2026-09-07T10:00:00+08:00", "", nil)
			results <- err
		})
	}
	close(start)
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, ErrConflict) {
			conflict++
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("success=%d conflict=%d", success, conflict)
	}
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
