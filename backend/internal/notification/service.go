package notification

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Item struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"userId"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	ReadAt      string `json:"readAt,omitempty"`
	EmailStatus string `json:"emailStatus"`
	CreatedAt   string `json:"createdAt"`
}
type Email struct {
	To      string
	Subject string
	Text    string
}
type Sender interface {
	Send(context.Context, Email) error
}
type LogSender struct {
	Logf func(format string, args ...any)
}

func (s LogSender) Send(_ context.Context, email Email) error {
	if s.Logf != nil {
		s.Logf("演示邮件：收件人=%s 主题=%s 内容=%s", email.To, email.Subject, email.Text)
	}
	return nil
}

type Service struct {
	db     *sql.DB
	sender Sender
}

func NewService(db *sql.DB, sender Sender) *Service {
	if sender == nil {
		sender = LogSender{}
	}
	return &Service{db: db, sender: sender}
}

func (s *Service) List(ctx context.Context, userID int64, unreadOnly bool) ([]Item, error) {
	query := `SELECT id, user_id, notification_type, title, body, read_at, email_status, created_at FROM notifications WHERE user_id = ?`
	if unreadOnly {
		query += ` AND read_at IS NULL`
	}
	query += ` ORDER BY created_at DESC, id DESC`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Item, 0)
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (s *Service) UnreadCount(ctx context.Context, userID int64) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = ? AND read_at IS NULL`, userID).Scan(&count)
	return count, err
}
func (s *Service) MarkRead(ctx context.Context, userID, notificationID int64) error {
	result, err := s.db.ExecContext(ctx, `UPDATE notifications SET read_at = COALESCE(read_at, ?) WHERE id = ? AND user_id = ?`, time.Now().UTC().Format(time.RFC3339), notificationID, userID)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count != 1 {
		return errors.New("通知不存在")
	}
	return nil
}
func (s *Service) MarkAllRead(ctx context.Context, userID, throughID int64) (int64, error) {
	if throughID <= 0 {
		return 0, errors.New("通知范围不正确")
	}
	result, err := s.db.ExecContext(ctx, `UPDATE notifications SET read_at = ? WHERE user_id = ? AND id <= ? AND read_at IS NULL`, time.Now().UTC().Format(time.RFC3339), userID, throughID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
func (s *Service) ProcessEmailQueue(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `SELECT n.id, u.email, u.name, n.notification_type, n.title, n.body FROM notifications n JOIN users u ON u.id = n.user_id WHERE n.email_status = 'pending' ORDER BY n.id LIMIT 50`)
	if err != nil {
		return err
	}
	defer rows.Close()
	type queued struct {
		id                             int64
		email, name, kind, title, body string
	}
	queue := make([]queued, 0)
	for rows.Next() {
		var item queued
		if err := rows.Scan(&item.id, &item.email, &item.name, &item.kind, &item.title, &item.body); err != nil {
			return err
		}
		queue = append(queue, item)
	}
	for _, item := range queue {
		err := s.sender.Send(ctx, Email{To: item.email, Subject: item.title, Text: fmt.Sprintf("%s：%s", item.name, item.body)})
		status := "sent"
		if err != nil {
			status = "failed"
		}
		if _, updateErr := s.db.ExecContext(ctx, `UPDATE notifications SET email_status = ? WHERE id = ? AND email_status = 'pending'`, status, item.id); updateErr != nil {
			return updateErr
		}
	}
	return nil
}

func (s *Service) CreateMembershipExpiryReminders(ctx context.Context, now time.Time) (int64, error) {
	from := now.UTC().Add(72 * time.Hour).Add(-30 * time.Minute).Format(time.RFC3339)
	to := now.UTC().Add(72 * time.Hour).Add(30 * time.Minute).Format(time.RFC3339)
	rows, err := s.db.QueryContext(ctx, `SELECT c.id, c.user_id, c.ends_at FROM membership_cards c JOIN membership_plans p ON p.id = c.plan_id WHERE p.plan_type = 'monthly' AND c.status IN ('active', 'scheduled') AND datetime(c.ends_at) BETWEEN datetime(?) AND datetime(?)`, from, to)
	if err != nil {
		return 0, err
	}
	type reminder struct {
		cardID, userID int64
		ends           string
	}
	reminders := make([]reminder, 0)
	for rows.Next() {
		var item reminder
		if err := rows.Scan(&item.cardID, &item.userID, &item.ends); err != nil {
			rows.Close()
			return 0, err
		}
		reminders = append(reminders, item)
	}
	rows.Close()
	var created int64
	for _, item := range reminders {
		body := fmt.Sprintf("会员卡 %d 将于 %s 到期，请及时续卡。", item.cardID, item.ends)
		result, err := s.db.ExecContext(ctx, `INSERT INTO notifications(user_id, notification_type, title, body, email_status) SELECT ?, 'membership_expiring', '会员卡即将到期', ?, 'pending' WHERE NOT EXISTS (SELECT 1 FROM notifications WHERE user_id = ? AND notification_type = 'membership_expiring' AND body = ?)`, item.userID, body, item.userID, body)
		if err != nil {
			return created, err
		}
		count, _ := result.RowsAffected()
		created += count
	}
	return created, nil
}
func (s *Service) CreateBookingNotifications(ctx context.Context, bookingID, customerID, roomID int64, startsAt, bandName string) error {
	body := fmt.Sprintf("预约 %s，房间 %d，开始时间 %s。", bandName, roomID, startsAt)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO notifications(user_id, notification_type, title, body, email_status) VALUES (?, 'booking_created', '预约成功', ?, 'pending')`, customerID, body); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id FROM users WHERE role = 'owner'`)
	if err != nil {
		return err
	}
	owners := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		owners = append(owners, id)
	}
	rows.Close()
	for _, ownerID := range owners {
		if _, err := tx.ExecContext(ctx, `INSERT INTO notifications(user_id, notification_type, title, body, email_status) VALUES (?, 'owner_booking_alert', '收到新预约', ?, 'pending')`, ownerID, body); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func scan(row interface{ Scan(...any) error }) (Item, error) {
	var item Item
	var read sql.NullString
	err := row.Scan(&item.ID, &item.UserID, &item.Type, &item.Title, &item.Body, &read, &item.EmailStatus, &item.CreatedAt)
	if read.Valid {
		item.ReadAt = read.String
	}
	return item, err
}
func RenderBookingEmail(item Item) Email {
	subject := item.Title
	if strings.TrimSpace(subject) == "" {
		subject = "BandRoom 预约通知"
	}
	return Email{Subject: subject, Text: item.Body}
}
