package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"sitangkap/backend/helpers"
	"sitangkap/backend/middlewares"
	"sitangkap/backend/models"
)

type daftarJSON struct {
	Data []models.TransaksiJSON `json:"data"`
	Meta struct {
		Jumlah         int     `json:"jumlah"`
		Dari           *string `json:"dari"`
		Sampai         *string `json:"sampai"`
		TarifRetribusi float64 `json:"tarif_retribusi"`
	} `json:"meta"`
}

type satuJSON struct {
	Data models.TransaksiJSON `json:"data"`
}

const bodyValid = `{"nama_nelayan":"Pak Roni","jenis_ikan":"Cakalang","berat_kg":120.5,"harga_per_kg":28000,"tanggal":"2026-10-22"}`

// ---------- T-05: daftar ----------

func TestT05DaftarKosongDanUrutan(t *testing.T) {
	u := newUji(t)
	c := u.login("petugas01", "petugas123")

	kosong := u.do(http.MethodGet, "/api/transaksi", "", c)
	if kosong.Code != http.StatusOK || !strings.HasPrefix(kosong.Body.String(), `{"data":[],`) {
		t.Fatalf("daftar kosong: %d %s", kosong.Code, kosong.Body.String())
	}
	if !strings.Contains(kosong.Body.String(), `"dari":null,"sampai":null`) {
		t.Errorf("meta dari/sampai harus null: %s", kosong.Body.String())
	}

	u.store.isiContohKontrak()
	var d daftarJSON
	u.do(http.MethodGet, "/api/transaksi", "", c).json(t, &d)
	var ids []string
	for _, tr := range d.Data {
		ids = append(ids, strconv.FormatInt(tr.ID, 10))
	}
	if strings.Join(ids, ",") != "3,2,1,5,4" {
		t.Errorf("urutan id = %v, mau 3,2,1,5,4 (tanggal terbaru, lalu id terbesar)", ids)
	}
	if d.Meta.Jumlah != 5 || d.Meta.TarifRetribusi != 0.03 {
		t.Errorf("meta = %+v", d.Meta)
	}
}

// ---------- T-06 sampai T-11: tambah, ubah, hapus ----------

func TestT06TambahValid(t *testing.T) {
	u := newUji(t)
	c := u.login("petugas01", "petugas123")
	res := u.do(http.MethodPost, "/api/transaksi", bodyValid, c)
	if res.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", res.Code, res.Body.String())
	}
	var s satuJSON
	res.json(t, &s)
	d := s.Data
	if d.ID == 0 || d.Nilai != 3374000 || d.Retribusi != 101220 || d.DibuatOleh != "petugas01" || d.BeratKg != 120.5 {
		t.Errorf("data = %+v", d)
	}
	// memStore mencatat dibuat_pada 2026-10-22 01:15 UTC; ditampilkan di WITA.
	if d.DibuatPada != "2026-10-22T09:15:00+08:00" {
		t.Errorf("dibuat_pada = %s", d.DibuatPada)
	}
}

func TestT07TigaFieldSalahSekaligus(t *testing.T) {
	u := newUji(t)
	c := u.login("petugas01", "petugas123")
	e := u.do(http.MethodPost, "/api/transaksi",
		`{"nama_nelayan":"","jenis_ikan":"Tuna","berat_kg":-1,"harga_per_kg":28000,"tanggal":"2026-10-23"}`, c).
		harusError(t, http.StatusUnprocessableEntity, "VALIDATION_ERROR", helpers.MsgValidation)
	want := map[string]string{
		"nama_nelayan": "Nama nelayan wajib diisi.",
		"berat_kg":     "Berat harus lebih dari 0.",
		"tanggal":      "Tanggal tidak boleh melewati hari ini.",
	}
	if len(e.Error.Fields) != 3 {
		t.Errorf("fields = %v, mau tepat 3", e.Error.Fields)
	}
	for k, v := range want {
		if e.Error.Fields[k] != v {
			t.Errorf("fields[%s] = %q, mau %q", k, e.Error.Fields[k], v)
		}
	}
}

func TestT08TipeSalahTetap422(t *testing.T) {
	u := newUji(t)
	c := u.login("petugas01", "petugas123")
	kasus := []struct {
		body, field, pesan string
	}{
		{`{"nama_nelayan":"A","jenis_ikan":"B","berat_kg":120.5,"harga_per_kg":28000.5,"tanggal":"2026-10-22"}`, "harga_per_kg", "Harga per kg wajib diisi dengan bilangan bulat."},
		{`{"nama_nelayan":"A","jenis_ikan":"B","berat_kg":"120.5","harga_per_kg":28000,"tanggal":"2026-10-22"}`, "berat_kg", "Berat wajib diisi dengan angka."},
		{`{"nama_nelayan":"A","jenis_ikan":"B","berat_kg":1.234,"harga_per_kg":28000,"tanggal":"2026-10-22"}`, "berat_kg", "Berat paling banyak 2 angka di belakang koma."},
	}
	for _, k := range kasus {
		e := u.do(http.MethodPost, "/api/transaksi", k.body, c).
			harusError(t, http.StatusUnprocessableEntity, "VALIDATION_ERROR", helpers.MsgValidation)
		if e.Error.Fields[k.field] != k.pesan {
			t.Errorf("fields[%s] = %q, mau %q", k.field, e.Error.Fields[k.field], k.pesan)
		}
	}
	// JSON rusak tetap 400.
	u.do(http.MethodPost, "/api/transaksi", `{"nama_nelayan":`, c).harusError(t, http.StatusBadRequest, "BAD_REQUEST", helpers.MsgBadRequest)
}

func TestT09UbahValid(t *testing.T) {
	u := newUji(t)
	u.store.isiContohKontrak()
	c := u.login("petugas01", "petugas123")
	res := u.do(http.MethodPut, "/api/transaksi/1",
		`{"nama_nelayan":"Pak Roni","jenis_ikan":"Tuna","berat_kg":1500,"harga_per_kg":30000,"tanggal":"2026-10-20"}`, c)
	if res.Code != http.StatusOK {
		t.Fatalf("status %d body %s", res.Code, res.Body.String())
	}
	var s satuJSON
	res.json(t, &s)
	if s.Data.JenisIkan != "Tuna" || s.Data.BeratKg != 1500 || s.Data.Nilai != 45000000 || s.Data.Tanggal != "2026-10-20" {
		t.Errorf("data = %+v", s.Data)
	}
	if s.Data.DibuatOleh != "petugas01" || s.Data.DibuatPada != "2026-10-22T09:15:00+08:00" {
		t.Errorf("dibuat_oleh/dibuat_pada berubah: %+v", s.Data)
	}
}

func TestT10IdTidakAdaDanBukanAngka(t *testing.T) {
	u := newUji(t)
	c := u.login("petugas01", "petugas123")
	u.do(http.MethodPut, "/api/transaksi/9999", bodyValid, c).harusError(t, http.StatusNotFound, "NOT_FOUND", helpers.MsgNotFound)
	u.do(http.MethodDelete, "/api/transaksi/9999", "", c).harusError(t, http.StatusNotFound, "NOT_FOUND", helpers.MsgNotFound)
	for _, id := range []string{"abc", "0", "-1", "1.5"} {
		u.do(http.MethodPut, "/api/transaksi/"+id, bodyValid, c).harusError(t, http.StatusBadRequest, "BAD_REQUEST", helpers.MsgBadRequest)
		u.do(http.MethodDelete, "/api/transaksi/"+id, "", c).harusError(t, http.StatusBadRequest, "BAD_REQUEST", helpers.MsgBadRequest)
	}
	// Urutan kontrak: 422 dicek sebelum 404.
	u.do(http.MethodPut, "/api/transaksi/9999", `{}`, c).harusError(t, http.StatusUnprocessableEntity, "VALIDATION_ERROR", helpers.MsgValidation)
}

func TestT11HapusLaluUbah(t *testing.T) {
	u := newUji(t)
	u.store.isiContohKontrak()
	c := u.login("petugas01", "petugas123")
	res := u.do(http.MethodDelete, "/api/transaksi/2", "", c)
	if res.Code != http.StatusNoContent || res.Body.Len() != 0 {
		t.Fatalf("hapus: %d %q", res.Code, res.Body.String())
	}
	u.do(http.MethodPut, "/api/transaksi/2", bodyValid, c).harusError(t, http.StatusNotFound, "NOT_FOUND", helpers.MsgNotFound)
}

// ---------- T-13, T-14: rekap dan rentang ----------

func TestT13Rekap(t *testing.T) {
	u := newUji(t)
	u.store.isiContohKontrak()
	u.store.tambah("Pak Yance", "Tuna", 6000, 63000, "2026-09-15") // di luar rentang
	c := u.login("dinas01", "dinas123")

	for _, path := range []string{"/api/rekap?dari=2026-10-01&sampai=2026-10-22", "/api/rekap"} {
		res := u.do(http.MethodGet, path, "", c)
		var r struct {
			Data models.RekapJSON `json:"data"`
		}
		res.json(t, &r)
		want := models.RekapJSON{Dari: "2026-10-01", Sampai: "2026-10-22", JumlahTransaksi: 5, TotalBeratKg: 370.75,
			TotalNilai: 11577000, TotalRetribusi: 347310, TarifRetribusi: 0.03}
		if res.Code != http.StatusOK || r.Data != want {
			t.Errorf("%s: %d %+v", path, res.Code, r.Data)
		}
	}

	var kosong struct {
		Data models.RekapJSON `json:"data"`
	}
	u.do(http.MethodGet, "/api/rekap?dari=2026-01-01&sampai=2026-01-31", "", c).json(t, &kosong)
	if kosong.Data.JumlahTransaksi != 0 || kosong.Data.TotalNilai != 0 || kosong.Data.TotalBeratKg != 0 {
		t.Errorf("rekap kosong = %+v", kosong.Data)
	}
}

func TestT14RentangSalah(t *testing.T) {
	u := newUji(t)
	c := u.login("dinas01", "dinas123")
	for _, q := range []string{"dari=2026-10-01", "dari=2026-10-22&sampai=2026-10-01", "dari=22-10-2026&sampai=2026-10-22"} {
		u.do(http.MethodGet, "/api/transaksi?"+q, "", c).harusError(t, http.StatusBadRequest, "BAD_REQUEST", helpers.MsgBadRequest)
		u.do(http.MethodGet, "/api/rekap?"+q, "", c).harusError(t, http.StatusBadRequest, "BAD_REQUEST", helpers.MsgBadRequest)
	}
	var d daftarJSON
	u.store.isiContohKontrak()
	u.do(http.MethodGet, "/api/transaksi?dari=2026-10-21&sampai=2026-10-21", "", c).json(t, &d)
	if d.Meta.Jumlah != 2 || d.Meta.Dari == nil || *d.Meta.Dari != "2026-10-21" {
		t.Errorf("rentang 21 Okt: %+v", d.Meta)
	}
}

// ---------- T-15: health ----------

func TestT15Health(t *testing.T) {
	u := newUji(t)
	res := u.do(http.MethodGet, "/health", "", nil)
	if res.Code != http.StatusOK || strings.TrimSpace(res.Body.String()) != `{"status":"ok","db":"up"}` {
		t.Errorf("sehat: %d %s", res.Code, res.Body.String())
	}
	u.store.pingErr = errPingGagal
	res = u.do(http.MethodGet, "/health", "", nil)
	if res.Code != http.StatusServiceUnavailable || strings.TrimSpace(res.Body.String()) != `{"status":"error","db":"down"}` {
		t.Errorf("database mati: %d %s", res.Code, res.Body.String())
	}
}

// ---------- T-16: peran ----------

func TestT16Peran(t *testing.T) {
	u := newUji(t)
	u.store.isiContohKontrak()
	dinas := u.login("dinas01", "dinas123")
	petugas := u.login("petugas01", "petugas123")

	u.do(http.MethodPost, "/api/transaksi", bodyValid, dinas).harusError(t, http.StatusForbidden, "FORBIDDEN", helpers.MsgForbidden)
	u.do(http.MethodPut, "/api/transaksi/1", bodyValid, dinas).harusError(t, http.StatusForbidden, "FORBIDDEN", helpers.MsgForbidden)
	u.do(http.MethodDelete, "/api/transaksi/1", "", dinas).harusError(t, http.StatusForbidden, "FORBIDDEN", helpers.MsgForbidden)
	u.do(http.MethodGet, "/api/rekap", "", petugas).harusError(t, http.StatusForbidden, "FORBIDDEN", helpers.MsgForbidden)

	// Urutan kontrak: 401 sebelum 403, 403 sebelum 400.
	u.do(http.MethodPost, "/api/transaksi", bodyValid, nil).harusError(t, http.StatusUnauthorized, "UNAUTHORIZED", helpers.MsgUnauthorized)
	u.do(http.MethodPut, "/api/transaksi/abc", `{`, dinas).harusError(t, http.StatusForbidden, "FORBIDDEN", helpers.MsgForbidden)

	// Dinas tetap boleh membaca daftar.
	if res := u.do(http.MethodGet, "/api/transaksi", "", dinas); res.Code != http.StatusOK {
		t.Errorf("dinas GET /api/transaksi: %d", res.Code)
	}
}

// ---------- Lain-lain ----------

func TestMethodTidakTerdaftar404(t *testing.T) {
	u := newUji(t)
	c := u.login("petugas01", "petugas123")
	u.do(http.MethodPatch, "/api/transaksi/1", bodyValid, c).harusError(t, http.StatusNotFound, "NOT_FOUND", helpers.MsgNotFound)
}

func TestPanicJadi500(t *testing.T) {
	u := newUji(t)
	h := middlewares.RecoverPanic(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("uji") }))
	u.h = h
	u.do(http.MethodGet, "/api/me", "", nil).harusError(t, http.StatusInternalServerError, "INTERNAL_ERROR", helpers.MsgInternal)
}

func TestFileStatis(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "petugas"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "_astro"), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, "index.html"), []byte("halaman login"), 0o644)
	os.WriteFile(filepath.Join(dir, "petugas", "index.html"), []byte("halaman petugas"), 0o644)
	os.WriteFile(filepath.Join(dir, "_astro", "app.js"), []byte("js"), 0o644)

	u := newUji(t)

	u.h = SetupRoutes(u.handler, u.auth, dir)

	if res := u.do(http.MethodGet, "/", "", nil); res.Code != http.StatusOK || res.Body.String() != "halaman login" {
		t.Errorf("/: %d %q", res.Code, res.Body.String())
	}
	if res := u.do(http.MethodGet, "/petugas/", "", nil); res.Code != http.StatusOK || res.Body.String() != "halaman petugas" {
		t.Errorf("/petugas/: %d %q", res.Code, res.Body.String())
	}
	if res := u.do(http.MethodGet, "/_astro/", "", nil); res.Code != http.StatusNotFound {
		t.Errorf("daftar isi folder harus 404, dapat %d %q", res.Code, res.Body.String())
	}
	// /api dan /health tetap ditangani backend, bukan file statis.
	u.do(http.MethodGet, "/api/tidak-ada", "", nil).harusError(t, http.StatusNotFound, "NOT_FOUND", helpers.MsgNotFound)
	if res := u.do(http.MethodGet, "/health", "", nil); res.Code != http.StatusOK {
		t.Errorf("/health: %d", res.Code)
	}
}
