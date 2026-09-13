package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthenticationHTTPFlow(t *testing.T) {
	router := testAuthRouter(t, nil)

	login := performAuthRequest(
		router,
		http.MethodPost,
		"/api/v1/auth/login",
		`{"email":"linh@demo.local","password":"demo1234"}`,
		nil,
	)
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d; body = %s", login.Code, http.StatusOK, login.Body.String())
	}
	if strings.Contains(login.Body.String(), "password") {
		t.Fatalf("login response exposes password data: %s", login.Body.String())
	}

	responseCookies := login.Result().Cookies()
	if len(responseCookies) != 1 {
		t.Fatalf("login cookies = %d, want 1", len(responseCookies))
	}
	sessionCookie := responseCookies[0]
	if sessionCookie.Name != sessionCookieName || !sessionCookie.HttpOnly || sessionCookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected session cookie: %#v", sessionCookie)
	}

	me := performAuthRequest(router, http.MethodGet, "/api/v1/auth/me", "", sessionCookie)
	if me.Code != http.StatusOK {
		t.Fatalf("me status = %d, want %d; body = %s", me.Code, http.StatusOK, me.Body.String())
	}
	if !strings.Contains(me.Body.String(), `"email":"linh@demo.local"`) {
		t.Fatalf("me body = %s", me.Body.String())
	}

	logout := performAuthRequest(router, http.MethodPost, "/api/v1/auth/logout", "", sessionCookie)
	if logout.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d, want %d", logout.Code, http.StatusNoContent)
	}
	logoutCookies := logout.Result().Cookies()
	if len(logoutCookies) != 1 || logoutCookies[0].MaxAge >= 0 {
		t.Fatalf("logout did not expire the session cookie: %#v", logoutCookies)
	}

	afterLogout := performAuthRequest(router, http.MethodGet, "/api/v1/auth/me", "", sessionCookie)
	if afterLogout.Code != http.StatusUnauthorized {
		t.Fatalf("me after logout status = %d, want %d", afterLogout.Code, http.StatusUnauthorized)
	}
}

func TestAuthenticationHTTPFailures(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		repoError  error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "missing cookie",
			method:     http.MethodGet,
			path:       "/api/v1/auth/me",
			wantStatus: http.StatusUnauthorized,
			wantCode:   "UNAUTHENTICATED",
		},
		{
			name:       "bad password",
			method:     http.MethodPost,
			path:       "/api/v1/auth/login",
			body:       `{"email":"linh@demo.local","password":"wrong"}`,
			wantStatus: http.StatusUnauthorized,
			wantCode:   "INVALID_CREDENTIALS",
		},
		{
			name:       "unknown request field",
			method:     http.MethodPost,
			path:       "/api/v1/auth/login",
			body:       `{"email":"linh@demo.local","password":"demo1234","role":"admin"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_REQUEST",
		},
		{
			name:       "repository failure is hidden",
			method:     http.MethodPost,
			path:       "/api/v1/auth/login",
			body:       `{"email":"linh@demo.local","password":"demo1234"}`,
			repoError:  errors.New("select coordinators: password_hash leaked"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := testAuthRouter(t, tt.repoError)
			response := performAuthRequest(router, tt.method, tt.path, tt.body, nil)

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

func TestLogoutWithoutSessionIsIdempotent(t *testing.T) {
	router := testAuthRouter(t, nil)
	response := performAuthRequest(router, http.MethodPost, "/api/v1/auth/logout", "", nil)

	if response.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d, want %d", response.Code, http.StatusNoContent)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge >= 0 {
		t.Fatalf("logout did not expire the session cookie: %#v", cookies)
	}
}

func testAuthRouter(t *testing.T, repositoryError error) chi.Router {
	t.Helper()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("demo1234"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("generate password hash: %v", err)
	}

	repository := coordinatorRepositoryStub{findByEmail: func(context.Context, string) (coordinatorWithPassword, error) {
		if repositoryError != nil {
			return coordinatorWithPassword{}, repositoryError
		}
		return coordinatorWithPassword{
			Coordinator: Coordinator{
				ID:    "10000000-0000-0000-0000-000000000001",
				Name:  "Linh Nguyen",
				Email: "linh@demo.local",
			},
			PasswordHash: string(passwordHash),
		}, nil
	}}
	service := NewService(repository, NewSessionStore(8*time.Hour))
	handler := NewHandler(service)
	router := chi.NewRouter()
	handler.MountRoutes(router)
	return router
}

func performAuthRequest(
	router http.Handler,
	method string,
	path string,
	body string,
	cookie *http.Cookie,
) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
