// Package middlewares berisi pembungkus handler yang dipakai berulang: cek sesi dan peran,
// log request, dan penangkap panic.
package middlewares

import (
	"errors"
	"net/http"
	"slices"

	"sitangkap/backend/helpers"
	"sitangkap/backend/models"
	"sitangkap/backend/repository"
	"sitangkap/backend/services"
)

// HandlerDenganUser adalah handler yang butuh user yang sudah login.
type HandlerDenganUser func(w http.ResponseWriter, r *http.Request, u models.User)

// Auth memeriksa cookie sesi dan peran sebelum memanggil handler.
type Auth struct {
	Store    repository.Store
	Sessions *services.Sessions
}

// Require membungkus handler yang butuh login. roles kosong = semua peran boleh.
// Urutan: cookie atau sesi tidak dikenal -> 401; user tidak aktif lagi -> hapus sesi, 401; peran salah -> 403.
func (a Auth) Require(roles []string, next HandlerDenganUser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(services.NamaCookie)
		if err != nil {
			helpers.Unauthorized(w)
			return
		}
		sess, ok := a.Sessions.Get(c.Value)
		if !ok {
			helpers.Unauthorized(w)
			return
		}
		u, err := a.Store.FindActiveUserByID(r.Context(), sess.UserID)
		if errors.Is(err, repository.ErrNotFound) {
			a.Sessions.Delete(c.Value)
			helpers.Unauthorized(w)
			return
		}
		if err != nil {
			helpers.InternalError(w, r, err)
			return
		}
		if len(roles) > 0 && !slices.Contains(roles, u.Role) {
			helpers.Forbidden(w)
			return
		}
		next(w, r, u)
	}
}
