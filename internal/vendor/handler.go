package vendor

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/khanhvunguyen/vendor-onboarding-tracker/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) MountRoutes(router chi.Router, requireAuthentication func(http.Handler) http.Handler) {
	router.Group(func(router chi.Router) {
		router.Use(requireAuthentication)
		router.Get("/api/v1/vendors", h.list)
		router.Get("/api/v1/vendors/{vendorID}/history", h.history)
	})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	vendors, err := h.service.List(r.Context())
	if err != nil {
		h.writeMappedError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string][]Vendor{"vendors": vendors})
}

func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	history, err := h.service.History(r.Context(), chi.URLParam(r, "vendorID"))
	if err != nil {
		h.writeMappedError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string][]HistoryEvent{"history": history})
}

func (h *Handler) writeMappedError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "VENDOR_NOT_FOUND", "Vendor was not found.")
		return
	}

	slog.ErrorContext(r.Context(), "vendor request failed", "error", err)
	httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
}
