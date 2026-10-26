package routes

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"sitangkap/backend/helpers"
	"sitangkap/backend/models"
)

type chatJSON struct {
	Data models.JawabanChat `json:"data"`
}

// tanya mengirim POST /api/chat dengan konteks opsional (nil = tanpa field konteks).
func (u *uji) tanya(c *http.Cookie, pesan string, konteks *models.KonteksChat) models.JawabanChat {
	u.t.Helper()
	body := map[string]any{"pesan": pesan}
	if konteks != nil {
		body["konteks"] = konteks
	}
	b, _ := json.Marshal(body)
	res := u.do(http.MethodPost, "/api/chat", string(b), c)
	if res.Code != http.StatusOK {
		u.t.Fatalf("chat %q: status %d %s", pesan, res.Code, res.Body.String())
	}
	var j chatJSON
	res.json(u.t, &j)
	return j.Data
}

// ---------- T-21: sesi, validasi pesan, dan konteks rusak ----------

func TestT21ChatValidasi(t *testing.T) {
	u := newUji(t)
	u.do(http.MethodPost, "/api/chat", `{"pesan":"rekap"}`, nil).
		harusError(t, http.StatusUnauthorized, "UNAUTHORIZED", helpers.MsgUnauthorized)

	c := u.login("petugas01", "petugas123")
	for body, mau := range map[string]string{
		`{"pesan":""}`:    "Pertanyaan wajib diisi.",
		`{"pesan":"   "}`: "Pertanyaan wajib diisi.",
		`{"pesan":5}`:     "Pertanyaan wajib diisi.",
		`{}`:              "Pertanyaan wajib diisi.",
		`{"pesan":"` + strings.Repeat("a", 201) + `"}`: "Pertanyaan paling banyak 200 karakter.",
	} {
		e := u.do(http.MethodPost, "/api/chat", body, c).
			harusError(t, http.StatusUnprocessableEntity, "VALIDATION_ERROR", helpers.MsgValidation)
		if e.Error.Fields["pesan"] != mau {
			t.Errorf("%s: fields.pesan = %q, mau %q", body[:min(len(body), 30)], e.Error.Fields["pesan"], mau)
		}
	}
	// 200 karakter multibyte masih boleh: panjang dihitung per karakter, bukan byte.
	u.tanya(c, strings.Repeat("é", 200), nil)
	u.do(http.MethodPost, "/api/chat", `{"pesan":`, c).
		harusError(t, http.StatusBadRequest, "BAD_REQUEST", helpers.MsgBadRequest)

	res := u.do(http.MethodPost, "/api/chat", `{"pesan":"cuaca besok","konteks":{"niat":5,"dari":"kemarin","gaya":"kasar"}}`, c)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(),
		`"konteks":{"niat":null,"metrik":null,"dari":null,"sampai":null,"gaya":"netral"}`) {
		t.Fatalf("konteks rusak harus diabaikan: %d %s", res.Code, res.Body.String())
	}
}

// ---------- T-22: petugas dan dinas, sama dengan rekap, percakapan berlanjut ----------

func TestT22ChatRekapDanLanjutan(t *testing.T) {
	u := newUji(t)
	u.store.isiSeed()
	mau := "Rekap pada 01/10/2026 sampai 22/10/2026: 5 transaksi, total berat 370,75 kg, nilai Rp 11.577.000, retribusi Rp 347.310 (tarif 3%)."
	for _, akun := range [][2]string{{"petugas01", "petugas123"}, {"dinas01", "dinas123"}} {
		j := u.tanya(u.login(akun[0], akun[1]), "rekap dari 01/10/2026 sampai 22/10/2026", nil)
		if j.Jawaban != mau || !j.Dikenali || *j.Niat != "rekap" {
			t.Errorf("%s: %+v", akun[0], j)
		}
	}

	c := u.login("dinas01", "dinas123")
	var rekap struct {
		Data models.RekapJSON `json:"data"`
	}
	u.do(http.MethodGet, "/api/rekap?dari=2026-10-01&sampai=2026-10-22", "", c).json(t, &rekap)
	if !strings.Contains(mau, "Rp 11.577.000") || rekap.Data.TotalNilai != 11577000 || rekap.Data.TotalRetribusi != 347310 {
		t.Errorf("angka chat harus sama dengan /api/rekap: %+v", rekap.Data)
	}

	res := u.do(http.MethodPost, "/api/chat", `{"pesan":"cuaca besok"}`, c)
	if !strings.Contains(res.Body.String(), `"dikenali":false,"niat":null,"metrik":null,"dari":null,"sampai":null`) {
		t.Errorf("tidak dikenali: %s", res.Body.String())
	}

	// Percakapan B di kontrak, baris 3–5: konteks dikirim balik di setiap pertanyaan.
	formal := "formal"
	k := models.KonteksChat{Gaya: formal}
	for _, langkah := range []struct{ pesan, jawaban string }{
		{"rekap bulan ini", "Rekap pada 01/10/2026 sampai 22/10/2026: 5 transaksi, total berat 370,75 kg, nilai Rp 11.577.000, retribusi Rp 347.310 (tarif 3%)."},
		{"yang retribusinya?", "Retribusi yang terkumpul pada 01/10/2026 sampai 22/10/2026 sebesar Rp 347.310 (tarif 3%)."},
		{"kalau bulan lalu?", "Retribusi yang terkumpul pada 01/09/2026 sampai 30/09/2026 sebesar Rp 135.000 (tarif 3%)."},
	} {
		j := u.tanya(c, langkah.pesan, &k)
		if j.Jawaban != langkah.jawaban {
			t.Errorf("%q:\n dapat %q\n mau   %q", langkah.pesan, j.Jawaban, langkah.jawaban)
		}
		if j.Konteks.Gaya != formal {
			t.Errorf("%q: gaya konteks %q", langkah.pesan, j.Konteks.Gaya)
		}
		k = j.Konteks
	}
	if *k.Niat != "retribusi" || *k.Dari != "2026-09-01" || *k.Sampai != "2026-09-30" {
		t.Errorf("konteks akhir = %+v", k)
	}

	j := u.tanya(c, "transaksi terakhir", nil)
	if j.Jawaban != "Transaksi terbaru: Ibu Merry, Julung-julung 30,25 kg dengan harga Rp 22.000/kg pada 22/10/2026, nilai Rp 665.500." ||
		j.Dari != nil {
		t.Errorf("transaksi terakhir: %+v", j)
	}
}
