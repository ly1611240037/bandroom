package membership

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNoMembership = errors.New("没有可用的会员卡，请联系老板开通会员")

type Plan struct {
	ID           int64  `json:"id"`
	PlanType     string `json:"planType"`
	Name         string `json:"name"`
	PriceCents   int64  `json:"priceCents"`
	IncludedUses *int   `json:"includedUses,omitempty"`
	IsActive     bool   `json:"isActive"`
}

type Card struct {
	ID              int64  `json:"id"`
	UserID          int64  `json:"userId"`
	PlanID          int64  `json:"planId"`
	PlanType        string `json:"planType"`
	PlanName        string `json:"planName"`
	StartsAt        string `json:"startsAt"`
	EndsAt          string `json:"endsAt,omitempty"`
	RemainingUses   *int   `json:"remainingUses,omitempty"`
	Status          string `json:"status"`
	PaidAmountCents int64  `json:"paidAmountCents"`
	PaymentMethod   string `json:"paymentMethod"`
	Notes           string `json:"notes"`
}
type CustomerSummary struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Verified  bool   `json:"verified"`
	CardCount int    `json:"cardCount"`
}

type Service struct{ db *sql.DB }

func NewService(database *sql.DB) *Service { return &Service{db: database} }

func (s *Service) CreatePlan(ctx context.Context, planType, name string, priceCents int64, includedUses *int) (Plan, error) {
	planType = strings.TrimSpace(planType)
	name = strings.TrimSpace(name)
	if planType != "monthly" && planType != "count" {
		return Plan{}, errors.New("方案类型必须是 monthly 或 count")
	}
	if name == "" || priceCents < 0 {
		return Plan{}, errors.New("方案名称不能为空，价格不能为负数")
	}
	if planType == "monthly" {
		includedUses = nil
	}
	if planType == "count" && (includedUses == nil || *includedUses <= 0) {
		return Plan{}, errors.New("次数卡必须填写正整数次数")
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO membership_plans(plan_type, name, price_cents, included_uses) VALUES (?, ?, ?, ?)`, planType, name, priceCents, includedUses)
	if err != nil {
		return Plan{}, fmt.Errorf("create membership plan: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Plan{}, err
	}
	return s.GetPlan(ctx, id)
}

func (s *Service) GetPlan(ctx context.Context, id int64) (Plan, error) {
	var plan Plan
	var active int
	var uses sql.NullInt64
	err := s.db.QueryRowContext(ctx, `SELECT id, plan_type, name, price_cents, included_uses, is_active FROM membership_plans WHERE id = ?`, id).Scan(&plan.ID, &plan.PlanType, &plan.Name, &plan.PriceCents, &uses, &active)
	if err != nil {
		return Plan{}, err
	}
	if uses.Valid {
		value := int(uses.Int64)
		plan.IncludedUses = &value
	}
	plan.IsActive = active == 1
	return plan, nil
}

func (s *Service) ListPlans(ctx context.Context, includeDisabled bool) ([]Plan, error) {
	query := `SELECT id, plan_type, name, price_cents, included_uses, is_active FROM membership_plans`
	if !includeDisabled {
		query += ` WHERE is_active = 1`
	}
	query += ` ORDER BY plan_type, id`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	plans := make([]Plan, 0)
	for rows.Next() {
		var plan Plan
		var active int
		var uses sql.NullInt64
		if err := rows.Scan(&plan.ID, &plan.PlanType, &plan.Name, &plan.PriceCents, &uses, &active); err != nil {
			return nil, err
		}
		if uses.Valid {
			value := int(uses.Int64)
			plan.IncludedUses = &value
		}
		plan.IsActive = active == 1
		plans = append(plans, plan)
	}
	return plans, rows.Err()
}

func (s *Service) SetPlanActive(ctx context.Context, id int64, active bool) error {
	result, err := s.db.ExecContext(ctx, `UPDATE membership_plans SET is_active = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, boolInt(active), id)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Service) UpdatePlan(ctx context.Context, id int64, name string, priceCents int64, includedUses *int) (Plan, error) {
	name = strings.TrimSpace(name)
	if name == "" || priceCents < 0 {
		return Plan{}, errors.New("方案名称不能为空，价格不能为负数")
	}
	plan, err := s.GetPlan(ctx, id)
	if err != nil {
		return Plan{}, err
	}
	if plan.PlanType == "monthly" {
		includedUses = nil
	}
	if plan.PlanType == "count" && (includedUses == nil || *includedUses <= 0) {
		return Plan{}, errors.New("次数卡必须填写正整数次数")
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE membership_plans SET name = ?, price_cents = ?, included_uses = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, name, priceCents, includedUses, id); err != nil {
		return Plan{}, err
	}
	return s.GetPlan(ctx, id)
}

func (s *Service) ActivateCard(ctx context.Context, userID, planID int64, paidAmount int64, paymentMethod, notes string, purchaseDate time.Time) (Card, error) {
	if userID <= 0 || planID <= 0 || paidAmount < 0 {
		return Card{}, errors.New("开卡参数不正确")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Card{}, err
	}
	defer tx.Rollback()
	var planType, planName string
	var included sql.NullInt64
	var active int
	if err := tx.QueryRowContext(ctx, `SELECT plan_type, name, included_uses, is_active FROM membership_plans WHERE id = ?`, planID).Scan(&planType, &planName, &included, &active); err != nil {
		return Card{}, err
	}
	if active != 1 {
		return Card{}, errors.New("该会员方案已停用")
	}
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE id = ? AND role = 'customer'`, userID).Scan(&exists); err != nil {
		return Card{}, err
	}
	if exists != 1 {
		return Card{}, errors.New("顾客不存在")
	}
	start := purchaseDate.UTC().Truncate(24 * time.Hour)
	var endsAt *string
	if planType == "monthly" {
		var latest sql.NullString
		err = tx.QueryRowContext(ctx, `SELECT MAX(CASE WHEN ends_at IS NOT NULL AND ends_at > ? THEN ends_at ELSE NULL END) FROM membership_cards WHERE user_id = ? AND plan_id IN (SELECT id FROM membership_plans WHERE plan_type = 'monthly') AND status IN ('active', 'scheduled')`, start.Format(time.RFC3339), userID).Scan(&latest)
		if err != nil {
			return Card{}, err
		}
		if latest.Valid {
			parsed, parseErr := time.Parse(time.RFC3339, latest.String)
			if parseErr == nil && parsed.After(start) {
				start = parsed
			}
		}
		end := start.AddDate(0, 0, 30).Format(time.RFC3339)
		endsAt = &end
	}
	status := "active"
	if start.After(purchaseDate.UTC()) {
		status = "scheduled"
	}
	var remaining any
	if planType == "count" {
		remaining = int(included.Int64)
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO membership_cards(user_id, plan_id, starts_at, ends_at, remaining_uses, status, paid_amount_cents, payment_method, notes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, userID, planID, start.Format(time.RFC3339), endsAt, remaining, status, paidAmount, strings.TrimSpace(paymentMethod), strings.TrimSpace(notes))
	if err != nil {
		return Card{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Card{}, err
	}
	if err := tx.Commit(); err != nil {
		return Card{}, err
	}
	return s.GetCard(ctx, id)
}

func (s *Service) GetCard(ctx context.Context, id int64) (Card, error) {
	return s.cardQuery(ctx, `WHERE c.id = ?`, id)
}

func (s *Service) ListCards(ctx context.Context, userID int64) ([]Card, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT c.id, c.user_id, c.plan_id, p.plan_type, p.name, c.starts_at, c.ends_at, c.remaining_uses, c.status, c.paid_amount_cents, c.payment_method, c.notes FROM membership_cards c JOIN membership_plans p ON p.id = c.plan_id WHERE c.user_id = ? ORDER BY c.starts_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cards := make([]Card, 0)
	for rows.Next() {
		card, err := scanCard(rows)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, rows.Err()
}

func (s *Service) ListCustomers(ctx context.Context, keyword string) ([]CustomerSummary, error) {
	like := "%" + keyword + "%"
	rows, err := s.db.QueryContext(ctx, `SELECT u.id, u.name, u.email, u.phone, u.email_verified_at, COUNT(c.id) FROM users u LEFT JOIN membership_cards c ON c.user_id = u.id WHERE u.role = 'customer' AND (? = '' OR u.name LIKE ? OR u.email LIKE ? OR u.phone LIKE ?) GROUP BY u.id ORDER BY u.id`, keyword, like, like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]CustomerSummary, 0)
	for rows.Next() {
		var item CustomerSummary
		var verified sql.NullString
		if err := rows.Scan(&item.ID, &item.Name, &item.Email, &item.Phone, &verified, &item.CardCount); err != nil {
			return nil, err
		}
		item.Verified = verified.Valid
		result = append(result, item)
	}
	return result, rows.Err()
}

func (s *Service) Eligibility(ctx context.Context, userID int64, at time.Time) (Card, error) {
	now := at.UTC().Format(time.RFC3339)
	card, err := s.cardQuery(ctx, `WHERE c.user_id = ? AND p.plan_type = 'monthly' AND c.status IN ('active', 'scheduled') AND c.starts_at <= ? AND (c.ends_at IS NULL OR c.ends_at > ?) ORDER BY c.starts_at DESC LIMIT 1`, userID, now, now)
	if err == nil {
		return card, nil
	}
	card, err = s.cardQuery(ctx, `WHERE c.user_id = ? AND p.plan_type = 'count' AND c.status IN ('active', 'scheduled') AND c.starts_at <= ? AND c.remaining_uses > 0 ORDER BY c.starts_at ASC LIMIT 1`, userID, now)
	if err != nil {
		return Card{}, ErrNoMembership
	}
	return card, nil
}

func (s *Service) ConsumeForBooking(ctx context.Context, userID int64, at time.Time) (Card, error) {
	card, err := s.Eligibility(ctx, userID, at)
	if err != nil {
		return Card{}, err
	}
	if card.PlanType == "monthly" {
		return card, nil
	}
	result, err := s.db.ExecContext(ctx, `UPDATE membership_cards SET remaining_uses = remaining_uses - 1, status = CASE WHEN remaining_uses - 1 <= 0 THEN 'depleted' ELSE status END WHERE id = ? AND remaining_uses > 0`, card.ID)
	if err != nil {
		return Card{}, err
	}
	changed, _ := result.RowsAffected()
	if changed != 1 {
		return Card{}, ErrNoMembership
	}
	return s.GetCard(ctx, card.ID)
}

func (s *Service) RestoreCountUse(ctx context.Context, cardID int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE membership_cards SET remaining_uses = remaining_uses + 1, status = 'active' WHERE id = ? AND remaining_uses IS NOT NULL`, cardID)
	return err
}

func (s *Service) cardQuery(ctx context.Context, where string, args ...any) (Card, error) {
	row := s.db.QueryRowContext(ctx, `SELECT c.id, c.user_id, c.plan_id, p.plan_type, p.name, c.starts_at, c.ends_at, c.remaining_uses, c.status, c.paid_amount_cents, c.payment_method, c.notes FROM membership_cards c JOIN membership_plans p ON p.id = c.plan_id `+where, args...)
	return scanCard(row)
}

type scanner interface{ Scan(dest ...any) error }

func scanCard(row scanner) (Card, error) {
	var card Card
	var ends, remaining sql.NullString
	if err := row.Scan(&card.ID, &card.UserID, &card.PlanID, &card.PlanType, &card.PlanName, &card.StartsAt, &ends, &remaining, &card.Status, &card.PaidAmountCents, &card.PaymentMethod, &card.Notes); err != nil {
		return Card{}, err
	}
	if ends.Valid {
		card.EndsAt = ends.String
	}
	if remaining.Valid {
		value := 0
		_, _ = fmt.Sscan(remaining.String, &value)
		card.RemainingUses = &value
	}
	return card, nil
}
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
