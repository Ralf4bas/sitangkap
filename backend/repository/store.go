// Package repository berisi akses database. Query diambil persis dari db/queries.md.
package repository

import (
	"context"
	"errors"

	"sitangkap/backend/models"
)

// ErrNotFound dikembalikan Store bila data dengan id atau username itu tidak ada.
var ErrNotFound = errors.New("not found")

// Store adalah kontrak repository. Implementasi: MySQLStore (mysql.go) dan memStore (khusus tes di routes).
type Store interface {
	Ping(ctx context.Context) error                                                             // Q-09
	FindUserByUsername(ctx context.Context, username string) (models.UserWithHash, error)       // Q-01, ErrNotFound
	FindActiveUserByID(ctx context.Context, id int64) (models.User, error)                      // Q-02, ErrNotFound
	ListTransaksi(ctx context.Context, dari, sampai string) ([]models.TransaksiRow, error)      // Q-03 bila dari=="", Q-04 bila terisi
	GetTransaksi(ctx context.Context, id int64) (models.TransaksiRow, error)                    // Q-05, ErrNotFound
	CreateTransaksi(ctx context.Context, in models.TransaksiInput, userID int64) (int64, error) // Q-06, id baru
	UpdateTransaksi(ctx context.Context, id int64, in models.TransaksiInput) error              // Q-07, ErrNotFound bila 0 baris cocok
	DeleteTransaksi(ctx context.Context, id int64) error                                        // Q-08, ErrNotFound bila 0 baris
}
