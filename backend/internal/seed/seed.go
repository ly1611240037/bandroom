package seed

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func SeedDemo(ctx context.Context, database *sql.DB) error {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO users(role, name, email, phone, password_hash, email_verified_at)
		VALUES ('owner', 'BandRoom 老板', 'owner@bandroom.test', '13800000000', 'demo-password-hash', CURRENT_TIMESTAMP)`); err != nil {
		return fmt.Errorf("seed owner: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO users(role, name, email, phone, password_hash, email_verified_at)
		VALUES ('customer', '演示联系人', 'customer@bandroom.test', '13900000000', 'demo-password-hash', CURRENT_TIMESTAMP)`); err != nil {
		return fmt.Errorf("seed customer: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO membership_plans(plan_type, name, price_cents, included_uses)
		VALUES ('monthly', '月卡', 99900, NULL), ('count', '10 次卡', 59900, 10)`); err != nil {
		return fmt.Errorf("seed plans: %w", err)
	}

	var customerID, monthlyPlanID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE email = 'customer@bandroom.test'`).Scan(&customerID); err != nil {
		return fmt.Errorf("find demo customer: %w", err)
	}
	if err := tx.QueryRowContext(ctx, `SELECT id FROM membership_plans WHERE name = '月卡'`).Scan(&monthlyPlanID); err != nil {
		return fmt.Errorf("find monthly plan: %w", err)
	}
	start := time.Now().UTC().Truncate(24 * time.Hour)
	end := start.AddDate(0, 0, 30)
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO membership_cards(user_id, plan_id, starts_at, ends_at, status, paid_amount_cents, payment_method, notes)
		SELECT ?, ?, ?, ?, 'active', 99900, '微信转账', '演示月卡' WHERE NOT EXISTS (
			SELECT 1 FROM membership_cards WHERE user_id = ? AND plan_id = ? AND notes = '演示月卡'
		)`, customerID, monthlyPlanID, start.Format(time.RFC3339), end.Format(time.RFC3339), customerID, monthlyPlanID); err != nil {
		return fmt.Errorf("seed monthly card: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO rooms(name, description, capacity) VALUES ('A 房', '适合乐队排练的标准房间', 6), ('B 房', '带架子鼓的宽敞房间', 8), ('C 房', '适合录音和小编制排练', 4)`); err != nil {
		return fmt.Errorf("seed rooms: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO fixed_equipment(room_id, name, description, quantity)
		SELECT id, '架子鼓', '房间固定架子鼓', 1 FROM rooms WHERE name = 'B 房'`); err != nil {
		return fmt.Errorf("seed fixed equipment: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO public_equipment(name, description, quantity)
		VALUES ('麦克风', '公共麦克风', 8), ('音箱', '公共音箱', 3), ('连接线', '常用连接线', 20)`); err != nil {
		return fmt.Errorf("seed public equipment: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO venue_content(content_key, content_value)
		VALUES ('name', 'BandRoom 音乐排练空间'), ('description', '面向乐队的多房间会员制排练空间'), ('phone', '400-000-0000')`); err != nil {
		return fmt.Errorf("seed venue content: %w", err)
	}

	return tx.Commit()
}
