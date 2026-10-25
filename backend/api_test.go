package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// TestMain membuang log request supaya output go test hanya berisi hasil tes.
// Jalankan dengan SITANGKAP_LOG_TES=1 untuk melihat log-nya.
func TestMain(m *testing.M) {
	if os.Getenv("SITANGKAP_LOG_TES") != "1" {
		log.SetOutput(io.Discard)
	}
	os.Exit(m.Run())
}

// ---------- Perangkat tes HTTP ----------

// waktuUji = 2026-10-22 10:00 WITA.
var waktuUji = time.Date(2026, 10, 22, 2, 0, 0, 0, time.UTC)

type uji struct {
	t     *testing.T
	app   *App
	store *memStore
	h     http.Handler
}

func newUji(t *testing.T) *uji {
	t.Helper()
	wita, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		t.Fatal(err)
	}
	cfg := Config{Tarif: 0.03, Location: wita, SessionTTL: 8 * time.Hour}
	store := newMemStore()
	app := newApp(store, cfg, func() time.Time { return waktuUji })
	return &uji{t: t, app: app, store: store, h: app.routes()}
}

type respons struct {
	*httptest.ResponseRecorder
}

// do mengirim request; cookie boleh kosong.
func (u *uji) do(method, path, body string, cookie *http.Cookie) respons {
	u.t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	u.h.ServeHTTP(rec, req)
	return respons{rec}
}

// login masuk dengan akun demo dan mengembalikan cookie sesinya.
func (u *uji) login(username, password string) *http.Cookie {
	u.t.Helper()
	res := u.do(http.MethodPost, "/api/login", `{"username":"`+username+`","password":"`+password+`"}`, nil)
	if res.Code != http.StatusOK {
		u.t.Fatalf("login %s: status %d body %s", username, res.Code, res.Body.String())
	}
	return res.cookie(u.t)
}

func (r respons) cookie(t *testing.T) *http.Cookie {
	t.Helper()
	for _, c := range r.Result().Cookies() {
		if c.Name == namaCookie {
			return c
		}
	}
	t.Fatal("tidak ada cookie sitangkap_session")
	return nil
}

func (r respons) json(t *testing.T, dst any) {
	t.Helper()
	if err := json.Unmarshal(r.Body.Bytes(), dst); err != nil {
		t.Fatalf("body bukan JSON (%v): %s", err, r.Body.String())
	}
}

type errorJSON struct {
	Error struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Fields  map[string]string `json:"fields"`
	} `json:"error"`
}

// harusError memeriksa status, kode, dan pesan default dari kontrak.
func (r respons) harusError(t *testing.T, status int, code, message string) errorJSON {
	t.Helper()
	var e errorJSON
	r.json(t, &e)
	if r.Code != status || e.Error.Code != code || e.Error.Message != message {
		t.Fatalf("mau %d %s %q, dapat %d %s", status, code, message, r.Code, r.Body.String())
	}
	return e
}

// ---------- T-01 sampai T-04: login, logout, me ----------

func TestT01LoginPetugas(t *testing.T) {
	u := newUji(t)
	res := u.do(http.MethodPost, "/api/login", `{"username":"petugas01","password":"petugas123"}`, nil)
	if res.Code != http.StatusOK {
		t.Fatalf("status %d body %s", res.Code, res.Body.String())
	}
	var body struct {
		Data map[string]any `json:"data"`
	}
	res.json(t, &body)
	if body.Data["role"] != "petugas" || body.Data["username"] != "petugas01" || body.Data["nama"] != "Petugas Contoh" {
		t.Errorf("data = %v", body.Data)
	}
	if _, ada := body.Data["password_hash"]; ada || strings.Contains(res.Body.String(), "$2a$") {
		t.Error("hash password tidak boleh dikirim")
	}
	setCookie := res.Header().Get("Set-Cookie")
	for _, bagian := range []string{"sitangkap_session=", "Path=/", "Max-Age=28800", "HttpOnly", "SameSite=Lax"} {
		if !strings.Contains(setCookie, bagian) {
			t.Errorf("Set-Cookie %q tidak berisi %q", setCookie, bagian)
		}
	}
	if strings.Contains(setCookie, "Secure") {
		t.Error("Secure hanya bila COOKIE_SECURE=true")
	}
	if c := res.cookie(t); len(c.Value) != 64 {
		t.Errorf("panjang token = %d", len(c.Value))
	}
}

func TestT02LoginGagalPesanSama(t *testing.T) {
	u := newUji(t)
	salah := u.do(http.MethodPost, "/api/login", `{"username":"petugas01","password":"salah"}`, nil)
	tidakAda := u.do(http.MethodPost, "/api/login", `{"username":"tidakada","password":"petugas123"}`, nil)
	salah.harusError(t, http.StatusUnauthorized, "INVALID_CREDENTIALS", msgInvalidCredentials)
	tidakAda.harusError(t, http.StatusUnauthorized, "INVALID_CREDENTIALS", msgInvalidCredentials)
	if salah.Body.String() != tidakAda.Body.String() {
		t.Errorf("body berbeda: %s vs %s", salah.Body.String(), tidakAda.Body.String())
	}
	if salah.Header().Get("Set-Cookie") != "" {
		t.Error("login gagal tidak boleh membuat cookie")
	}
}

func TestLoginAkunNonaktif(t *testing.T) {
	u := newUji(t)
	u.store.users[0].Aktif = false
	u.do(http.MethodPost, "/api/login", `{"username":"petugas01","password":"petugas123"}`, nil).
		harusError(t, http.StatusUnauthorized, "INVALID_CREDENTIALS", msgInvalidCredentials)
}

func TestLoginValidasiDanJSONRusak(t *testing.T) {
	u := newUji(t)
	e := u.do(http.MethodPost, "/api/login", `{"username":"  ","password":""}`, nil).
		harusError(t, http.StatusUnprocessableEntity, "VALIDATION_ERROR", msgValidation)
	if e.Error.Fields["username"] != "Username wajib diisi." || e.Error.Fields["password"] != "Password wajib diisi." {
		t.Errorf("fields = %v", e.Error.Fields)
	}
	u.do(http.MethodPost, "/api/login", `{"username":`, nil).harusError(t, http.StatusBadRequest, "BAD_REQUEST", msgBadRequest)
}

func TestT03MeTanpaCookie(t *testing.T) {
	u := newUji(t)
	u.do(http.MethodGet, "/api/me", "", nil).harusError(t, http.StatusUnauthorized, "UNAUTHORIZED", msgUnauthorized)
	palsu := &http.Cookie{Name: namaCookie, Value: strings.Repeat("0", 64)}
	u.do(http.MethodGet, "/api/me", "", palsu).harusError(t, http.StatusUnauthorized, "UNAUTHORIZED", msgUnauthorized)
}

func TestT04LogoutLaluMe(t *testing.T) {
	u := newUji(t)
	c := u.login("dinas01", "dinas123")

	me := u.do(http.MethodGet, "/api/me", "", c)
	var body struct {
		Data User `json:"data"`
	}
	me.json(t, &body)
	if me.Code != http.StatusOK || body.Data.Username != "dinas01" || body.Data.Role != "dinas" {
		t.Fatalf("me: %d %s", me.Code, me.Body.String())
	}

	out := u.do(http.MethodPost, "/api/logout", "", c)
	if out.Code != http.StatusNoContent || out.Body.Len() != 0 {
		t.Fatalf("logout: %d %q", out.Code, out.Body.String())
	}
	if sc := out.Header().Get("Set-Cookie"); !strings.Contains(sc, "Max-Age=0") || !strings.Contains(sc, "sitangkap_session=;") {
		t.Errorf("Set-Cookie logout = %q", sc)
	}
	u.do(http.MethodGet, "/api/me", "", c).harusError(t, http.StatusUnauthorized, "UNAUTHORIZED", msgUnauthorized)

	// Logout tanpa sesi tetap 204.
	if res := u.do(http.MethodPost, "/api/logout", "", nil); res.Code != http.StatusNoContent {
		t.Errorf("logout tanpa sesi: %d", res.Code)
	}
}

func TestSesiUserDinonaktifkan(t *testing.T) {
	u := newUji(t)
	c := u.login("petugas01", "petugas123")
	u.store.users[0].Aktif = false
	u.do(http.MethodGet, "/api/me", "", c).harusError(t, http.StatusUnauthorized, "UNAUTHORIZED", msgUnauthorized)
	u.store.users[0].Aktif = true
	u.do(http.MethodGet, "/api/me", "", c).harusError(t, http.StatusUnauthorized, "UNAUTHORIZED", msgUnauthorized)
}

func TestCookieSecure(t *testing.T) {
	u := newUji(t)
	u.app.cfg.CookieSecure = true
	res := u.do(http.MethodPost, "/api/login", `{"username":"petugas01","password":"petugas123"}`, nil)
	if !strings.Contains(res.Header().Get("Set-Cookie"), "Secure") {
		t.Errorf("Set-Cookie = %q, mau ada Secure", res.Header().Get("Set-Cookie"))
	}
}

func TestPathAPITidakDikenal(t *testing.T) {
	u := newUji(t)
	u.do(http.MethodGet, "/api/tidak-ada", "", nil).harusError(t, http.StatusNotFound, "NOT_FOUND", msgNotFound)
}
