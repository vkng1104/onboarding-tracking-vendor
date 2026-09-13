package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
	"time"
)

const sessionTokenBytes = 32

type session struct {
	coordinator Coordinator
	expiresAt   time.Time
}

type SessionStore struct {
	mu       sync.Mutex
	sessions map[string]session
	ttl      time.Duration
	now      func() time.Time
	random   io.Reader
}

func NewSessionStore(ttl time.Duration) *SessionStore {
	return newSessionStore(ttl, time.Now, rand.Reader)
}

func newSessionStore(ttl time.Duration, now func() time.Time, random io.Reader) *SessionStore {
	return &SessionStore{
		sessions: make(map[string]session),
		ttl:      ttl,
		now:      now,
		random:   random,
	}
}

func (s *SessionStore) Create(coordinator Coordinator) (string, time.Time, error) {
	bytes := make([]byte, sessionTokenBytes)
	if _, err := io.ReadFull(s.random, bytes); err != nil {
		return "", time.Time{}, fmt.Errorf("generate session token: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(bytes)
	expiresAt := s.now().Add(s.ttl)

	s.mu.Lock()
	s.sessions[token] = session{coordinator: coordinator, expiresAt: expiresAt}
	s.mu.Unlock()

	return token, expiresAt, nil
}

func (s *SessionStore) Get(token string) (Coordinator, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored, ok := s.sessions[token]
	if !ok {
		return Coordinator{}, false
	}
	if !s.now().Before(stored.expiresAt) {
		delete(s.sessions, token)
		return Coordinator{}, false
	}
	return stored.coordinator, true
}

func (s *SessionStore) Delete(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}
