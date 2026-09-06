package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/ly1611240037/bandroom/backend/internal/httpx"
)

const sessionCookie = "bandroom_session"

type Handler struct{ service *Service }

func NewHandler(service *Service) Handler { return Handler{service: service} }

func (h Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/register", h.register)
	mux.HandleFunc("GET /api/auth/verify", h.verify)
	mux.HandleFunc("POST /api/auth/login", h.login)
	mux.HandleFunc("POST /api/auth/password-reset/request", h.requestPasswordReset)
	mux.HandleFunc("POST /api/auth/password-reset/confirm", h.resetPassword)
	mux.HandleFunc("POST /api/auth/logout", h.logout)
	mux.HandleFunc("GET /api/auth/me", h.me)
	mux.Handle("GET /api/owner/ping", h.RequireRole("owner", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})))
}

func (h Handler) register(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	user, err := h.service.Register(r.Context(), input.Name, input.Email, input.Phone, input.Password)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"user": user, "message": "注册成功，请查收邮箱并完成验证"})
}

func (h Handler) verify(w http.ResponseWriter, r *http.Request) {
	if err := h.service.VerifyEmail(r.Context(), r.URL.Query().Get("token")); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "邮箱验证成功"})
}

func (h Handler) login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	user, token, err := h.service.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, ErrEmailNotVerified) {
			status = http.StatusForbidden
		}
		httpx.WriteError(w, status, err.Error())
		return
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: int(h.service.sessionTTL.Seconds())})
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h Handler) requestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := h.service.RequestPasswordReset(r.Context(), input.Email); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "暂时无法发送重置邮件")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "如果邮箱已注册，重置链接将发送到该邮箱"})
}

func (h Handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := h.service.ResetPassword(r.Context(), input.Token, input.Password); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "密码重置成功，请重新登录"})
}

func (h Handler) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		_ = h.service.Logout(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", HttpOnly: true, MaxAge: -1, SameSite: http.SameSiteLaxMode})
	writeJSON(w, http.StatusOK, map[string]string{"message": "已退出登录"})
}

func (h Handler) me(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h Handler) RequireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := h.currentUser(r)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, err.Error())
			return
		}
		if user.Role != role {
			httpx.WriteError(w, http.StatusForbidden, ErrForbidden.Error())
			return
		}
		ctx := withUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h Handler) RequireVerifiedCustomer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := h.currentUser(r)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, err.Error())
			return
		}
		if user.Role != "customer" {
			httpx.WriteError(w, http.StatusForbidden, ErrForbidden.Error())
			return
		}
		if err := h.service.IsVerifiedCustomer(user); err != nil {
			httpx.WriteError(w, http.StatusForbidden, err.Error())
			return
		}
		next.ServeHTTP(w, r.WithContext(withUser(r.Context(), user)))
	})
}

type contextKey string

const userContextKey contextKey = "bandroom.user"

func withUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}
func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userContextKey).(User)
	return user, ok
}

func (h Handler) currentUser(r *http.Request) (User, error) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return User{}, ErrUnauthorized
	}
	return h.service.CurrentUser(r.Context(), cookie.Value)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "请求格式不正确")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
