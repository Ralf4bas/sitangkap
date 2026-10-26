package services

import (
	"testing"
	"time"
)

func TestSessionsKedaluwarsa(t *testing.T) {
	sekarang := time.Date(2026, 10, 22, 2, 0, 0, 0, time.UTC)
	s := NewSessions(8*time.Hour, func() time.Time { return sekarang })

	token, err := s.Create(1, "petugas")
	if err != nil {
		t.Fatal(err)
	}
	if len(token) != 64 {
		t.Errorf("panjang token = %d, mau 64", len(token))
	}
	if sess, ok := s.Get(token); !ok || sess.UserID != 1 || sess.Role != "petugas" {
		t.Fatalf("get = %+v, %v", sess, ok)
	}

	sekarang = sekarang.Add(8 * time.Hour)
	if _, ok := s.Get(token); ok {
		t.Error("sesi harus kedaluwarsa tepat setelah 8 jam")
	}
	if len(s.m) != 0 {
		t.Error("sesi kedaluwarsa harus dihapus")
	}
}

func TestSessionsTokenUnikDanDelete(t *testing.T) {
	s := NewSessions(time.Hour, time.Now)
	a, _ := s.Create(1, "petugas")
	b, _ := s.Create(1, "petugas")
	if a == b {
		t.Fatal("token harus unik")
	}
	s.Delete(a)
	if _, ok := s.Get(a); ok {
		t.Error("token a harus terhapus")
	}
	if _, ok := s.Get(b); !ok {
		t.Error("token b harus tetap ada")
	}
}
