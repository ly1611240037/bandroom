package repository

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func New(database *sql.DB) *Repository {
	return &Repository{db: database}
}

type Room struct {
	ID          int64
	Name        string
	Description string
	Capacity    sql.NullInt64
	Status      string
}

func (r *Repository) CreateRoom(ctx context.Context, name, description string, capacity *int64) (int64, error) {
	result, err := r.db.ExecContext(ctx, `INSERT INTO rooms(name, description, capacity) VALUES (?, ?, ?)`, name, description, capacity)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) GetRoom(ctx context.Context, id int64) (Room, error) {
	var room Room
	err := r.db.QueryRowContext(ctx, `SELECT id, name, description, capacity, status FROM rooms WHERE id = ?`, id).
		Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.Status)
	return room, err
}

func (r *Repository) UpdateRoomStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE rooms SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, status, id)
	return err
}

type PublicEquipment struct {
	ID       int64
	Name     string
	Quantity int
	Status   string
}

func (r *Repository) CreatePublicEquipment(ctx context.Context, name string, quantity int) (int64, error) {
	result, err := r.db.ExecContext(ctx, `INSERT INTO public_equipment(name, quantity) VALUES (?, ?)`, name, quantity)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) GetPublicEquipment(ctx context.Context, id int64) (PublicEquipment, error) {
	var equipment PublicEquipment
	err := r.db.QueryRowContext(ctx, `SELECT id, name, quantity, status FROM public_equipment WHERE id = ?`, id).
		Scan(&equipment.ID, &equipment.Name, &equipment.Quantity, &equipment.Status)
	return equipment, err
}

func (r *Repository) UpdatePublicEquipmentStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE public_equipment SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, status, id)
	return err
}
