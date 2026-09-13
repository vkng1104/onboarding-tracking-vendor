package vendor

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/khanhvunguyen/vendor-onboarding-tracker/internal/auth"
	"github.com/khanhvunguyen/vendor-onboarding-tracker/internal/httpx"
)

type coordinatorFromContext func(context.Context) (CoordinatorSummary, bool)

type Handler struct {
	service                *Service
	coordinatorFromContext coordinatorFromContext
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
		coordinatorFromContext: func(ctx context.Context) (CoordinatorSummary, bool) {
			coordinator, ok := auth.CoordinatorFromContext(ctx)
			return CoordinatorSummary{ID: coordinator.ID, Name: coordinator.Name}, ok
		},
	}
}

func (h *Handler) MountRoutes(router chi.Router, requireAuthentication func(http.Handler) http.Handler) {
	router.Group(func(router chi.Router) {
		router.Use(requireAuthentication)
		router.Get("/api/v1/vendors", h.list)
		router.Get("/api/v1/vendors/{vendorID}/history", h.history)
		router.Patch("/api/v1/vendors/{vendorID}/stage", h.updateStage)
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

func (h *Handler) updateStage(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ExpectedCurrentStage Stage `json:"expected_current_stage"`
		NewStage             Stage `json:"new_stage"`
	}
	if err := httpx.DecodeJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Request body must be valid JSON.")
		return
	}

	coordinator, ok := h.coordinatorFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
		return
	}

	transition, err := h.service.UpdateStage(
		r.Context(),
		chi.URLParam(r, "vendorID"),
		coordinator,
		request.ExpectedCurrentStage,
		request.NewStage,
	)
	if err != nil {
		h.writeMappedError(w, r, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]HistoryEvent{"transition": transition})
}

func (h *Handler) writeMappedError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "VENDOR_NOT_FOUND", "Vendor was not found.")
		return
	}
	if errors.Is(err, ErrInvalidStage) {
		httpx.WriteError(w, http.StatusUnprocessableEntity, "INVALID_STAGE", "Stage must be a known workflow stage.")
		return
	}
	if errors.Is(err, ErrStageUnchanged) {
		httpx.WriteError(w, http.StatusConflict, "STAGE_UNCHANGED", "Choose a stage different from the current stage.")
		return
	}
	if errors.Is(err, ErrStageConflict) {
		httpx.WriteError(w, http.StatusConflict, "STAGE_CONFLICT", "Vendor stage changed since it was loaded.")
		return
	}

	slog.ErrorContext(r.Context(), "vendor request failed", "error", err)
	httpx.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.")
}
