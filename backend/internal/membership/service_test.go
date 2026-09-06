package membership

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/ly1611240037/bandroom/backend/internal/db"
)

func membershipService(t *testing.T) *Service {
	t.Helper()
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "membership.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.Migrate(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO users(role, name, email, phone, password_hash, email_verified_at) VALUES ('customer', '顾客', 'customer@example.com', '13800000000', 'hash', CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}
	return NewService(database)
}

func TestPlansAndManualCardActivation(t *testing.T) {
	service := membershipService(t)
	ctx := context.Background()
	uses := 10
	monthly, err := service.CreatePlan(ctx, "monthly", "月卡", 99900, nil)
	if err != nil {
		t.Fatal(err)
	}
	count, err := service.CreatePlan(ctx, "count", "10 次卡", 59900, &uses)
	if err != nil {
		t.Fatal(err)
	}
	purchase := time.Date(2026, 9, 6, 12, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	first, err := service.ActivateCard(ctx, 1, monthly.ID, 99900, "微信转账", "首次购买", purchase)
	if err != nil {
		t.Fatal(err)
	}
	if first.Status != "active" || first.EndsAt == "" {
		t.Fatalf("unexpected monthly card: %+v", first)
	}
	second, err := service.ActivateCard(ctx, 1, monthly.ID, 99900, "微信转账", "续卡", purchase)
	if err != nil {
		t.Fatal(err)
	}
	if second.Status != "scheduled" || second.StartsAt != first.EndsAt {
		t.Fatalf("monthly cards should be sequential: first=%+v second=%+v", first, second)
	}
	countCard, err := service.ActivateCard(ctx, 1, count.ID, 59900, "现金", "备用次数卡", purchase)
	if err != nil {
		t.Fatal(err)
	}
	if countCard.RemainingUses == nil || *countCard.RemainingUses != 10 {
		t.Fatalf("unexpected count card: %+v", countCard)
	}
}

func TestMonthlyPriorityAndCountRestore(t *testing.T) {
	service := membershipService(t)
	ctx := context.Background()
	uses := 2
	monthly, _ := service.CreatePlan(ctx, "monthly", "月卡", 1, nil)
	count, _ := service.CreatePlan(ctx, "count", "次数卡", 1, &uses)
	now := time.Now().UTC()
	countCard, err := service.ActivateCard(ctx, 1, count.ID, 1, "现金", "", now)
	if err != nil {
		t.Fatal(err)
	}
	monthlyCard, err := service.ActivateCard(ctx, 1, monthly.ID, 1, "现金", "", now)
	if err != nil {
		t.Fatal(err)
	}
	eligible, err := service.Eligibility(ctx, 1, now)
	if err != nil || eligible.ID != monthlyCard.ID {
		t.Fatalf("monthly card should have priority: %+v, %v", eligible, err)
	}
	used, err := service.ConsumeForBooking(ctx, 1, now)
	if err != nil || used.ID != monthlyCard.ID {
		t.Fatalf("booking should use monthly card: %+v, %v", used, err)
	}
	// Disable the monthly card to demonstrate count consumption and restoration.
	if _, err := service.db.Exec(`UPDATE membership_cards SET status = 'disabled' WHERE id = ?`, monthlyCard.ID); err != nil {
		t.Fatal(err)
	}
	used, err = service.ConsumeForBooking(ctx, 1, now)
	if err != nil || used.ID != countCard.ID || used.RemainingUses == nil || *used.RemainingUses != 1 {
		t.Fatalf("count card should be consumed: %+v, %v", used, err)
	}
	if err := service.RestoreCountUse(ctx, countCard.ID); err != nil {
		t.Fatal(err)
	}
	restored, err := service.GetCard(ctx, countCard.ID)
	if err != nil || restored.RemainingUses == nil || *restored.RemainingUses != 2 {
		t.Fatalf("count use should be restored: %+v, %v", restored, err)
	}
}

func TestExhaustedCountCardRejectsEligibility(t *testing.T) {
	service := membershipService(t)
	ctx := context.Background()
	uses := 1
	count, _ := service.CreatePlan(ctx, "count", "次数卡", 1, &uses)
	card, err := service.ActivateCard(ctx, 1, count.ID, 1, "现金", "", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ConsumeForBooking(ctx, 1, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Eligibility(ctx, 1, time.Now().UTC()); err != ErrNoMembership {
		t.Fatalf("expected exhausted card to reject, got %v", err)
	}
	if card.ID == 0 {
		t.Fatal("card was not created")
	}
}
