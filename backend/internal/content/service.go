package content

import (
	"context"
	"database/sql"
	"strings"
)

type Item struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type Service struct{ db *sql.DB }

func NewService(db *sql.DB) *Service { return &Service{db: db} }
func (s *Service) List(ctx context.Context) ([]Item, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT content_key, content_value FROM venue_content ORDER BY content_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.Key, &item.Value); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (s *Service) Set(ctx context.Context, key, value string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return sql.ErrNoRows
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO venue_content(content_key, content_value) VALUES (?, ?) ON CONFLICT(content_key) DO UPDATE SET content_value = excluded.content_value, updated_at = CURRENT_TIMESTAMP`, key, value)
	return err
}
