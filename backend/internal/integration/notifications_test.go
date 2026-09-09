package integration

import (
	"net/http"
	"testing"
)

func TestBatchReadNotificationsIsolationAndSnapshot(t *testing.T) {
	app := newE2EApp(t)
	endpoint := app.server.URL + "/api/notifications/read-all"
	if status := jsonRequest(t, client(t), http.MethodPatch, endpoint, map[string]any{"throughId": 100}, nil); status != http.StatusUnauthorized {
		t.Fatalf("anonymous batch status = %d", status)
	}
	one, two := client(t), client(t)
	idOne := registerCustomer(t, app, one, "一", "notify-one@bandroom.test")
	idTwo := registerCustomer(t, app, two, "二", "notify-two@bandroom.test")
	login(t, app, one, "notify-one@bandroom.test")
	if _, err := app.db.Exec(`INSERT INTO notifications(id, user_id, notification_type, title, body, read_at) VALUES
		(1, ?, 'booking_created', '旧已读', '', '2020-01-01T00:00:00Z'),
		(2, ?, 'booking_created', '已加载', '', NULL),
		(3, ?, 'booking_created', '他人', '', NULL),
		(4, ?, 'booking_created', '新到达', '', NULL)`, idOne, idOne, idTwo, idOne); err != nil {
		t.Fatal(err)
	}
	for _, input := range []any{map[string]any{}, map[string]any{"throughId": -1}, map[string]any{"throughId": "bad"}} {
		if status := jsonRequest(t, one, http.MethodPatch, endpoint, input, nil); status != http.StatusBadRequest {
			t.Fatalf("invalid input status = %d", status)
		}
	}
	for _, expected := range []int64{1, 0} {
		var response struct {
			Updated int64 `json:"updated"`
		}
		// A forged userId must never replace the authenticated recipient.
		status := jsonRequest(t, one, http.MethodPatch, endpoint, map[string]any{"throughId": 3, "userId": idTwo}, &response)
		if status != http.StatusOK || response.Updated != expected {
			t.Fatalf("status=%d updated=%d want=%d", status, response.Updated, expected)
		}
	}
	var oldRead string
	if err := app.db.QueryRow(`SELECT read_at FROM notifications WHERE id = 1`).Scan(&oldRead); err != nil || oldRead != "2020-01-01T00:00:00Z" {
		t.Fatalf("existing read timestamp changed: %s, %v", oldRead, err)
	}
	var unread int
	if err := app.db.QueryRow(`SELECT COUNT(*) FROM notifications WHERE id IN (3, 4) AND read_at IS NULL`).Scan(&unread); err != nil || unread != 2 {
		t.Fatalf("other user's notification or concurrent arrival modified: %d, %v", unread, err)
	}
}
