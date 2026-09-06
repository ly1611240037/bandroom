package booking

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
	mux.HandleFunc("GET /api/availability", h.availability)
	mux.Handle("POST /api/customer/bookings", h.auth.RequireVerifiedCustomer(http.HandlerFunc(h.create)))
	mux.Handle("GET /api/customer/bookings", h.auth.RequireRole("customer", http.HandlerFunc(h.list)))
}
func (h Handler) availability(w http.ResponseWriter, r *http.Request) {
	roomID, err := parseID(r.URL.Query().Get("roomId"))
	if err != nil {
		httpx.WriteError(w, 400, "房间编号不正确")
		return
	}
	slots, err := h.service.Availability(r.Context(), roomID, r.URL.Query().Get("date"))
	if err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"slots": slots})
}
func (h Handler) create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RoomID    int64              `json:"roomId"`
		BandName  string             `json:"bandName"`
		Phone     string             `json:"phone"`
		StartsAt  string             `json:"startsAt"`
		EndsAt    string             `json:"endsAt"`
		Notes     string             `json:"notes"`
		Equipment []EquipmentRequest `json:"equipment"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, 401, "请先登录")
		return
	}
	item, err := h.service.Create(r.Context(), user.ID, input.RoomID, input.BandName, input.Phone, input.StartsAt, input.EndsAt, input.Notes, input.Equipment)
	if err != nil {
		status := 400
		if err == ErrConflict {
			status = 409
		}
		httpx.WriteError(w, status, err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{"booking": item})
}
func (h Handler) list(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, 401, "请先登录")
		return
	}
	items, err := h.service.ListByUser(r.Context(), user.ID)
	if err != nil {
		httpx.WriteError(w, 500, "读取预约失败")
		return
	}
	writeJSON(w, 200, map[string]any{"bookings": items})
}
func parseID(value string) (int64, error) {
	var id int64
	_, err := fmt.Sscan(value, &id)
	return id, err
}
func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		httpx.WriteError(w, 400, "请求格式不正确")
		return false
	}
	return true
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
