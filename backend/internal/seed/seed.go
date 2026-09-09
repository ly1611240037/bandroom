package seed

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ly1611240037/bandroom/backend/internal/auth"
)

func SeedDemo(ctx context.Context, database *sql.DB) error {
	demoPasswordHash, err := auth.HashPassword("demo123456")
	if err != nil {
		return fmt.Errorf("hash demo password: %w", err)
	}
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `INSERT INTO users(role, name, email, phone, password_hash, email_verified_at)
		VALUES ('owner', 'BandRoom 老板', 'owner@bandroom.test', '13800000000', ?, CURRENT_TIMESTAMP)
		ON CONFLICT(email) DO UPDATE SET role = excluded.role, name = excluded.name, phone = excluded.phone, password_hash = excluded.password_hash, email_verified_at = COALESCE(users.email_verified_at, excluded.email_verified_at)`, demoPasswordHash); err != nil {
		return fmt.Errorf("seed owner: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO users(role, name, email, phone, password_hash, email_verified_at)
		VALUES ('customer', '演示联系人', 'customer@bandroom.test', '13900000000', ?, CURRENT_TIMESTAMP)
		ON CONFLICT(email) DO UPDATE SET role = excluded.role, name = excluded.name, phone = excluded.phone, password_hash = excluded.password_hash, email_verified_at = COALESCE(users.email_verified_at, excluded.email_verified_at)`, demoPasswordHash); err != nil {
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
	if err := seedRoomScenes(ctx, tx); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO public_equipment(name, description, quantity)
		VALUES ('麦克风', '公共麦克风', 8), ('音箱', '公共音箱', 3), ('连接线', '常用连接线', 20)`); err != nil {
		return fmt.Errorf("seed public equipment: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO venue_content(content_key, content_value)
		VALUES ('name', 'BandRoom 音乐排练空间'), ('description', '面向乐队的多房间会员制排练空间'), ('phone', '400-000-0000'), ('announcement', '欢迎提前预约，设备可在预约时一并借用'), ('membershipInstructions', '月卡或次数卡，老板线下开通；预约不单独收取租金'), ('openingRules', '每日 08:00 — 23:00，预约按半小时计算并预留 30 分钟清场')`); err != nil {
		return fmt.Errorf("seed venue content: %w", err)
	}

	return tx.Commit()
}

// Upgrade only the original demo descriptions; preserve owner-customized rooms.
func seedRoomScenes(ctx context.Context, tx *sql.Tx) error {
	scenes := []struct {
		name, original, description string
		equipment                   [][3]string
	}{
		{"A 房", "适合乐队排练的标准房间", "落日合奏室 · 约 28㎡，适合 4–6 人流行、摇滚乐队。鼓手靠后、吉他与贝斯分列两侧，中央留给主唱；建议自带吉他、贝斯和鼓棒。演示场景。", [][3]string{
			{"五鼓架子鼓", "一套五鼓配置：底鼓 1、军鼓 1、悬挂通鼓 2、落地通鼓 1；另含踩镲 1 对、吊镲 1、叮叮镲 1、支架与鼓凳，鼓棒自备", "1"},
			{"吉他音箱", "左右分区的电吉他扩音", "2"},
			{"贝斯音箱", "贝斯专用扩音", "1"},
			{"人声扩声系统", "含调音台与一对主扩；麦克风可另借", "1"},
		}},
		{"B 房", "带架子鼓的宽敞房间", "蓝调大排练室 · 约 38㎡，适合 6–8 人完整编制。键盘、人声与节奏组可同时排练；适合校园演出前的整场走台。演示场景。", [][3]string{
			{"架子鼓", "一套五鼓配置：底鼓 1、军鼓 1、悬挂通鼓 2、落地通鼓 1；另含踩镲 1 对、吊镲 1、叮叮镲 1、支架与鼓凳，鼓棒自备", "1"},
			{"吉他音箱", "双吉他声部分区扩音", "2"},
			{"贝斯音箱", "低频节奏声部扩音", "1"},
			{"电钢琴", "88 键，含琴架与踏板", "1"},
			{"人声扩声系统", "含调音台与一对主扩；麦克风可另借", "1"},
		}},
		{"C 房", "适合录音和小编制排练", "微光创作室 · 约 16㎡，适合 1–4 人人声、木吉他与键盘小编制创作。无原声鼓，适合练唱和编曲讨论；不提供专业录音服务。演示场景。", [][3]string{
			{"电钢琴", "88 键，含琴架与踏板", "1"},
			{"木吉他音箱", "木吉他拾音与扩声", "1"},
			{"人声扩声系统", "小编制人声扩声；麦克风可另借", "1"},
			{"谱架", "可调高度，随房使用", "2"},
		}},
	}
	for _, scene := range scenes {
		if _, err := tx.ExecContext(ctx, `UPDATE rooms SET description = ? WHERE name = ? AND description = ?`, scene.description, scene.name, scene.original); err != nil {
			return fmt.Errorf("seed room scene: %w", err)
		}
		for _, item := range scene.equipment {
			if _, err := tx.ExecContext(ctx, `INSERT INTO fixed_equipment(room_id, name, description, quantity)
				SELECT r.id, ?, ?, ? FROM rooms r WHERE r.name = ? AND r.description = ?
				AND NOT EXISTS (SELECT 1 FROM fixed_equipment f WHERE f.room_id = r.id AND f.name = ?)`, item[0], item[1], item[2], scene.name, scene.description, item[0]); err != nil {
				return fmt.Errorf("seed scene equipment: %w", err)
			}
		}
	}
	return nil
}
