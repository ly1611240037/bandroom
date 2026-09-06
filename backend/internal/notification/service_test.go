package notification

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/ly1611240037/bandroom/backend/internal/db"
)

type captureSender struct {
	emails []Email
	err    error
}

func (s *captureSender) Send(_ context.Context, email Email) error {
	s.emails = append(s.emails, email)
	return s.err
}

func notificationService(t *testing.T, sender Sender) *Service {
	t.Helper()
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "notification.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.Migrate(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO users(role, name, email, phone, password_hash) VALUES ('customer', '顾客', 'customer@example.com', '13800000000', 'hash'), ('owner', '老板', 'owner@example.com', '13900000000', 'hash')`); err != nil {
		t.Fatal(err)
	}
	return NewService(database, sender)
}

func TestInAppNotificationsAndEmailQueue(t *testing.T) {
	sender := &captureSender{}
	service := notificationService(t, sender)
	ctx := context.Background()
	if err := service.CreateBookingNotifications(ctx, 7, 1, 2, "2026-09-10T10:00:00+08:00", "测试乐队"); err != nil {
		t.Fatal(err)
	}
	items, err := service.List(ctx, 1, true)
	if err != nil || len(items) != 1 {
		t.Fatalf("unexpected customer notifications: %+v, %v", items, err)
	}
	if count, _ := service.UnreadCount(ctx, 1); count != 1 {
		t.Fatalf("expected one unread notification, got %d", count)
	}
	if err := service.MarkRead(ctx, 1, items[0].ID); err != nil {
		t.Fatal(err)
	}
	if count, _ := service.UnreadCount(ctx, 1); count != 0 {
		t.Fatalf("expected no unread notifications, got %d", count)
	}
	if err := service.ProcessEmailQueue(ctx); err != nil {
		t.Fatal(err)
	}
	if len(sender.emails) != 2 {
		t.Fatalf("expected customer and owner emails, got %d", len(sender.emails))
	}
}

func TestEmailFailureIsRecordedWithoutChangingBusinessData(t *testing.T) {
	sender := &captureSender{err: errors.New("mail server unavailable")}
	service := notificationService(t, sender)
	ctx := context.Background()
	if err := service.CreateBookingNotifications(ctx, 8, 1, 1, "2026-09-10T10:00:00+08:00", "失败测试"); err != nil {
		t.Fatal(err)
	}
	if err := service.ProcessEmailQueue(ctx); err != nil {
		t.Fatal(err)
	}
	var failed int
	if err := service.db.QueryRow(`SELECT COUNT(*) FROM notifications WHERE email_status = 'failed'`).Scan(&failed); err != nil {
		t.Fatal(err)
	}
	if failed != 2 {
		t.Fatalf("expected two failed emails, got %d", failed)
	}
}

func TestMembershipExpiryReminderIsIdempotent(t *testing.T) {
	service := notificationService(t, &captureSender{})
	ctx := context.Background()
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	if _, err := service.db.Exec(`INSERT INTO membership_plans(plan_type, name, price_cents) VALUES ('monthly', '月卡', 1)`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.db.Exec(`INSERT INTO membership_cards(user_id, plan_id, starts_at, ends_at, status, paid_amount_cents) VALUES (1, 1, '2026-09-01T00:00:00Z', '2026-09-09T12:00:00Z', 'active', 1)`); err != nil {
		t.Fatal(err)
	}
	created, err := service.CreateMembershipExpiryReminders(ctx, now)
	if err != nil || created != 1 {
		t.Fatalf("expected one reminder, created=%d err=%v", created, err)
	}
	created, err = service.CreateMembershipExpiryReminders(ctx, now)
	if err != nil || created != 0 {
		t.Fatalf("reminder should be idempotent, created=%d err=%v", created, err)
	}
}
