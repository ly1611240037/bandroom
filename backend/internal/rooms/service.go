package rooms

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var ErrUnavailable = errors.New("该房间当前不可预约")

type Room struct {
	ID             int64            `json:"id"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	Capacity       *int64           `json:"capacity,omitempty"`
	Status         string           `json:"status"`
	Photos         []Photo          `json:"photos"`
	FixedEquipment []FixedEquipment `json:"fixedEquipment"`
}
type Photo struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	SortOrder int    `json:"sortOrder"`
}
type FixedEquipment struct {
	ID          int64  `json:"id"`
	RoomID      int64  `json:"roomId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	Status      string `json:"status"`
}
type PublicEquipment struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Quantity         int    `json:"quantity"`
	HourlyPriceCents int64  `json:"hourlyPriceCents"`
	Status           string `json:"status"`
}
type BusinessHour struct {
	Weekday  int    `json:"weekday"`
	OpensAt  string `json:"opensAt"`
	ClosesAt string `json:"closesAt"`
	Enabled  bool   `json:"enabled"`
}
type Closure struct {
	ID       int64  `json:"id"`
	RoomID   *int64 `json:"roomId,omitempty"`
	StartsAt string `json:"startsAt"`
	EndsAt   string `json:"endsAt"`
	Reason   string `json:"reason"`
}
type Service struct{ db *sql.DB }

func NewService(db *sql.DB) *Service { return &Service{db: db} }

func (s *Service) ListRooms(ctx context.Context) ([]Room, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, description, capacity, status FROM rooms ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rooms := make([]Room, 0)
	for rows.Next() {
		room, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		if err := s.fillRoom(ctx, &room); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, rows.Err()
}
func (s *Service) GetRoom(ctx context.Context, id int64) (Room, error) {
	room, err := scanRoom(s.db.QueryRowContext(ctx, `SELECT id, name, description, capacity, status FROM rooms WHERE id = ?`, id))
	if err != nil {
		return Room{}, err
	}
	if err := s.fillRoom(ctx, &room); err != nil {
		return Room{}, err
	}
	return room, nil
}
func (s *Service) CreateRoom(ctx context.Context, name, description string, capacity *int64) (Room, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Room{}, errors.New("房间名称不能为空")
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO rooms(name, description, capacity) VALUES (?, ?, ?)`, name, strings.TrimSpace(description), capacity)
	if err != nil {
		return Room{}, err
	}
	id, _ := result.LastInsertId()
	return s.GetRoom(ctx, id)
}
func (s *Service) UpdateRoom(ctx context.Context, id int64, name, description string, capacity *int64, status string) error {
	if name == "" || (status != "available" && status != "maintenance" && status != "suspended") {
		return errors.New("房间资料或状态不正确")
	}
	result, err := s.db.ExecContext(ctx, `UPDATE rooms SET name = ?, description = ?, capacity = ?, status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, strings.TrimSpace(name), strings.TrimSpace(description), capacity, status, id)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count != 1 {
		return sql.ErrNoRows
	}
	return nil
}
func (s *Service) AddPhoto(ctx context.Context, roomID int64, url string, sortOrder int) error {
	if strings.TrimSpace(url) == "" {
		return errors.New("照片地址不能为空")
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO room_photos(room_id, url, sort_order) VALUES (?, ?, ?)`, roomID, strings.TrimSpace(url), sortOrder)
	return err
}
func (s *Service) AddFixedEquipment(ctx context.Context, roomID int64, name, description string, quantity int) (int64, error) {
	if strings.TrimSpace(name) == "" || quantity < 0 {
		return 0, errors.New("固定设备资料不正确")
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO fixed_equipment(room_id, name, description, quantity) VALUES (?, ?, ?, ?)`, roomID, name, description, quantity)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
func (s *Service) SetFixedEquipmentStatus(ctx context.Context, id int64, status string) error {
	if status != "available" && status != "maintenance" && status != "disabled" {
		return errors.New("固定设备状态不正确")
	}
	_, err := s.db.ExecContext(ctx, `UPDATE fixed_equipment SET status = ? WHERE id = ?`, status, id)
	return err
}
func (s *Service) ListEquipment(ctx context.Context) ([]PublicEquipment, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, description, quantity, hourly_price_cents, status FROM public_equipment ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]PublicEquipment, 0)
	for rows.Next() {
		var item PublicEquipment
		var price int64
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Quantity, &price, &item.Status); err != nil {
			return nil, err
		}
		item.HourlyPriceCents = price
		result = append(result, item)
	}
	return result, rows.Err()
}
func (s *Service) CreateEquipment(ctx context.Context, name, description string, quantity int, hourlyPriceCents int64) (PublicEquipment, error) {
	if strings.TrimSpace(name) == "" || quantity < 0 || hourlyPriceCents < 0 {
		return PublicEquipment{}, errors.New("公共设备资料不正确")
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO public_equipment(name, description, quantity, hourly_price_cents) VALUES (?, ?, ?, ?)`, name, description, quantity, hourlyPriceCents)
	if err != nil {
		return PublicEquipment{}, err
	}
	id, _ := result.LastInsertId()
	return s.GetEquipment(ctx, id)
}
func (s *Service) GetEquipment(ctx context.Context, id int64) (PublicEquipment, error) {
	var item PublicEquipment
	err := s.db.QueryRowContext(ctx, `SELECT id, name, description, quantity, hourly_price_cents, status FROM public_equipment WHERE id = ?`, id).Scan(&item.ID, &item.Name, &item.Description, &item.Quantity, &item.HourlyPriceCents, &item.Status)
	return item, err
}

func (s *Service) CanSelectEquipment(ctx context.Context, id int64, quantity int) error {
	if quantity <= 0 {
		return errors.New("设备数量必须大于 0")
	}
	var status string
	var available int
	if err := s.db.QueryRowContext(ctx, `SELECT status, quantity FROM public_equipment WHERE id = ?`, id).Scan(&status, &available); err != nil {
		return err
	}
	if status != "available" || quantity > available {
		return errors.New("该设备当前不可用或库存不足")
	}
	return nil
}
func (s *Service) UpdateEquipment(ctx context.Context, id int64, name, description string, quantity int, price int64, status string) error {
	if name == "" || quantity < 0 || price < 0 || (status != "available" && status != "maintenance" && status != "disabled") {
		return errors.New("公共设备资料不正确")
	}
	_, err := s.db.ExecContext(ctx, `UPDATE public_equipment SET name = ?, description = ?, quantity = ?, hourly_price_cents = ?, status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, name, description, quantity, price, status, id)
	return err
}
func (s *Service) ListBusinessHours(ctx context.Context) ([]BusinessHour, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT weekday, opens_at, closes_at, enabled FROM business_hours ORDER BY weekday`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]BusinessHour, 0)
	for rows.Next() {
		var item BusinessHour
		var enabled int
		if err := rows.Scan(&item.Weekday, &item.OpensAt, &item.ClosesAt, &enabled); err != nil {
			return nil, err
		}
		item.Enabled = enabled == 1
		result = append(result, item)
	}
	return result, rows.Err()
}
func (s *Service) SetBusinessHour(ctx context.Context, item BusinessHour) error {
	if item.Weekday < 0 || item.Weekday > 6 {
		return errors.New("星期编号不正确")
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO business_hours(weekday, opens_at, closes_at, enabled) VALUES (?, ?, ?, ?) ON CONFLICT(weekday) DO UPDATE SET opens_at = excluded.opens_at, closes_at = excluded.closes_at, enabled = excluded.enabled`, item.Weekday, item.OpensAt, item.ClosesAt, boolInt(item.Enabled))
	return err
}
func (s *Service) ListClosures(ctx context.Context) ([]Closure, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, room_id, starts_at, ends_at, reason FROM closures ORDER BY starts_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Closure, 0)
	for rows.Next() {
		var item Closure
		var room sql.NullInt64
		if err := rows.Scan(&item.ID, &room, &item.StartsAt, &item.EndsAt, &item.Reason); err != nil {
			return nil, err
		}
		if room.Valid {
			item.RoomID = &room.Int64
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (s *Service) AddClosure(ctx context.Context, roomID *int64, startsAt, endsAt, reason string) (int64, error) {
	if startsAt == "" || endsAt == "" {
		return 0, errors.New("闭店时段不能为空")
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO closures(room_id, starts_at, ends_at, reason) VALUES (?, ?, ?, ?)`, roomID, startsAt, endsAt, reason)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
func (s *Service) GetCleaningBuffer(ctx context.Context) (int, error) {
	var value int
	err := s.db.QueryRowContext(ctx, `SELECT setting_value FROM venue_settings WHERE setting_key = 'cleaning_buffer_minutes'`).Scan(&value)
	return value, err
}
func (s *Service) CanBookRoom(ctx context.Context, id int64) error {
	var status string
	if err := s.db.QueryRowContext(ctx, `SELECT status FROM rooms WHERE id = ?`, id).Scan(&status); err != nil {
		return err
	}
	if status != "available" {
		return ErrUnavailable
	}
	return nil
}
func (s *Service) fillRoom(ctx context.Context, room *Room) error {
	photos, err := s.db.QueryContext(ctx, `SELECT id, url, sort_order FROM room_photos WHERE room_id = ? ORDER BY sort_order, id`, room.ID)
	if err != nil {
		return err
	}
	defer photos.Close()
	room.Photos = []Photo{}
	for photos.Next() {
		var item Photo
		if err := photos.Scan(&item.ID, &item.URL, &item.SortOrder); err != nil {
			return err
		}
		room.Photos = append(room.Photos, item)
	}
	equipment, err := s.db.QueryContext(ctx, `SELECT id, room_id, name, description, quantity, status FROM fixed_equipment WHERE room_id = ? ORDER BY id`, room.ID)
	if err != nil {
		return err
	}
	defer equipment.Close()
	room.FixedEquipment = []FixedEquipment{}
	for equipment.Next() {
		var item FixedEquipment
		if err := equipment.Scan(&item.ID, &item.RoomID, &item.Name, &item.Description, &item.Quantity, &item.Status); err != nil {
			return err
		}
		room.FixedEquipment = append(room.FixedEquipment, item)
	}
	return nil
}
func scanRoom(row interface{ Scan(...any) error }) (Room, error) {
	var room Room
	var capacity sql.NullInt64
	if err := row.Scan(&room.ID, &room.Name, &room.Description, &capacity, &room.Status); err != nil {
		return Room{}, err
	}
	if capacity.Valid {
		room.Capacity = &capacity.Int64
	}
	return room, nil
}
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
