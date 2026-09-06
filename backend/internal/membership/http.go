package membership

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	mux.Handle("GET /api/customer/membership/cards", h.auth.RequireRole("customer", http.HandlerFunc(h.listMyCards)))
	mux.Handle("GET /api/owner/membership/plans", h.auth.RequireRole("owner", http.HandlerFunc(h.listPlans)))
	mux.Handle("POST /api/owner/membership/plans", h.auth.RequireRole("owner", http.HandlerFunc(h.createPlan)))
	mux.Handle("PATCH /api/owner/membership/plans/{id}", h.auth.RequireRole("owner", http.HandlerFunc(h.updatePlan)))
	mux.Handle("PATCH /api/owner/membership/plans/{id}/status", h.auth.RequireRole("owner", http.HandlerFunc(h.setPlanStatus)))
	mux.Handle("POST /api/owner/membership/cards", h.auth.RequireRole("owner", http.HandlerFunc(h.activateCard)))
}

func (h Handler) listMyCards(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "请先登录")
		return
	}
	cards, err := h.service.ListCards(r.Context(), user.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "读取会员卡失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"cards": cards})
}

func (h Handler) listPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.service.ListPlans(r.Context(), true)
	if err != nil {
		httpx.WriteError(w, 500, "读取会员方案失败")
		return
	}
	writeJSON(w, 200, map[string]any{"plans": plans})
}
func (h Handler) createPlan(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PlanType     string `json:"planType"`
		Name         string `json:"name"`
		PriceCents   int64  `json:"priceCents"`
		IncludedUses *int   `json:"includedUses"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	plan, err := h.service.CreatePlan(r.Context(), input.PlanType, input.Name, input.PriceCents, input.IncludedUses)
	if err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{"plan": plan})
}
func (h Handler) updatePlan(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name         string `json:"name"`
		PriceCents   int64  `json:"priceCents"`
		IncludedUses *int   `json:"includedUses"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, 400, "方案编号不正确")
		return
	}
	plan, err := h.service.UpdatePlan(r.Context(), id, input.Name, input.PriceCents, input.IncludedUses)
	if err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"plan": plan})
}
func (h Handler) setPlanStatus(w http.ResponseWriter, r *http.Request) {
	var input struct {
		IsActive bool `json:"isActive"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, 400, "方案编号不正确")
		return
	}
	if err := h.service.SetPlanActive(r.Context(), id, input.IsActive); err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"message": "方案状态已更新"})
}
func (h Handler) activateCard(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID          int64  `json:"userId"`
		PlanID          int64  `json:"planId"`
		PaidAmountCents int64  `json:"paidAmountCents"`
		PaymentMethod   string `json:"paymentMethod"`
		Notes           string `json:"notes"`
		PurchaseDate    string `json:"purchaseDate"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	purchase := time.Now()
	var err error
	if input.PurchaseDate != "" {
		purchase, err = time.Parse(time.RFC3339, input.PurchaseDate)
		if err != nil {
			httpx.WriteError(w, 400, "购买日期格式不正确")
			return
		}
	}
	card, err := h.service.ActivateCard(r.Context(), input.UserID, input.PlanID, input.PaidAmountCents, input.PaymentMethod, input.Notes, purchase)
	if err != nil {
		httpx.WriteError(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{"card": card})
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
