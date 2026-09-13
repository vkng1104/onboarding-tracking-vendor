package auth

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type coordinatorRepositoryStub struct {
	findByEmail func(context.Context, string) (coordinatorWithPassword, error)
}

func (stub coordinatorRepositoryStub) FindByEmail(
	ctx context.Context,
	email string,
) (coordinatorWithPassword, error) {
	return stub.findByEmail(ctx, email)
}

func TestServiceLoginAuthenticateAndLogout(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("demo1234"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("generate password hash: %v", err)
	}

	wantCoordinator := Coordinator{ID: "coordinator-1", Name: "Linh Nguyen", Email: "linh@demo.local"}
	repository := coordinatorRepositoryStub{
		findByEmail: func(_ context.Context, email string) (coordinatorWithPassword, error) {
			if email != wantCoordinator.Email {
				t.Fatalf("email = %q, want %q", email, wantCoordinator.Email)
			}
			return coordinatorWithPassword{Coordinator: wantCoordinator, PasswordHash: string(passwordHash)}, nil
		},
	}
	now := time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC)
	sessions := newSessionStore(8*time.Hour, func() time.Time { return now }, bytes.NewReader(make([]byte, 64)))
	service := NewService(repository, sessions)

	result, err := service.Login(context.Background(), "  LINH@DEMO.LOCAL ", "demo1234")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if result.Coordinator != wantCoordinator {
		t.Fatalf("coordinator = %#v, want %#v", result.Coordinator, wantCoordinator)
	}
	if result.Token == "" {
		t.Fatal("token is empty")
	}
	if want := now.Add(8 * time.Hour); !result.ExpiresAt.Equal(want) {
		t.Fatalf("expires at = %s, want %s", result.ExpiresAt, want)
	}

	authenticated, err := service.Authenticate(result.Token)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if authenticated != wantCoordinator {
		t.Fatalf("authenticated coordinator = %#v, want %#v", authenticated, wantCoordinator)
	}

	service.Logout(result.Token)
	if _, err := service.Authenticate(result.Token); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("authenticate after logout error = %v, want %v", err, ErrUnauthenticated)
	}
}

func TestServiceRejectsInvalidCredentials(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("demo1234"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("generate password hash: %v", err)
	}

	tests := []struct {
		name       string
		email      string
		password   string
		repository coordinatorRepositoryStub
	}{
		{
			name:     "missing email",
			password: "demo1234",
			repository: coordinatorRepositoryStub{findByEmail: func(context.Context, string) (coordinatorWithPassword, error) {
				t.Fatal("repository should not be called")
				return coordinatorWithPassword{}, nil
			}},
		},
		{
			name:     "unknown email",
			email:    "unknown@demo.local",
			password: "demo1234",
			repository: coordinatorRepositoryStub{findByEmail: func(context.Context, string) (coordinatorWithPassword, error) {
				return coordinatorWithPassword{}, errCoordinatorNotFound
			}},
		},
		{
			name:     "wrong password",
			email:    "linh@demo.local",
			password: "wrong",
			repository: coordinatorRepositoryStub{findByEmail: func(context.Context, string) (coordinatorWithPassword, error) {
				return coordinatorWithPassword{
					Coordinator:  Coordinator{ID: "coordinator-1", Email: "linh@demo.local"},
					PasswordHash: string(passwordHash),
				}, nil
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.repository, NewSessionStore(8*time.Hour))
			if _, err := service.Login(context.Background(), tt.email, tt.password); !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("login error = %v, want %v", err, ErrInvalidCredentials)
			}
		})
	}
}

func TestSessionExpiresAtBoundary(t *testing.T) {
	now := time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC)
	store := newSessionStore(8*time.Hour, func() time.Time { return now }, bytes.NewReader(make([]byte, 64)))
	token, _, err := store.Create(Coordinator{ID: "coordinator-1"})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	now = now.Add(8 * time.Hour)
	if _, ok := store.Get(token); ok {
		t.Fatal("session remains valid at the expiry boundary")
	}
}

func TestServiceDoesNotTreatCorruptHashAsBadCredentials(t *testing.T) {
	repository := coordinatorRepositoryStub{findByEmail: func(context.Context, string) (coordinatorWithPassword, error) {
		return coordinatorWithPassword{
			Coordinator:  Coordinator{ID: "coordinator-1", Email: "linh@demo.local"},
			PasswordHash: "not-a-bcrypt-hash",
		}, nil
	}}
	service := NewService(repository, NewSessionStore(8*time.Hour))

	_, err := service.Login(context.Background(), "linh@demo.local", "demo1234")
	if err == nil || errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("login error = %v, want an internal hash error", err)
	}
}
