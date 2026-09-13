package health

import (
	"context"
	"net/http"

	"github.com/khanhvunguyen/vendor-onboarding-tracker/internal/httpx"
)

type Pinger interface {
	Ping(context.Context) error
}

type Handler struct {
	database Pinger
}

func NewHandler(database Pinger) *Handler {
	return &Handler{database: database}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.database.Ping(r.Context()); err != nil {
		httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
