package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/ly1611240037/bandroom/backend/internal/admin"
	"github.com/ly1611240037/bandroom/backend/internal/auth"
	"github.com/ly1611240037/bandroom/backend/internal/booking"
	"github.com/ly1611240037/bandroom/backend/internal/content"
	"github.com/ly1611240037/bandroom/backend/internal/db"
	"github.com/ly1611240037/bandroom/backend/internal/membership"
	"github.com/ly1611240037/bandroom/backend/internal/notification"
	"github.com/ly1611240037/bandroom/backend/internal/rooms"
)

type e2eMailer struct{ verificationLink string }

func (m *e2eMailer) SendVerification(_ context.Context, _, _, link string) error {
	m.verificationLink = link
	return nil
}
func (m *e2eMailer) SendPasswordReset(context.Context, string, string, string) error { return nil }

type e2eApp struct {
	db      *sqlDB
	server  *httptest.Server
	mailer  *e2eMailer
	booking *booking.Service
}

// sqlDB keeps the test setup type small while still allowing direct lifecycle assertions.
type sqlDB = sql.DB

func newE2EApp(t *testing.T) *e2eApp {
	t.Helper()
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "e2e.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(context.Background(), database); err != nil {
		database.Close()
		t.Fatal(err)
	}
	mailer := &e2eMailer{}
	authHandler := auth.NewHandler(auth.NewService(database, "http://test.local", mailer))
	notifications := notification.NewService(database, notification.LogSender{})
	bookingService := booking.NewService(database, notifications)
	mux := http.NewServeMux()
	authHandler.RegisterRoutes(mux)
	membership.NewHandler(membership.NewService(database), authHandler).RegisterRoutes(mux)
	rooms.NewHandler(rooms.NewService(database), authHandler).RegisterRoutes(mux)
	booking.NewHandler(bookingService, authHandler).RegisterRoutes(mux)
	notification.NewHandler(notifications, authHandler).RegisterRoutes(mux)
	content.NewHandler(content.NewService(database), authHandler).RegisterRoutes(mux)
	admin.NewHandler(admin.NewService(database), authHandler).RegisterRoutes(mux)
	server := httptest.NewServer(mux)
	t.Cleanup(func() { server.Close(); database.Close() })
	return &e2eApp{db: database, server: server, mailer: mailer, booking: bookingService}
}

func jsonRequest(t *testing.T, client *http.Client, method, endpoint string, input any, output any) int {
	t.Helper()
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		t.Fatal(err)
	}
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if output != nil {
		if err := json.NewDecoder(res.Body).Decode(output); err != nil {
			t.Fatal(err)
		}
	} else {
		_, _ = io.Copy(io.Discard, res.Body)
	}
	return res.StatusCode
}

func client(t *testing.T) *http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Client{Jar: jar}
}

func registerCustomer(t *testing.T, app *e2eApp, c *http.Client, name, email string) int64 {
	var response struct {
		User struct {
			ID int64 `json:"id"`
		} `json:"user"`
	}
	status := jsonRequest(t, c, http.MethodPost, app.server.URL+"/api/auth/register", map[string]any{
		"name": name, "email": email, "phone": "13800000000", "password": "password123",
	}, &response)
	if status != http.StatusCreated {
		t.Fatalf("registration status = %d", status)
	}
	parsed, err := url.Parse(app.mailer.verificationLink)
	if err != nil {
		t.Fatal(err)
	}
	if status := jsonRequest(t, c, http.MethodGet, app.server.URL+"/api/auth/verify?token="+url.QueryEscape(parsed.Query().Get("token")), nil, &map[string]any{}); status != http.StatusOK {
		t.Fatalf("verification status = %d", status)
	}
	return response.User.ID
}

func login(t *testing.T, app *e2eApp, c *http.Client, email string) {
	status := jsonRequest(t, c, http.MethodPost, app.server.URL+"/api/auth/login", map[string]any{
		"email": email, "password": "password123",
	}, &map[string]any{})
	if status != http.StatusOK {
		t.Fatalf("login %s status = %d", email, status)
	}
}

func TestBookingManagementEndToEnd(t *testing.T) {
	app := newE2EApp(t)
	passwordHash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.db.Exec(`INSERT INTO users(role, name, email, phone, password_hash, email_verified_at) VALUES ('owner', '老板', 'owner@bandroom.test', '13900000000', ?, CURRENT_TIMESTAMP)`, passwordHash); err != nil {
		t.Fatal(err)
	}

	owner := client(t)
	login(t, app, owner, "owner@bandroom.test")
	var planResponse struct {
		Plan struct {
			ID int64 `json:"id"`
		} `json:"plan"`
	}
	if status := jsonRequest(t, owner, http.MethodPost, app.server.URL+"/api/owner/membership/plans", map[string]any{
		"planType": "count", "name": "演示次数卡", "priceCents": 1000, "includedUses": 2,
	}, &planResponse); status != http.StatusCreated {
		t.Fatalf("create plan status = %d", status)
	}
	var roomResponse struct {
		Room struct {
			ID int64 `json:"id"`
		} `json:"room"`
	}
	for _, name := range []string{"A 房", "B 房"} {
		if status := jsonRequest(t, owner, http.MethodPost, app.server.URL+"/api/owner/rooms", map[string]any{"name": name, "description": "集成测试房间", "capacity": 6}, &roomResponse); status != http.StatusCreated {
			t.Fatalf("create room status = %d", status)
		}
	}
	roomA, roomB := int64(1), int64(2)
	var equipmentResponse struct {
		Equipment struct {
			ID int64 `json:"id"`
		} `json:"equipment"`
	}
	if status := jsonRequest(t, owner, http.MethodPost, app.server.URL+"/api/owner/equipment", map[string]any{"name": "测试麦克风", "description": "一件库存", "quantity": 1}, &equipmentResponse); status != http.StatusCreated {
		t.Fatalf("create equipment status = %d", status)
	}

	customerOne, customerTwo := client(t), client(t)
	idOne := registerCustomer(t, app, customerOne, "顾客一", "one@example.com")
	idTwo := registerCustomer(t, app, customerTwo, "顾客二", "two@example.com")
	for _, userID := range []int64{idOne, idTwo} {
		status := jsonRequest(t, owner, http.MethodPost, app.server.URL+"/api/owner/membership/cards", map[string]any{
			"userId": userID, "planId": planResponse.Plan.ID, "paidAmountCents": 1000, "paymentMethod": "线下支付",
		}, &map[string]any{})
		if status != http.StatusCreated {
			t.Fatalf("activate card for %d status = %d", userID, status)
		}
	}
	login(t, app, customerOne, "one@example.com")
	login(t, app, customerTwo, "two@example.com")

	zone := time.FixedZone("CST", 8*60*60)
	start := time.Date(time.Now().In(zone).Year(), time.Now().In(zone).Month(), time.Now().In(zone).Day()+1, 10, 0, 0, 0, zone)
	end := start.Add(time.Hour)
	bookingInput := func(roomID int64) map[string]any {
		return map[string]any{"roomId": roomID, "bandName": "集成测试乐队", "phone": "13800000000", "startsAt": start.Format(time.RFC3339), "endsAt": end.Format(time.RFC3339), "notes": "e2e", "equipment": []map[string]any{{"equipmentId": equipmentResponse.Equipment.ID, "quantity": 1}}}
	}
	var first struct {
		Booking struct {
			ID int64 `json:"id"`
		} `json:"booking"`
	}
	if status := jsonRequest(t, customerOne, http.MethodPost, app.server.URL+"/api/customer/bookings", bookingInput(roomA), &first); status != http.StatusCreated {
		t.Fatalf("valid booking status = %d", status)
	}
	if status := jsonRequest(t, customerTwo, http.MethodPost, app.server.URL+"/api/customer/bookings", bookingInput(roomA), &map[string]any{}); status != http.StatusConflict {
		t.Fatalf("room conflict status = %d", status)
	}
	if status := jsonRequest(t, customerOne, http.MethodPost, app.server.URL+fmt.Sprintf("/api/customer/bookings/%d/cancel", first.Booking.ID), map[string]any{"reason": "改期"}, &map[string]any{}); status != http.StatusOK {
		t.Fatalf("cancel status = %d", status)
	}
	var second struct {
		Booking struct {
			ID int64 `json:"id"`
		} `json:"booking"`
	}
	if status := jsonRequest(t, customerTwo, http.MethodPost, app.server.URL+"/api/customer/bookings", bookingInput(roomB), &second); status != http.StatusCreated {
		t.Fatalf("second booking status = %d", status)
	}
	if status := jsonRequest(t, customerOne, http.MethodPost, app.server.URL+"/api/customer/bookings", bookingInput(roomB), &map[string]any{}); status != http.StatusConflict {
		t.Fatalf("equipment conflict status = %d", status)
	}
	var remaining int
	if err := app.db.QueryRow(`SELECT remaining_uses FROM membership_cards WHERE user_id = ?`, idOne).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 2 {
		t.Fatalf("cancel should restore count card, remaining = %d", remaining)
	}
	past := time.Now().UTC().Add(-time.Hour)
	if _, err := app.db.Exec(`UPDATE bookings SET starts_at = ?, ends_at = ?, occupied_until = ? WHERE id = ?`, past.Add(-time.Hour).Format(time.RFC3339), past.Add(-30*time.Minute).Format(time.RFC3339), past.Format(time.RFC3339), second.Booking.ID); err != nil {
		t.Fatal(err)
	}
	if count, err := app.booking.CompleteDue(context.Background(), time.Now().Add(time.Hour)); err != nil || count != 1 {
		t.Fatalf("automatic completion count=%d err=%v", count, err)
	}
	if status := jsonRequest(t, owner, http.MethodGet, app.server.URL+"/api/owner/stats", nil, &map[string]any{}); status != http.StatusOK {
		t.Fatalf("owner stats status = %d", status)
	}
	if status := jsonRequest(t, owner, http.MethodPut, app.server.URL+"/api/owner/venue/announcement", map[string]any{"value": "测试公告"}, &map[string]any{}); status != http.StatusOK {
		t.Fatalf("owner content status = %d", status)
	}
}
