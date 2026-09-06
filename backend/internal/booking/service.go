package booking

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ly1611240037/bandroom/backend/internal/timeutil"
)

var (
	ErrConflict     = errors.New("所选时段已被占用")
	ErrNoMembership = errors.New("没有可用的会员卡，请先联系老板开通")
)

type EquipmentRequest struct {
	EquipmentID int64 `json:"equipmentId"`
	Quantity    int   `json:"quantity"`
}
type Booking struct {
	ID               int64              `json:"id"`
	UserID           int64              `json:"userId"`
	RoomID           int64              `json:"roomId"`
	MembershipCardID int64              `json:"membershipCardId"`
	BandName         string             `json:"bandName"`
	Phone            string             `json:"phone"`
	StartsAt         string             `json:"startsAt"`
	EndsAt           string             `json:"endsAt"`
	OccupiedUntil    string             `json:"occupiedUntil"`
	Status           string             `json:"status"`
	Notes            string             `json:"notes"`
	Equipment        []EquipmentRequest `json:"equipment"`
}
type Slot struct {
	StartsAt        string `json:"startsAt"`
	EndsAt          string `json:"endsAt"`
	DurationMinutes int    `json:"durationMinutes"`
}
type Service struct {
	db       *sql.DB
	location *time.Location
}

func NewService(database *sql.DB) *Service {
	return &Service{db: database, location: time.FixedZone("China Standard Time", 8*60*60)}
}

func (s *Service) Availability(ctx context.Context, roomID int64, date string) ([]Slot, error) {
	day, err := time.ParseInLocation("2006-01-02", date, s.location)
	if err != nil {
		return nil, errors.New("日期格式应为 YYYY-MM-DD")
	}
	var opens, closes string
	var enabled int
	if err := s.db.QueryRowContext(ctx, `SELECT opens_at, closes_at, enabled FROM business_hours WHERE weekday = ?`, int(day.Weekday())).Scan(&opens, &closes, &enabled); err != nil {
		return nil, err
	}
	if enabled == 0 {
		return []Slot{}, nil
	}
	openAt, _ := time.ParseInLocation("15:04", opens, s.location)
	closeAt, _ := time.ParseInLocation("15:04", closes, s.location)
	open := time.Date(day.Year(), day.Month(), day.Day(), openAt.Hour(), openAt.Minute(), 0, 0, s.location)
	close := time.Date(day.Year(), day.Month(), day.Day(), closeAt.Hour(), closeAt.Minute(), 0, 0, s.location)
	closures, err := s.closures(ctx, roomID)
	if err != nil {
		return nil, err
	}
	result := make([]Slot, 0)
	for start := open; start.Before(close); start = start.Add(30 * time.Minute) {
		if start.Before(time.Now().In(s.location)) {
			continue
		}
		for duration := timeutil.MinBookingMin; duration <= timeutil.MaxBookingMin; duration += timeutil.SlotMinutes {
			end := start.Add(time.Duration(duration) * time.Minute)
			occupied := timeutil.OccupiedEnd(end)
			if occupied.After(close) || s.overlapsClosure(start, occupied, closures) {
				continue
			}
			free, err := s.roomFree(ctx, roomID, start, occupied)
			if err != nil {
				return nil, err
			}
			if free {
				result = append(result, Slot{StartsAt: start.Format(time.RFC3339), EndsAt: end.Format(time.RFC3339), DurationMinutes: duration})
			}
		}
	}
	return result, nil
}

func (s *Service) Create(ctx context.Context, userID int64, roomID int64, bandName, phone, startsAt, endsAt, notes string, equipment []EquipmentRequest) (Booking, error) {
	start, end, err := s.validateInterval(startsAt, endsAt)
	if err != nil {
		return Booking{}, err
	}
	if strings.TrimSpace(bandName) == "" || strings.TrimSpace(phone) == "" {
		return Booking{}, errors.New("乐队名称和手机号不能为空")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Booking{}, err
	}
	defer tx.Rollback()
	var role string
	var verified sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT role, email_verified_at FROM users WHERE id = ?`, userID).Scan(&role, &verified); err != nil {
		return Booking{}, err
	}
	if role != "customer" || !verified.Valid {
		return Booking{}, errors.New("请先完成顾客邮箱验证")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	var future int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM bookings WHERE user_id = ? AND status = 'confirmed' AND datetime(starts_at) > datetime(?)`, userID, now).Scan(&future); err != nil {
		return Booking{}, err
	}
	if future > 0 {
		return Booking{}, errors.New("同一时间只能保留一条未来预约")
	}
	var roomStatus string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM rooms WHERE id = ?`, roomID).Scan(&roomStatus); err != nil {
		return Booking{}, err
	}
	if roomStatus != "available" {
		return Booking{}, errors.New("该房间当前不可预约")
	}
	occupied := timeutil.OccupiedEnd(end)
	free, err := s.roomFreeTx(ctx, tx, roomID, start, occupied)
	if err != nil {
		return Booking{}, err
	}
	if !free {
		return Booking{}, ErrConflict
	}
	if err := s.checkScheduleTx(ctx, tx, roomID, start, occupied); err != nil {
		return Booking{}, err
	}
	if err := s.checkEquipmentTx(ctx, tx, equipment, start, occupied); err != nil {
		return Booking{}, err
	}
	cardID, err := s.consumeMembershipTx(ctx, tx, userID, start)
	if err != nil {
		return Booking{}, err
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO bookings(user_id, room_id, membership_card_id, band_name, phone, starts_at, ends_at, occupied_until, status, notes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'confirmed', ?)`, userID, roomID, cardID, strings.TrimSpace(bandName), strings.TrimSpace(phone), start.Format(time.RFC3339), end.Format(time.RFC3339), occupied.Format(time.RFC3339), strings.TrimSpace(notes))
	if err != nil {
		return Booking{}, err
	}
	id, _ := result.LastInsertId()
	for _, item := range equipment {
		if _, err := tx.ExecContext(ctx, `INSERT INTO booking_equipment(booking_id, equipment_id, quantity) VALUES (?, ?, ?)`, id, item.EquipmentID, item.Quantity); err != nil {
			return Booking{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return Booking{}, err
	}
	return s.Get(ctx, id)
}

func (s *Service) ListByUser(ctx context.Context, userID int64) ([]Booking, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM bookings WHERE user_id = ? ORDER BY starts_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Booking, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		item, err := s.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (s *Service) Get(ctx context.Context, id int64) (Booking, error) {
	var item Booking
	if err := s.db.QueryRowContext(ctx, `SELECT id, user_id, room_id, membership_card_id, band_name, phone, starts_at, ends_at, occupied_until, status, notes FROM bookings WHERE id = ?`, id).Scan(&item.ID, &item.UserID, &item.RoomID, &item.MembershipCardID, &item.BandName, &item.Phone, &item.StartsAt, &item.EndsAt, &item.OccupiedUntil, &item.Status, &item.Notes); err != nil {
		return Booking{}, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT equipment_id, quantity FROM booking_equipment WHERE booking_id = ?`, id)
	if err != nil {
		return Booking{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var equipment EquipmentRequest
		if err := rows.Scan(&equipment.EquipmentID, &equipment.Quantity); err != nil {
			return Booking{}, err
		}
		item.Equipment = append(item.Equipment, equipment)
	}
	return item, rows.Err()
}

func (s *Service) CancelByCustomer(ctx context.Context, userID, bookingID int64, reason string) error {
	return s.cancel(ctx, bookingID, userID, false, reason)
}
func (s *Service) CancelByOwner(ctx context.Context, actorID, bookingID int64, reason string) error {
	return s.cancel(ctx, bookingID, actorID, true, reason)
}
func (s *Service) cancel(ctx context.Context, bookingID, userID int64, owner bool, reason string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var bookingUser, cardID int64
	var status, planType, bandName string
	if err := tx.QueryRowContext(ctx, `SELECT b.user_id, b.membership_card_id, b.status, b.band_name, p.plan_type FROM bookings b JOIN membership_cards c ON c.id = b.membership_card_id JOIN membership_plans p ON p.id = c.plan_id WHERE b.id = ?`, bookingID).Scan(&bookingUser, &cardID, &status, &bandName, &planType); err != nil {
		return err
	}
	if !owner && bookingUser != userID {
		return errors.New("不能操作其他顾客的预约")
	}
	if status != "confirmed" {
		return errors.New("当前预约不能取消")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE bookings SET status = 'cancelled', cancellation_reason = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'confirmed'`, strings.TrimSpace(reason), bookingID); err != nil {
		return err
	}
	if planType == "count" {
		if _, err := tx.ExecContext(ctx, `UPDATE membership_cards SET remaining_uses = remaining_uses + 1, status = 'active' WHERE id = ?`, cardID); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO notifications(user_id, notification_type, title, body, email_status) VALUES (?, 'booking_cancelled', '预约已取消', ?, 'not_required')`, bookingUser, fmt.Sprintf("%s 的预约已取消。原因：%s", bandName, strings.TrimSpace(reason))); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO audit_logs(actor_user_id, action, target_type, target_id, after_json) VALUES (?, 'cancel_booking', 'booking', ?, ?)`, userID, bookingID, fmt.Sprintf(`{"reason":%q}`, strings.TrimSpace(reason))); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *Service) CompleteDue(ctx context.Context, at time.Time) (int64, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE bookings SET status = 'completed', updated_at = CURRENT_TIMESTAMP WHERE status = 'confirmed' AND datetime(occupied_until) <= datetime(?)`, at.UTC().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	count, _ := result.RowsAffected()
	return count, nil
}
func (s *Service) MarkNoShow(ctx context.Context, actorID, bookingID int64) error {
	result, err := s.db.ExecContext(ctx, `UPDATE bookings SET status = 'no_show', updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'confirmed'`, bookingID)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count != 1 {
		return errors.New("当前预约不能标记为未到场")
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO audit_logs(actor_user_id, action, target_type, target_id) VALUES (?, 'mark_no_show', 'booking', ?)`, actorID, bookingID); err != nil {
		return err
	}
	return nil
}
func (s *Service) OwnerEdit(ctx context.Context, actorID, bookingID int64, startsAt, endsAt, notes string) error {
	start, end, err := s.validateInterval(startsAt, endsAt)
	if err != nil {
		return err
	}
	occupied := timeutil.OccupiedEnd(end)
	result, err := s.db.ExecContext(ctx, `UPDATE bookings SET starts_at = ?, ends_at = ?, occupied_until = ?, notes = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'confirmed'`, start.Format(time.RFC3339), end.Format(time.RFC3339), occupied.Format(time.RFC3339), strings.TrimSpace(notes), bookingID)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count != 1 {
		return errors.New("当前预约不能修改")
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO audit_logs(actor_user_id, action, target_type, target_id, after_json) VALUES (?, 'edit_booking', 'booking', ?, ?)`, actorID, bookingID, fmt.Sprintf(`{"startsAt":%q,"endsAt":%q}`, startsAt, endsAt)); err != nil {
		return err
	}
	return nil
}

func (s *Service) validateInterval(startsAt, endsAt string) (time.Time, time.Time, error) {
	start, err := time.Parse(time.RFC3339, startsAt)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("开始时间格式不正确")
	}
	end, err := time.Parse(time.RFC3339, endsAt)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("结束时间格式不正确")
	}
	if !timeutil.IsHalfHour(start) || !timeutil.IsHalfHour(end) || !timeutil.IsValidDuration(int(end.Sub(start).Minutes())) {
		return time.Time{}, time.Time{}, errors.New("预约必须按半小时选择，时长为 0.5 至 3 小时")
	}
	if !end.After(start) {
		return time.Time{}, time.Time{}, errors.New("结束时间必须晚于开始时间")
	}
	if !start.After(time.Now()) {
		return time.Time{}, time.Time{}, errors.New("不能预约已经开始的时间")
	}
	return start, end, nil
}
func (s *Service) roomFree(ctx context.Context, roomID int64, start, occupied time.Time) (bool, error) {
	return s.roomFreeTx(ctx, nil, roomID, start, occupied)
}
func (s *Service) roomFreeTx(ctx context.Context, tx *sql.Tx, roomID int64, start, occupied time.Time) (bool, error) {
	query := `SELECT COUNT(*) FROM bookings WHERE room_id = ? AND status = 'confirmed' AND datetime(starts_at) < datetime(?) AND datetime(occupied_until) > datetime(?)`
	var count int
	var err error
	args := []any{roomID, occupied.Format(time.RFC3339), start.Format(time.RFC3339)}
	if tx != nil {
		err = tx.QueryRowContext(ctx, query, args...).Scan(&count)
	} else {
		err = s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	}
	return count == 0, err
}
func (s *Service) checkEquipmentTx(ctx context.Context, tx *sql.Tx, equipment []EquipmentRequest, start, occupied time.Time) error {
	for _, item := range equipment {
		if item.Quantity <= 0 {
			return errors.New("设备数量必须大于 0")
		}
		var status string
		var total int
		if err := tx.QueryRowContext(ctx, `SELECT status, quantity FROM public_equipment WHERE id = ?`, item.EquipmentID).Scan(&status, &total); err != nil {
			return err
		}
		if status != "available" {
			return errors.New("所选设备当前不可用")
		}
		var used int
		if err := tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(be.quantity), 0) FROM booking_equipment be JOIN bookings b ON b.id = be.booking_id WHERE be.equipment_id = ? AND b.status = 'confirmed' AND datetime(b.starts_at) < datetime(?) AND datetime(b.occupied_until) > datetime(?)`, item.EquipmentID, occupied.Format(time.RFC3339), start.Format(time.RFC3339)).Scan(&used); err != nil {
			return err
		}
		if used+item.Quantity > total {
			return fmt.Errorf("设备库存不足：设备 %d", item.EquipmentID)
		}
	}
	return nil
}
func (s *Service) consumeMembershipTx(ctx context.Context, tx *sql.Tx, userID int64, start time.Time) (int64, error) {
	now := start.Format(time.RFC3339)
	var id int64
	var planType string
	var remaining sql.NullInt64
	err := tx.QueryRowContext(ctx, `SELECT c.id, p.plan_type, c.remaining_uses FROM membership_cards c JOIN membership_plans p ON p.id = c.plan_id WHERE c.user_id = ? AND c.status IN ('active', 'scheduled') AND datetime(c.starts_at) <= datetime(?) AND (p.plan_type = 'monthly' AND (c.ends_at IS NULL OR datetime(c.ends_at) > datetime(?)) OR p.plan_type = 'count' AND c.remaining_uses > 0) ORDER BY CASE WHEN p.plan_type = 'monthly' THEN 0 ELSE 1 END, c.starts_at LIMIT 1`, userID, now, now).Scan(&id, &planType, &remaining)
	if err != nil {
		return 0, ErrNoMembership
	}
	if planType == "count" {
		result, err := tx.ExecContext(ctx, `UPDATE membership_cards SET remaining_uses = remaining_uses - 1, status = CASE WHEN remaining_uses - 1 <= 0 THEN 'depleted' ELSE status END WHERE id = ? AND remaining_uses > 0`, id)
		if err != nil {
			return 0, err
		}
		changed, _ := result.RowsAffected()
		if changed != 1 {
			return 0, ErrNoMembership
		}
	}
	return id, nil
}
func (s *Service) checkScheduleTx(ctx context.Context, tx *sql.Tx, roomID int64, start, occupied time.Time) error {
	var opens, closes string
	var enabled int
	if err := tx.QueryRowContext(ctx, `SELECT opens_at, closes_at, enabled FROM business_hours WHERE weekday = ?`, int(start.Weekday())).Scan(&opens, &closes, &enabled); err != nil {
		return err
	}
	if enabled == 0 {
		return errors.New("当天不营业")
	}
	open, _ := time.ParseInLocation("15:04", opens, start.Location())
	close, _ := time.ParseInLocation("15:04", closes, start.Location())
	closeAt := time.Date(start.Year(), start.Month(), start.Day(), close.Hour(), close.Minute(), 0, 0, start.Location())
	openAt := time.Date(start.Year(), start.Month(), start.Day(), open.Hour(), open.Minute(), 0, 0, start.Location())
	if start.Before(openAt) || occupied.After(closeAt) {
		return errors.New("预约时段超出营业时间")
	}
	rows, err := tx.QueryContext(ctx, `SELECT starts_at, ends_at FROM closures WHERE room_id IS NULL OR room_id = ?`, roomID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var a, b string
		if err := rows.Scan(&a, &b); err != nil {
			return err
		}
		closureStart, e1 := time.Parse(time.RFC3339, a)
		closureEnd, e2 := time.Parse(time.RFC3339, b)
		if e1 == nil && e2 == nil && start.Before(closureEnd) && occupied.After(closureStart) {
			return errors.New("所选时段处于闭店安排")
		}
	}
	return rows.Err()
}
func (s *Service) closures(ctx context.Context, roomID int64) ([][2]time.Time, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT starts_at, ends_at FROM closures WHERE room_id IS NULL OR room_id = ?`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([][2]time.Time, 0)
	for rows.Next() {
		var a, b string
		if err := rows.Scan(&a, &b); err != nil {
			return nil, err
		}
		start, e1 := time.Parse(time.RFC3339, a)
		end, e2 := time.Parse(time.RFC3339, b)
		if e1 == nil && e2 == nil {
			result = append(result, [2]time.Time{start, end})
		}
	}
	return result, rows.Err()
}
func (s *Service) overlapsClosure(start, end time.Time, closures [][2]time.Time) bool {
	for _, item := range closures {
		if start.Before(item[1]) && end.After(item[0]) {
			return true
		}
	}
	return false
}
