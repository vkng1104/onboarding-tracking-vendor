package vendor

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func TestVendorHTTPResponses(t *testing.T) {
	now := time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC)
	store := storeStub{
		list: func(context.Context) ([]record, error) {
			return []record{{
				ID:              "vendor-1",
				Name:            "Company A",
				Region:          "HCMC",
				CurrentStage:    StageContractSigned,
				StageEnteredAt:  now.Add(-48 * time.Hour),
				CoordinatorID:   "coordinator-1",
				CoordinatorName: "Linh Nguyen",
			}}, nil
		},
		history: func(_ context.Context, vendorID string) ([]HistoryEvent, error) {
			if vendorID != "vendor-1" {
				t.Fatalf("vendor id = %q, want vendor-1", vendorID)
			}
			return []HistoryEvent{{
				ID:            "event-1",
				OccurredAt:    now.Add(-48 * time.Hour),
				Actor:         CoordinatorSummary{ID: "coordinator-1", Name: "Linh Nguyen"},
				PreviousStage: StageContractSent,
				NewStage:      StageContractSigned,
			}}, nil
		},
	}
	router := testVendorRouter(store)

	list := performVendorRequest(router, "/api/v1/vendors")
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d; body = %s", list.Code, http.StatusOK, list.Body.String())
	}
	for _, expected := range []string{
		`"name":"Company A"`,
		`"hours_in_current_stage":48`,
		`"is_stuck":false`,
		`"next_stage":"KYC_DOCS_RECEIVED"`,
	} {
		if !strings.Contains(list.Body.String(), expected) {
			t.Errorf("list body = %s, want %s", list.Body.String(), expected)
		}
	}

	history := performVendorRequest(router, "/api/v1/vendors/vendor-1/history")
	if history.Code != http.StatusOK {
		t.Fatalf("history status = %d, want %d; body = %s", history.Code, http.StatusOK, history.Body.String())
	}
	if !strings.Contains(history.Body.String(), `"actor":{"id":"coordinator-1","name":"Linh Nguyen"}`) {
		t.Fatalf("history body = %s, want actor attribution", history.Body.String())
	}
}

func TestVendorHTTPErrorMappingDoesNotLeakInternals(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "missing vendor", err: ErrNotFound, wantStatus: http.StatusNotFound, wantCode: "VENDOR_NOT_FOUND"},
		{name: "repository failure", err: errors.New("select vendors: password_hash leaked"), wantStatus: http.StatusInternalServerError, wantCode: "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storeStub{
				list:    func(context.Context) ([]record, error) { return nil, tt.err },
				history: func(context.Context, string) ([]HistoryEvent, error) { return nil, tt.err },
			}
			router := testVendorRouter(store)
			path := "/api/v1/vendors"
			if errors.Is(tt.err, ErrNotFound) {
				path = "/api/v1/vendors/missing/history"
			}
			response := performVendorRequest(router, path)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, tt.wantStatus, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), `"code":"`+tt.wantCode+`"`) {
				t.Fatalf("body = %s, want code %s", response.Body.String(), tt.wantCode)
			}
			if strings.Contains(response.Body.String(), "password_hash") {
				t.Fatalf("response leaks internal error: %s", response.Body.String())
			}
		})
	}
}

func TestVendorRoutesApplyAuthenticationMiddleware(t *testing.T) {
	store := storeStub{
		list: func(context.Context) ([]record, error) {
			t.Fatal("store should not be called")
			return nil, nil
		},
		history: func(context.Context, string) ([]HistoryEvent, error) {
			t.Fatal("store should not be called")
			return nil, nil
		},
	}
	handler := NewHandler(NewService(store, 7))
	router := chi.NewRouter()
	handler.MountRoutes(router, func(http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		})
	})

	response := performVendorRequest(router, "/api/v1/vendors")
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func testVendorRouter(store Store) chi.Router {
	service := newService(store, 7*24*time.Hour, func() time.Time {
		return time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC)
	})
	handler := NewHandler(service)
	router := chi.NewRouter()
	handler.MountRoutes(router, func(next http.Handler) http.Handler { return next })
	return router
}

func performVendorRequest(router http.Handler, path string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, path, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
