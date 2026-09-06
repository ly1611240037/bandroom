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
	mux.Handle("GET /api/owner/bookings", h.auth.RequireRole("owner", http.HandlerFunc(h.listAll)))
	mux.Handle("POST /api/customer/bookings/{id}/cancel", h.auth.RequireRole("customer", http.HandlerFunc(h.cancelCustomer)))
	mux.Handle("POST /api/owner/bookings/{id}/cancel", h.auth.RequireRole("owner", http.HandlerFunc(h.cancelOwner)))
	mux.Handle("POST /api/owner/bookings/{id}/no-show", h.auth.RequireRole("owner", http.HandlerFunc(h.noShow)))
	mux.Handle("PATCH /api/owner/bookings/{id}", h.auth.RequireRole("owner", http.HandlerFunc(h.editOwner)))
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
func (h Handler) listAll(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListAll(r.Context())
	if err != nil {
		httpx.WriteError(w, 500, "读取全部预约失败")
		return
	}
	writeJSON(w, 200, map[string]any{"bookings": items})
}
func (h Handler) cancelCustomer(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, 401, "请先登录")
		return
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, 400, "预约编号不正确")
		return
	}
	var input struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&input)
	if err := h.service.CancelByCustomer(r.Context(), user.ID, id, input.Reason); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"message": "预约已取消"})
}
func (h Handler) cancelOwner(w http.ResponseWriter, r *http.Request) {
	actor, ok := auth.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, 401, "请先登录")
		return
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, 400, "预约编号不正确")
		return
	}
	var input struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&input)
	if err := h.service.CancelByOwner(r.Context(), actor.ID, id, input.Reason); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"message": "预约已取消"})
}
func (h Handler) noShow(w http.ResponseWriter, r *http.Request) {
	actor, ok := auth.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, 401, "请先登录")
		return
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, 400, "预约编号不正确")
		return
	}
	if err := h.service.MarkNoShow(r.Context(), actor.ID, id); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"message": "已标记为未到场"})
}
func (h Handler) editOwner(w http.ResponseWriter, r *http.Request) {
	actor, ok := auth.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, 401, "请先登录")
		return
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, 400, "预约编号不正确")
		return
	}
	var input struct {
		StartsAt string `json:"startsAt"`
		EndsAt   string `json:"endsAt"`
		Notes    string `json:"notes"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := h.service.OwnerEdit(r.Context(), actor.ID, id, input.StartsAt, input.EndsAt, input.Notes); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"message": "预约已修改"})
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
