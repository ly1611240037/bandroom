package rooms

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
	mux.HandleFunc("GET /api/rooms", h.listRooms)
	mux.HandleFunc("GET /api/rooms/{id}", h.getRoom)
	mux.HandleFunc("GET /api/equipment", h.listEquipment)
	mux.Handle("POST /api/owner/rooms", h.auth.RequireRole("owner", http.HandlerFunc(h.createRoom)))
	mux.Handle("PATCH /api/owner/rooms/{id}", h.auth.RequireRole("owner", http.HandlerFunc(h.updateRoom)))
	mux.Handle("POST /api/owner/rooms/{id}/photos", h.auth.RequireRole("owner", http.HandlerFunc(h.addPhoto)))
	mux.Handle("POST /api/owner/rooms/{id}/equipment", h.auth.RequireRole("owner", http.HandlerFunc(h.addFixedEquipment)))
	mux.Handle("PATCH /api/owner/fixed-equipment/{id}/status", h.auth.RequireRole("owner", http.HandlerFunc(h.setFixedStatus)))
	mux.Handle("GET /api/owner/equipment", h.auth.RequireRole("owner", http.HandlerFunc(h.listEquipment)))
	mux.Handle("POST /api/owner/equipment", h.auth.RequireRole("owner", http.HandlerFunc(h.createEquipment)))
	mux.Handle("PATCH /api/owner/equipment/{id}", h.auth.RequireRole("owner", http.HandlerFunc(h.updateEquipment)))
	mux.Handle("GET /api/owner/schedule", h.auth.RequireRole("owner", http.HandlerFunc(h.schedule)))
	mux.Handle("PUT /api/owner/schedule/{weekday}", h.auth.RequireRole("owner", http.HandlerFunc(h.setSchedule)))
	mux.Handle("GET /api/owner/closures", h.auth.RequireRole("owner", http.HandlerFunc(h.closures)))
	mux.Handle("POST /api/owner/closures", h.auth.RequireRole("owner", http.HandlerFunc(h.addClosure)))
	mux.Handle("POST /api/customer/issues", h.auth.RequireRole("customer", http.HandlerFunc(h.createIssue)))
	mux.Handle("GET /api/customer/issues", h.auth.RequireRole("customer", http.HandlerFunc(h.listMyIssues)))
	mux.Handle("GET /api/owner/issues", h.auth.RequireRole("owner", http.HandlerFunc(h.listIssues)))
	mux.Handle("PATCH /api/owner/issues/{id}", h.auth.RequireRole("owner", http.HandlerFunc(h.updateIssue)))
}
func (h Handler) listRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.service.ListRooms(r.Context())
	if err != nil {
		httpx.WriteError(w, 500, "读取房间失败")
		return
	}
	writeJSON(w, 200, map[string]any{"rooms": rooms})
}
func (h Handler) getRoom(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, 400, "房间编号不正确")
		return
	}
	room, err := h.service.GetRoom(r.Context(), id)
	if err != nil {
		httpx.WriteError(w, 404, "房间不存在")
		return
	}
	writeJSON(w, 200, map[string]any{"room": room})
}
func (h Handler) createRoom(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Capacity    *int64 `json:"capacity"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	room, err := h.service.CreateRoom(r.Context(), input.Name, input.Description, input.Capacity)
	if err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{"room": room})
}
func (h Handler) updateRoom(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Capacity    *int64 `json:"capacity"`
		Status      string `json:"status"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, 400, "房间编号不正确")
		return
	}
	if err := h.service.UpdateRoom(r.Context(), id, input.Name, input.Description, input.Capacity, input.Status); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"message": "房间已更新"})
}
func (h Handler) addPhoto(w http.ResponseWriter, r *http.Request) {
	var input struct {
		URL       string `json:"url"`
		SortOrder int    `json:"sortOrder"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, 400, "房间编号不正确")
		return
	}
	if err := h.service.AddPhoto(r.Context(), id, input.URL, input.SortOrder); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, map[string]string{"message": "照片已添加"})
}
func (h Handler) addFixedEquipment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Quantity    int    `json:"quantity"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	roomID, err := parseID(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, 400, "房间编号不正确")
		return
	}
	id, err := h.service.AddFixedEquipment(r.Context(), roomID, input.Name, input.Description, input.Quantity)
	if err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, map[string]int64{"id": id})
}
func (h Handler) setFixedStatus(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Status string `json:"status"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, 400, "设备编号不正确")
		return
	}
	if err := h.service.SetFixedEquipmentStatus(r.Context(), id, input.Status); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"message": "固定设备状态已更新"})
}
func (h Handler) listEquipment(w http.ResponseWriter, r *http.Request) {
	equipment, err := h.service.ListEquipment(r.Context())
	if err != nil {
		httpx.WriteError(w, 500, "读取公共设备失败")
		return
	}
	writeJSON(w, 200, map[string]any{"equipment": equipment})
}
func (h Handler) createEquipment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name             string `json:"name"`
		Description      string `json:"description"`
		Quantity         int    `json:"quantity"`
		HourlyPriceCents int64  `json:"hourlyPriceCents"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	item, err := h.service.CreateEquipment(r.Context(), input.Name, input.Description, input.Quantity, input.HourlyPriceCents)
	if err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{"equipment": item})
}
func (h Handler) updateEquipment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name             string `json:"name"`
		Description      string `json:"description"`
		Quantity         int    `json:"quantity"`
		HourlyPriceCents int64  `json:"hourlyPriceCents"`
		Status           string `json:"status"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, 400, "设备编号不正确")
		return
	}
	if err := h.service.UpdateEquipment(r.Context(), id, input.Name, input.Description, input.Quantity, input.HourlyPriceCents, input.Status); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"message": "公共设备已更新"})
}
func (h Handler) schedule(w http.ResponseWriter, r *http.Request) {
	hours, err := h.service.ListBusinessHours(r.Context())
	if err != nil {
		httpx.WriteError(w, 500, "读取营业时间失败")
		return
	}
	buffer, err := h.service.GetCleaningBuffer(r.Context())
	if err != nil {
		httpx.WriteError(w, 500, "读取清场时间失败")
		return
	}
	writeJSON(w, 200, map[string]any{"businessHours": hours, "cleaningBufferMinutes": buffer})
}
func (h Handler) setSchedule(w http.ResponseWriter, r *http.Request) {
	weekday, err := parseID(r.PathValue("weekday"))
	if err != nil {
		httpx.WriteError(w, 400, "星期编号不正确")
		return
	}
	var input struct {
		OpensAt  string `json:"opensAt"`
		ClosesAt string `json:"closesAt"`
		Enabled  bool   `json:"enabled"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := h.service.SetBusinessHour(r.Context(), BusinessHour{Weekday: int(weekday), OpensAt: input.OpensAt, ClosesAt: input.ClosesAt, Enabled: input.Enabled}); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"message": "营业时间已更新"})
}
func (h Handler) closures(w http.ResponseWriter, r *http.Request) {
	closures, err := h.service.ListClosures(r.Context())
	if err != nil {
		httpx.WriteError(w, 500, "读取闭店安排失败")
		return
	}
	writeJSON(w, 200, map[string]any{"closures": closures})
}
func (h Handler) addClosure(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RoomID   *int64 `json:"roomId"`
		StartsAt string `json:"startsAt"`
		EndsAt   string `json:"endsAt"`
		Reason   string `json:"reason"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	id, err := h.service.AddClosure(r.Context(), input.RoomID, input.StartsAt, input.EndsAt, input.Reason)
	if err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, map[string]int64{"id": id})
}
func (h Handler) createIssue(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RoomID      *int64 `json:"roomId"`
		EquipmentID *int64 `json:"equipmentId"`
		Description string `json:"description"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "请先登录")
		return
	}
	item, err := h.service.CreateIssueReport(r.Context(), user.ID, input.RoomID, input.EquipmentID, input.Description)
	if err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{"issue": item})
}
func (h Handler) listMyIssues(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, 401, "请先登录")
		return
	}
	issues, err := h.service.ListIssueReports(r.Context(), &user.ID)
	if err != nil {
		httpx.WriteError(w, 500, "读取报修记录失败")
		return
	}
	writeJSON(w, 200, map[string]any{"issues": issues})
}
func (h Handler) listIssues(w http.ResponseWriter, r *http.Request) {
	issues, err := h.service.ListIssueReports(r.Context(), nil)
	if err != nil {
		httpx.WriteError(w, 500, "读取报修记录失败")
		return
	}
	writeJSON(w, 200, map[string]any{"issues": issues})
}
func (h Handler) updateIssue(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Status          string `json:"status"`
		EquipmentStatus string `json:"equipmentStatus"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, 400, "报修编号不正确")
		return
	}
	if err := h.service.UpdateIssueReport(r.Context(), id, input.Status, input.EquipmentStatus); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"message": "报修状态已更新"})
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
