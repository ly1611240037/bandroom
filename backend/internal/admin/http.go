package admin

import (
	"encoding/json"
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
	mux.Handle("GET /api/owner/stats", h.auth.RequireRole("owner", http.HandlerFunc(h.stats)))
	mux.Handle("GET /api/owner/audits", h.auth.RequireRole("owner", http.HandlerFunc(h.audits)))
}
func (h Handler) stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.Stats(r.Context())
	if err != nil {
		httpx.WriteError(w, 500, "读取统计失败")
		return
	}
	writeJSON(w, 200, map[string]any{"stats": stats})
}
func (h Handler) audits(w http.ResponseWriter, r *http.Request) {
	audits, err := h.service.Audits(r.Context())
	if err != nil {
		httpx.WriteError(w, 500, "读取审计日志失败")
		return
	}
	writeJSON(w, 200, map[string]any{"audits": audits})
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
