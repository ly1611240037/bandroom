package content

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
	mux.HandleFunc("GET /api/venue", h.list)
	mux.Handle("PUT /api/owner/venue/{key}", h.auth.RequireRole("owner", http.HandlerFunc(h.set)))
}
func (h Handler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		httpx.WriteError(w, 500, "读取场馆信息失败")
		return
	}
	writeJSON(w, 200, map[string]any{"content": items})
}
func (h Handler) set(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpx.WriteError(w, 400, "请求格式不正确")
		return
	}
	if err := h.service.Set(r.Context(), r.PathValue("key"), input.Value); err != nil {
		httpx.WriteError(w, 400, "场馆信息保存失败")
		return
	}
	writeJSON(w, 200, map[string]string{"message": "场馆信息已保存"})
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
