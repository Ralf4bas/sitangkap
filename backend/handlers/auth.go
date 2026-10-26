package handlers

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"sitangkap/backend/helpers"
	"sitangkap/backend/models"
	"sitangkap/backend/repository"
	"sitangkap/backend/services"
)

// hashPalsu adalah hash bcrypt cost 12 yang valid. Bila username tidak ada, password tetap dibandingkan
// dengan hash ini supaya waktu respons tidak membocorkan username mana yang terdaftar.
var hashPalsu = sync.OnceValue(func() []byte {
	h, err := bcrypt.GenerateFromPassword([]byte("dummy"), 12)
	if err != nil {
		panic(err)
	}
	return h
})

func (h *Handler) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     services.NamaCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   int(h.Cfg.SessionTTL / time.Second),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.Cfg.CookieSecure,
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     services.NamaCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // ditulis sebagai Max-Age=0
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.Cfg.CookieSecure,
	})
}

type loginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginHandler: POST /api/login.
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var body loginBody
	if !helpers.DecodeBody(w, r, &body) {
		return
	}
	username := strings.TrimSpace(body.Username)
	fields := map[string]string{}
	if username == "" {
		fields["username"] = "Username wajib diisi."
	}
	if body.Password == "" {
		fields["password"] = "Password wajib diisi."
	}
	if len(fields) > 0 {
		helpers.ValidationError(w, fields)
		return
	}

	u, err := h.Store.FindUserByUsername(r.Context(), username)
	if errors.Is(err, repository.ErrNotFound) {
		bcrypt.CompareHashAndPassword(hashPalsu(), []byte(body.Password))
		helpers.InvalidCredentials(w)
		return
	}
	if err != nil {
		helpers.InternalError(w, r, err)
		return
	}
	cocok := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(body.Password)) == nil
	if !cocok || !u.Aktif {
		helpers.InvalidCredentials(w)
		return
	}

	token, err := h.Sessions.Create(u.ID, u.Role)
	if err != nil {
		helpers.InternalError(w, r, err)
		return
	}
	h.setSessionCookie(w, token)
	helpers.WriteData(w, http.StatusOK, u.User)
}

// LogoutHandler: POST /api/logout. Selalu 204, ada sesi maupun tidak.
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(services.NamaCookie); err == nil {
		h.Sessions.Delete(c.Value)
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// MeHandler: GET /api/me.
func (h *Handler) MeHandler(w http.ResponseWriter, r *http.Request, u models.User) {
	helpers.WriteData(w, http.StatusOK, u)
}
