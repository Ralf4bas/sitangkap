package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const namaCookie = "sitangkap_session"

// ---------- Sesi di memori (restart backend = semua pengguna login ulang, disengaja) ----------

type session struct {
	userID  int64
	role    string
	expires time.Time
}

type sessions struct {
	mu  sync.Mutex
	m   map[string]session
	ttl time.Duration
	now func() time.Time
}

func newSessions(ttl time.Duration, now func() time.Time) *sessions {
	return &sessions{m: map[string]session{}, ttl: ttl, now: now}
}

// create membuat token 32 byte dari crypto/rand (hex, 64 karakter). Sesi kedaluwarsa ikut dibersihkan.
func (s *sessions) create(userID int64, role string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for t, sess := range s.m {
		if !now.Before(sess.expires) {
			delete(s.m, t)
		}
	}
	s.m[token] = session{userID: userID, role: role, expires: now.Add(s.ttl)}
	return token, nil
}

// get mengembalikan sesi yang masih berlaku; sesi kedaluwarsa dihapus dan dianggap tidak ada.
func (s *sessions) get(token string) (session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.m[token]
	if !ok {
		return session{}, false
	}
	if !s.now().Before(sess.expires) {
		delete(s.m, token)
		return session{}, false
	}
	return sess, true
}

func (s *sessions) delete(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, token)
}

// ---------- Cookie ----------

func (a *App) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     namaCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   int(a.cfg.SessionTTL / time.Second),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   a.cfg.CookieSecure,
	})
}

func (a *App) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     namaCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // ditulis sebagai Max-Age=0
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   a.cfg.CookieSecure,
	})
}

// ---------- Hash palsu untuk login ----------

// hashPalsu adalah hash bcrypt cost 12 yang valid. Bila username tidak ada, password tetap dibandingkan
// dengan hash ini supaya waktu respons tidak membocorkan username mana yang terdaftar.
var hashPalsu = sync.OnceValue(func() []byte {
	h, err := bcrypt.GenerateFromPassword([]byte("dummy"), 12)
	if err != nil {
		panic(err)
	}
	return h
})

// ---------- Handler ----------

type loginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body loginBody
	if !decodeBody(w, r, &body) {
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
		validationError(w, fields)
		return
	}

	u, err := a.store.FindUserByUsername(r.Context(), username)
	if errors.Is(err, ErrNotFound) {
		bcrypt.CompareHashAndPassword(hashPalsu(), []byte(body.Password))
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", msgInvalidCredentials, nil)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	cocok := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(body.Password)) == nil
	if !cocok || !u.Aktif {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", msgInvalidCredentials, nil)
		return
	}

	token, err := a.sessions.create(u.ID, u.Role)
	if err != nil {
		internalError(w, r, err)
		return
	}
	a.setSessionCookie(w, token)
	writeData(w, http.StatusOK, u.User)
}

// handleLogout selalu 204, ada sesi maupun tidak.
func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(namaCookie); err == nil {
		a.sessions.delete(c.Value)
	}
	a.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleMe(w http.ResponseWriter, r *http.Request, u User) {
	writeData(w, http.StatusOK, u)
}

// withAuth membungkus handler yang butuh login. roles kosong = semua peran boleh.
// Urutan: cookie atau sesi tidak dikenal -> 401; user tidak aktif lagi -> hapus sesi, 401; peran salah -> 403.
func (a *App) withAuth(roles []string, h func(w http.ResponseWriter, r *http.Request, u User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(namaCookie)
		if err != nil {
			unauthorized(w)
			return
		}
		sess, ok := a.sessions.get(c.Value)
		if !ok {
			unauthorized(w)
			return
		}
		u, err := a.store.FindActiveUserByID(r.Context(), sess.userID)
		if errors.Is(err, ErrNotFound) {
			a.sessions.delete(c.Value)
			unauthorized(w)
			return
		}
		if err != nil {
			internalError(w, r, err)
			return
		}
		if len(roles) > 0 && !slices.Contains(roles, u.Role) {
			forbidden(w)
			return
		}
		h(w, r, u)
	}
}
