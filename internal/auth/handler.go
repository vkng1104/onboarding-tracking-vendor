package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/khanhvunguyen/vendor-onboarding-tracker/internal/httpx"
)

const sessionCookieName = "vendor_onboarding_session"

type coordinatorContextKey struct{}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) MountRoutes(router chi.Router) {
	router.Route("/api/v1/auth", func(router chi.Router) {
		router.Post("/login", h.login)
		router.Post("/logout", h.logout)
		router.Group(func(router chi.Router) {
			router.Use(h.RequireAuthentication)
			router.Get("/me", h.me)
		})
	})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Request body must be valid JSON.")
		return
	}

	result, err := h.service.Login(r.Context(), request.Email, request.Password)
	if err != nil {
		h.writeMappedError(w, r, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    result.Token,
		Path:     "/",
		Expires:  result.ExpiresAt,
		MaxAge:   int(time.Until(result.ExpiresAt).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	httpx.WriteJSON(w, http.StatusOK, map[string]Coordinator{"coordinator": result.Coordinator})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	coordinator, ok := CoordinatorFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]Coordinator{"coordinator": coordinator})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		h.service.Logout(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
			return
		}

		coordinator, err := h.service.Authenticate(cookie.Value)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
			return
		}

		ctx := context.WithValue(r.Context(), coordinatorContextKey{}, coordinator)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CoordinatorFromContext(ctx context.Context) (Coordinator, bool) {
	coordinator, ok := ctx.Value(coordinatorContextKey{}).(Coordinator)
	return coordinator, ok
}

func (h *Handler) writeMappedError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, ErrInvalidCredentials) {
		httpx.WriteError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Email or password is incorrect.")
		return
	}

	slog.ErrorContext(r.Context(), "authentication request failed", "error", err)
	httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
}
