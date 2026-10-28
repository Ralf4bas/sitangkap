package services

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// NamaCookie adalah nama cookie sesi dari kontrak API.
const NamaCookie = "sitangkap_session"

// Session adalah satu sesi login di memori (restart backend = semua pengguna login ulang, disengaja).
type Session struct {
	UserID  int64
	Role    string
	Expires time.Time
}

// Sessions menyimpan sesi di memori (map + mutex), sesuai kontrak API POST /api/login.
type Sessions struct {
	mu  sync.Mutex
	m   map[string]Session
	ttl time.Duration
	now func() time.Time
}

func NewSessions(ttl time.Duration, now func() time.Time) *Sessions {
	return &Sessions{m: map[string]Session{}, ttl: ttl, now: now}
}

// Create membuat token 32 byte dari crypto/rand (hex, 64 karakter). Sesi kedaluwarsa ikut dibersihkan.
func (s *Sessions) Create(userID int64, role string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for t, sess := range s.m {
		if !now.Before(sess.Expires) {
			delete(s.m, t)
		}
	}
	s.m[token] = Session{UserID: userID, Role: role, Expires: now.Add(s.ttl)}
	return token, nil
}

// Get mengembalikan sesi yang masih berlaku; sesi kedaluwarsa dihapus dan dianggap tidak ada.
func (s *Sessions) Get(token string) (Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.m[token]
	if !ok {
		return Session{}, false
	}
	if !s.now().Before(sess.Expires) {
		delete(s.m, token)
		return Session{}, false
	}
	return sess, true
}

func (s *Sessions) Delete(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, token)
}

// Jumlah mengembalikan banyaknya sesi yang tersimpan (untuk tes).
func (s *Sessions) Jumlah() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.m)
}
