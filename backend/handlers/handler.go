// Package handlers menerima request dan menulis respons. Aturan bisnis ada di services,
// query database di repository.
package handlers

import (
	"time"

	"sitangkap/backend/config"
	"sitangkap/backend/repository"
	"sitangkap/backend/services"
)

// Handler menyimpan dependensi yang dipakai semua handler.
type Handler struct {
	Store    repository.Store
	Sessions *services.Sessions
	Cfg      config.Config
	Now      func() time.Time // time.Now di production, waktu tetap di tes
}

func New(store repository.Store, sessions *services.Sessions, cfg config.Config, now func() time.Time) *Handler {
	return &Handler{Store: store, Sessions: sessions, Cfg: cfg, Now: now}
}

// today adalah "hari ini" di zona APP_TIMEZONE.
func (h *Handler) today() time.Time {
	return h.Now().In(h.Cfg.Location)
}
