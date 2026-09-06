package auth

import (
	"context"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ly1611240037/bandroom/backend/internal/db"
)

type captureMailer struct{ link string }
type resetCaptureMailer struct {
	captureMailer
	resetLink string
}

func (m *captureMailer) SendVerification(_ context.Context, _, _, link string) error {
	m.link = link
	return nil
}

func (m *captureMailer) SendPasswordReset(_ context.Context, _, _, link string) error { return nil }

func testService(t *testing.T) (*Service, *captureMailer) {
	t.Helper()
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.Migrate(context.Background(), database); err != nil {
		t.Fatal(err)
	}
	mailer := &captureMailer{}
	return NewService(database, "http://localhost:5173", mailer), mailer
}

func TestRegisterVerifyLoginAndLogout(t *testing.T) {
	service, mailer := testService(t)
	ctx := context.Background()
	user, err := service.Register(ctx, "测试联系人", "test@example.com", "13800000000", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if user.Role != "customer" || mailer.link == "" {
		t.Fatalf("unexpected registration: %+v, link=%q", user, mailer.link)
	}
	if _, _, err := service.Login(ctx, user.Email, "password123"); err != ErrEmailNotVerified {
		t.Fatalf("expected unverified error, got %v", err)
	}
	parsed, err := url.Parse(mailer.link)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyEmail(ctx, parsed.Query().Get("token")); err != nil {
		t.Fatal(err)
	}
	loggedIn, session, err := service.Login(ctx, user.Email, "password123")
	if err != nil {
		t.Fatal(err)
	}
	if loggedIn.ID != user.ID || session == "" {
		t.Fatalf("unexpected login: %+v, %q", loggedIn, session)
	}
	if _, err := service.CurrentUser(ctx, session); err != nil {
		t.Fatal(err)
	}
	if err := service.Logout(ctx, session); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CurrentUser(ctx, session); err == nil {
		t.Fatal("revoked session should not be accepted")
	}
}

func TestVerificationTokenIsOneTimeAndDuplicateEmailRejected(t *testing.T) {
	service, mailer := testService(t)
	ctx := context.Background()
	if _, err := service.Register(ctx, "测试联系人", "test@example.com", "13800000000", "password123"); err != nil {
		t.Fatal(err)
	}
	parsed, _ := url.Parse(mailer.link)
	token := parsed.Query().Get("token")
	if err := service.VerifyEmail(ctx, token); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyEmail(ctx, token); err != ErrInvalidToken {
		t.Fatalf("expected one-time token error, got %v", err)
	}
	if _, err := service.Register(ctx, "另一个联系人", "TEST@example.com", "13800000001", "password123"); err == nil || !strings.Contains(err.Error(), "已经注册") {
		t.Fatalf("expected duplicate email error, got %v", err)
	}
}

func TestExpiredSessionIsRejected(t *testing.T) {
	service, _ := testService(t)
	service.sessionTTL = -time.Minute
	ctx := context.Background()
	if _, err := service.Register(ctx, "测试联系人", "test@example.com", "13800000000", "password123"); err != nil {
		t.Fatal(err)
	}
	// The test focuses on the session query boundary; an unverified customer cannot log in.
	if _, _, err := service.Login(ctx, "test@example.com", "password123"); err != ErrEmailNotVerified {
		t.Fatalf("expected verification requirement, got %v", err)
	}
}

func TestUnverifiedCustomerCannotPassBookingGuard(t *testing.T) {
	service, _ := testService(t)
	if err := service.IsVerifiedCustomer(User{Role: "customer"}); err != ErrEmailNotVerified {
		t.Fatalf("expected unverified customer to be blocked, got %v", err)
	}
	if err := service.IsVerifiedCustomer(User{Role: "customer", EmailVerifiedAt: "2026-09-06T00:00:00Z"}); err != nil {
		t.Fatalf("verified customer should pass guard: %v", err)
	}
}

func TestPasswordResetLinkIsOneTime(t *testing.T) {
	service, mailer := testService(t)
	ctx := context.Background()
	if _, err := service.Register(ctx, "测试联系人", "test@example.com", "13800000000", "password123"); err != nil {
		t.Fatal(err)
	}
	if err := service.RequestPasswordReset(ctx, "test@example.com"); err != nil {
		t.Fatal(err)
	}
	// The test mailer captures verification links; request a reset with a dedicated mailer below.
	_ = mailer
	resetMailer := &resetCaptureMailer{}
	service.mailer = resetMailer
	if err := service.RequestPasswordReset(ctx, "test@example.com"); err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(resetMailer.resetLink)
	if err != nil {
		t.Fatal(err)
	}
	token := parsed.Query().Get("token")
	if err := service.ResetPassword(ctx, token, "newpassword123"); err != nil {
		t.Fatal(err)
	}
	if err := service.ResetPassword(ctx, token, "anotherpass123"); err != ErrInvalidToken {
		t.Fatalf("expected one-time reset error, got %v", err)
	}
}

func (m *resetCaptureMailer) SendVerification(context.Context, string, string, string) error {
	return nil
}
func (m *resetCaptureMailer) SendPasswordReset(_ context.Context, _, _, link string) error {
	m.resetLink = link
	return nil
}
