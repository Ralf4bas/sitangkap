// Package models berisi struktur data yang dipakai bersama oleh repository, services, dan handlers.
package models

// User adalah objek User di kontrak API. password_hash tidak pernah dikirim ke klien.
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nama     string `json:"nama"`
	Role     string `json:"role"`
}

// UserWithHash dipakai saat login (Q-01).
type UserWithHash struct {
	User
	PasswordHash string
	Aktif        bool
}
