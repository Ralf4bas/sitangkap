package main

import (
	"testing"
	"time"
)

func TestSessionsKedaluwarsa(t *testing.T) {
	sekarang := time.Date(2026, 10, 22, 2, 0, 0, 0, time.UTC)
	s := newSessions(8*time.Hour, func() time.Time { return sekarang })

	token, err := s.create(1, "petugas")
	if err != nil {
		t.Fatal(err)
	}
	if len(token) != 64 {
		t.Errorf("panjang token = %d, mau 64", len(token))
	}
	if sess, ok := s.get(token); !ok || sess.userID != 1 || sess.role != "petugas" {
		t.Fatalf("get = %+v, %v", sess, ok)
	}

	sekarang = sekarang.Add(8 * time.Hour)
	if _, ok := s.get(token); ok {
		t.Error("sesi harus kedaluwarsa tepat setelah 8 jam")
	}
	if len(s.m) != 0 {
		t.Error("sesi kedaluwarsa harus dihapus")
	}
}

func TestSessionsTokenUnikDanDelete(t *testing.T) {
	s := newSessions(time.Hour, time.Now)
	a, _ := s.create(1, "petugas")
	b, _ := s.create(1, "petugas")
	if a == b {
		t.Fatal("token harus unik")
	}
	s.delete(a)
	if _, ok := s.get(a); ok {
		t.Error("token a harus terhapus")
	}
	if _, ok := s.get(b); !ok {
		t.Error("token b harus tetap ada")
	}
}
