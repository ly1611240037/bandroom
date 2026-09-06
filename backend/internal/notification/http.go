package notification

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ly1611240037/bandroom/backend/internal/auth"
	"github.com/ly1611240037/bandroom/backend/internal/httpx"
)

type Handler struct {
	service *Service
	auth    auth.Handler
}

func NewHandler(service *Service, authHandler auth.Handler) Handler {
	return Handler{service: service, auth: authHandler}
}
func (h Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/notifications", h.auth.RequireLogin(http.HandlerFunc(h.list)))
	mux.Handle("GET /api/notifications/unread-count", h.auth.RequireLogin(http.HandlerFunc(h.unread)))
	mux.Handle("PATCH /api/notifications/{id}/read", h.auth.RequireLogin(http.HandlerFunc(h.markRead)))
}
func (h Handler) list(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, 401, "请先登录")
		return
	}
	items, err := h.service.List(r.Context(), user.ID, r.URL.Query().Get("unread") == "true")
	if err != nil {
		httpx.WriteError(w, 500, "读取通知失败")
		return
	}
	writeJSON(w, 200, map[string]any{"notifications": items})
}
func (h Handler) unread(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, 401, "请先登录")
		return
	}
	count, err := h.service.UnreadCount(r.Context(), user.ID)
	if err != nil {
		httpx.WriteError(w, 500, "读取未读数量失败")
		return
	}
	writeJSON(w, 200, map[string]int{"unread": count})
}
func (h Handler) markRead(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, 401, "请先登录")
		return
	}
	var id int64
	if _, err := fmt.Sscan(r.PathValue("id"), &id); err != nil {
		httpx.WriteError(w, 400, "通知编号不正确")
		return
	}
	if err := h.service.MarkRead(r.Context(), user.ID, id); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"message": "通知已读"})
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
