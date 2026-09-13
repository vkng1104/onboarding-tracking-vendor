package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type CoordinatorRepository interface {
	FindByEmail(context.Context, string) (coordinatorWithPassword, error)
}

type LoginResult struct {
	Coordinator Coordinator
	Token       string
	ExpiresAt   time.Time
}

type Service struct {
	coordinators CoordinatorRepository
	sessions     *SessionStore
}

func NewService(coordinators CoordinatorRepository, sessions *SessionStore) *Service {
	return &Service{coordinators: coordinators, sessions: sessions}
}

func (s *Service) Login(ctx context.Context, email, password string) (LoginResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return LoginResult{}, ErrInvalidCredentials
	}

	coordinator, err := s.coordinators.FindByEmail(ctx, email)
	if errors.Is(err, errCoordinatorNotFound) {
		return LoginResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, fmt.Errorf("find coordinator: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(coordinator.PasswordHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, fmt.Errorf("verify password hash: %w", err)
	}

	token, expiresAt, err := s.sessions.Create(coordinator.Coordinator)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		Coordinator: coordinator.Coordinator,
		Token:       token,
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *Service) Authenticate(token string) (Coordinator, error) {
	coordinator, ok := s.sessions.Get(token)
	if !ok {
		return Coordinator{}, ErrUnauthenticated
	}
	return coordinator, nil
}

func (s *Service) Logout(token string) {
	s.sessions.Delete(token)
}
