package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ly1611240037/bandroom/backend/internal/validation"
)

var (
	ErrInvalidCredentials = errors.New("邮箱或密码不正确")
	ErrEmailNotVerified   = errors.New("请先完成邮箱验证")
	ErrInvalidToken       = errors.New("验证链接无效或已过期")
	ErrUnauthorized       = errors.New("请先登录")
	ErrForbidden          = errors.New("没有权限执行此操作")
)

type Mailer interface {
	SendVerification(ctx context.Context, email, name, link string) error
	SendPasswordReset(ctx context.Context, email, name, link string) error
}

type LogMailer struct {
	Logf func(format string, args ...any)
}

func (m LogMailer) SendVerification(_ context.Context, email, name, link string) error {
	if m.Logf != nil {
		m.Logf("演示邮件：发送给 %s(%s) 的邮箱验证链接：%s", name, email, link)
	}
	return nil
}

func (m LogMailer) SendPasswordReset(_ context.Context, email, name, link string) error {
	if m.Logf != nil {
		m.Logf("演示邮件：发送给 %s(%s) 的密码重置链接：%s", name, email, link)
	}
	return nil
}

type User struct {
	ID              int64  `json:"id"`
	Role            string `json:"role"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	EmailVerifiedAt string `json:"emailVerifiedAt,omitempty"`
}

type Service struct {
	db         *sql.DB
	mailer     Mailer
	appURL     string
	sessionTTL time.Duration
	verifyTTL  time.Duration
}

func NewService(database *sql.DB, appURL string, mailer Mailer) *Service {
	if mailer == nil {
		mailer = LogMailer{}
	}
	return &Service{db: database, mailer: mailer, appURL: strings.TrimRight(appURL, "/"), sessionTTL: 7 * 24 * time.Hour, verifyTTL: 24 * time.Hour}
}

func (s *Service) Register(ctx context.Context, name, email, phone, password string) (User, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	phone = strings.TrimSpace(phone)
	if name == "" || len(name) > 80 {
		return User{}, errors.New("姓名不能为空且不能超过 80 个字符")
	}
	if !validation.Email(email) {
		return User{}, errors.New("请输入有效的邮箱")
	}
	if phone == "" || len(phone) > 30 {
		return User{}, errors.New("手机号不能为空且不能超过 30 个字符")
	}
	if len(password) < 8 || len(password) > 72 {
		return User{}, errors.New("密码长度必须为 8 到 72 个字符")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}
	token, tokenHash, err := newToken()
	if err != nil {
		return User{}, err
	}
	expires := time.Now().UTC().Add(s.verifyTTL)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `INSERT INTO users(role, name, email, phone, password_hash) VALUES ('customer', ?, ?, ?, ?)`, name, email, phone, hash)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return User{}, errors.New("该邮箱已经注册")
		}
		return User{}, fmt.Errorf("create customer: %w", err)
	}
	userID, err := result.LastInsertId()
	if err != nil {
		return User{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO email_verification_tokens(user_id, token_hash, expires_at) VALUES (?, ?, ?)`, userID, tokenHash, expires.Format(time.RFC3339)); err != nil {
		return User{}, fmt.Errorf("create verification token: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return User{}, err
	}
	link := s.appURL + "/verify-email?token=" + url.QueryEscape(token)
	if err := s.mailer.SendVerification(ctx, email, name, link); err != nil {
		return User{}, fmt.Errorf("send verification email: %w", err)
	}
	return User{ID: userID, Role: "customer", Name: name, Email: email, Phone: phone}, nil
}

func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	hash := hashToken(token)
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := s.db.ExecContext(ctx, `UPDATE users SET email_verified_at = ? WHERE id = (
		SELECT user_id FROM email_verification_tokens WHERE token_hash = ? AND used_at IS NULL AND expires_at > ?
	)`, now, hash, now)
	if err != nil {
		return err
	}
	if count, _ := result.RowsAffected(); count != 1 {
		return ErrInvalidToken
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE email_verification_tokens SET used_at = ? WHERE token_hash = ?`, now, hash); err != nil {
		return err
	}
	return nil
}

func (s *Service) Login(ctx context.Context, email, password string) (User, string, error) {
	var user User
	var hash, verified sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT id, role, name, email, phone, password_hash, email_verified_at FROM users WHERE email = ?`, strings.ToLower(strings.TrimSpace(email))).Scan(&user.ID, &user.Role, &user.Name, &user.Email, &user.Phone, &hash, &verified)
	if err != nil || !CheckPassword(hash.String, password) {
		return User{}, "", ErrInvalidCredentials
	}
	if user.Role == "customer" && !verified.Valid {
		return User{}, "", ErrEmailNotVerified
	}
	if verified.Valid {
		user.EmailVerifiedAt = verified.String
	}
	token, tokenHash, err := newToken()
	if err != nil {
		return User{}, "", err
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO sessions(token_hash, user_id, expires_at) VALUES (?, ?, ?)`, tokenHash, user.ID, time.Now().UTC().Add(s.sessionTTL).Format(time.RFC3339)); err != nil {
		return User{}, "", err
	}
	return user, token, nil
}

func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	var user User
	if err := s.db.QueryRowContext(ctx, `SELECT id, name, email FROM users WHERE email = ?`, email).Scan(&user.ID, &user.Name, &user.Email); err != nil {
		// Password recovery should not reveal whether an email exists.
		return nil
	}
	token, tokenHash, err := newToken()
	if err != nil {
		return err
	}
	expires := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	if _, err := s.db.ExecContext(ctx, `INSERT INTO password_reset_tokens(user_id, token_hash, expires_at) VALUES (?, ?, ?)`, user.ID, tokenHash, expires); err != nil {
		return err
	}
	link := s.appURL + "/reset-password?token=" + url.QueryEscape(token)
	return s.mailer.SendPasswordReset(ctx, user.Email, user.Name, link)
}

func (s *Service) ResetPassword(ctx context.Context, token, password string) error {
	if len(password) < 8 || len(password) > 72 {
		return errors.New("密码长度必须为 8 到 72 个字符")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var userID int64
	if err := tx.QueryRowContext(ctx, `SELECT user_id FROM password_reset_tokens WHERE token_hash = ? AND used_at IS NULL AND expires_at > ?`, hashToken(token), now).Scan(&userID); err != nil {
		return ErrInvalidToken
	}
	if _, err := tx.ExecContext(ctx, `UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`, hash, now, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE password_reset_tokens SET used_at = ? WHERE token_hash = ?`, now, hashToken(token)); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sessions SET revoked_at = ? WHERE user_id = ? AND revoked_at IS NULL`, now, userID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) CurrentUser(ctx context.Context, token string) (User, error) {
	var user User
	var verified sql.NullString
	now := time.Now().UTC().Format(time.RFC3339)
	err := s.db.QueryRowContext(ctx, `SELECT u.id, u.role, u.name, u.email, u.phone, u.email_verified_at
		FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = ? AND s.revoked_at IS NULL AND s.expires_at > ?`, hashToken(token), now).
		Scan(&user.ID, &user.Role, &user.Name, &user.Email, &user.Phone, &verified)
	if err != nil {
		return User{}, ErrUnauthorized
	}
	if verified.Valid {
		user.EmailVerifiedAt = verified.String
	}
	return user, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `UPDATE sessions SET revoked_at = ? WHERE token_hash = ? AND revoked_at IS NULL`, time.Now().UTC().Format(time.RFC3339), hashToken(token))
	return err
}

func (s *Service) IsVerifiedCustomer(user User) error {
	if user.Role != "customer" {
		return nil
	}
	if user.EmailVerifiedAt == "" {
		return ErrEmailNotVerified
	}
	return nil
}

func newToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(bytes)
	return token, hashToken(token), nil
}

func hashToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}
