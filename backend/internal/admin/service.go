package admin

import (
	"context"
	"database/sql"
)

type Stats struct {
	TotalBookings        int `json:"totalBookings"`
	ConfirmedBookings    int `json:"confirmedBookings"`
	CompletedBookings    int `json:"completedBookings"`
	CancelledBookings    int `json:"cancelledBookings"`
	NoShowBookings       int `json:"noShowBookings"`
	ActiveCards          int `json:"activeCards"`
	MaintenanceEquipment int `json:"maintenanceEquipment"`
}
type Audit struct {
	ID          int64  `json:"id"`
	ActorUserID int64  `json:"actorUserId"`
	Action      string `json:"action"`
	TargetType  string `json:"targetType"`
	TargetID    *int64 `json:"targetId,omitempty"`
	AfterJSON   string `json:"afterJson,omitempty"`
	CreatedAt   string `json:"createdAt"`
}
type Service struct{ db *sql.DB }

func NewService(db *sql.DB) *Service { return &Service{db: db} }
func (s *Service) Stats(ctx context.Context) (Stats, error) {
	var result Stats
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(status = 'confirmed'), 0), COALESCE(SUM(status = 'completed'), 0), COALESCE(SUM(status = 'cancelled'), 0), COALESCE(SUM(status = 'no_show'), 0) FROM bookings`).Scan(&result.TotalBookings, &result.ConfirmedBookings, &result.CompletedBookings, &result.CancelledBookings, &result.NoShowBookings)
	if err != nil {
		return result, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM membership_cards WHERE status = 'active'`).Scan(&result.ActiveCards); err != nil {
		return result, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM public_equipment WHERE status = 'maintenance'`).Scan(&result.MaintenanceEquipment); err != nil {
		return result, err
	}
	return result, nil
}
func (s *Service) Audits(ctx context.Context) ([]Audit, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, actor_user_id, action, target_type, target_id, after_json, created_at FROM audit_logs ORDER BY id DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Audit, 0)
	for rows.Next() {
		var item Audit
		var target sql.NullInt64
		var after sql.NullString
		if err := rows.Scan(&item.ID, &item.ActorUserID, &item.Action, &item.TargetType, &target, &after, &item.CreatedAt); err != nil {
			return nil, err
		}
		if target.Valid {
			item.TargetID = &target.Int64
		}
		if after.Valid {
			item.AfterJSON = after.String
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
