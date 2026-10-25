package main

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound dikembalikan Store bila data dengan id atau username itu tidak ada.
var ErrNotFound = errors.New("not found")

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nama     string `json:"nama"`
	Role     string `json:"role"`
}

type UserWithHash struct {
	User
	PasswordHash string
	Aktif        bool
}

// TransaksiInput adalah input yang sudah lolos validasi.
type TransaksiInput struct {
	NamaNelayan string
	JenisIkan   string
	BeratCenti  int64 // berat dalam 0,01 kg: 120.5 kg -> 12050
	HargaPerKg  int64
	Tanggal     string // "YYYY-MM-DD"
}

// TransaksiRow adalah satu baris dari database (Q-03/Q-04/Q-05).
type TransaksiRow struct {
	ID          int64
	NamaNelayan string
	JenisIkan   string
	BeratCenti  int64
	HargaPerKg  int64
	Tanggal     time.Time // DATE, dibaca UTC tengah malam
	DibuatOleh  string    // users.username
	DibuatPada  time.Time // UTC
}

// Store adalah kontrak repository. Implementasi: mysqlStore (store_mysql.go) dan memStore (khusus tes).
type Store interface {
	Ping(ctx context.Context) error                                                      // Q-09
	FindUserByUsername(ctx context.Context, username string) (UserWithHash, error)       // Q-01, ErrNotFound
	FindActiveUserByID(ctx context.Context, id int64) (User, error)                      // Q-02, ErrNotFound
	ListTransaksi(ctx context.Context, dari, sampai string) ([]TransaksiRow, error)      // Q-03 bila dari=="", Q-04 bila terisi
	GetTransaksi(ctx context.Context, id int64) (TransaksiRow, error)                    // Q-05, ErrNotFound
	CreateTransaksi(ctx context.Context, in TransaksiInput, userID int64) (int64, error) // Q-06, id baru
	UpdateTransaksi(ctx context.Context, id int64, in TransaksiInput) error              // Q-07, ErrNotFound bila 0 baris cocok
	DeleteTransaksi(ctx context.Context, id int64) error                                 // Q-08, ErrNotFound bila 0 baris
}
