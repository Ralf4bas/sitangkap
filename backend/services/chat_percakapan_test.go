package services

import (
	"encoding/json"
	"strings"
	"testing"

	"sitangkap/backend/models"
)

// T-23 · Setiap niat percakapan P1–P10, gaya biasa dan santai.
func TestT23NiatPercakapan(t *testing.T) {
	kasus := []struct{ pesan, niat, gaya string }{
		{"Siapa kamu?", "siapa_kamu", GayaNetral},
		{"kamu bot ya bro", "siapa_kamu", GayaSantai},
		{"Apa kabar?", "apa_kabar", GayaNetral},
		{"apa kabar bro", "apa_kabar", GayaSantai},
		{"Terima kasih", "terima_kasih", GayaFormal},
		{"makasih", "terima_kasih", GayaSantai},
		{"maaf ya", "maaf", GayaNetral},
		{"sorry bro", "maaf", GayaSantai},
		{"sampai jumpa", "pamit", GayaNetral},
		{"bye", "pamit", GayaNetral},
		{"capek nih", "capek", GayaSantai},
		{"saya lelah", "capek", GayaNetral},
		{"wkwkwk", "tawa", GayaSantai},
		{"lol", "tawa", GayaNetral},
		{"sip", "oke", GayaSantai},
		{"oke", "oke", GayaNetral},
		{"gak usah", "tidak", GayaSantai},
		{"tidak", "tidak", GayaNetral},
		{"selamat sore", "sapaan", GayaFormal},
		{"hai", "sapaan", GayaNetral},
	}
	for _, k := range kasus {
		j := tanya(k.pesan, Konteks{Gaya: GayaNetral}, "petugas01")
		p := PahamiPesan(k.pesan, Konteks{Gaya: GayaNetral}, hariChat)
		g := 0
		if k.gaya == GayaSantai {
			g = 1
		}
		v := templatPercakapan[k.niat][g]
		mau := v[indeksVariasi(p.PesanAkhir, len(v))]
		mau = strings.ReplaceAll(mau, "{username}", "petugas01")
		if str(j.Niat) != k.niat || j.Konteks.Gaya != k.gaya || j.Jawaban != mau || !j.Dikenali || j.Dari != nil {
			t.Errorf("%q: niat=%s gaya=%s jawaban=%q, mau niat=%s gaya=%s jawaban=%q",
				k.pesan, str(j.Niat), j.Konteks.Gaya, j.Jawaban, k.niat, k.gaya, mau)
		}
	}
}

// langkahChat adalah satu baris tabel percakapan berlanjut di kontrak.
type langkahChat struct {
	pesan, niat, jawaban         string
	kNiat, kDari, kSampai, kGaya string
}

func jalankanPercakapan(t *testing.T, nama, username string, langkah []langkahChat) {
	t.Helper()
	var k models.KonteksChat
	for i, l := range langkah {
		masuk := Konteks{Gaya: GayaNetral}
		if i > 0 {
			masuk = BacaKonteksJSON(k)
		}
		j := tanya(l.pesan, masuk, username)
		k = j.Konteks
		if str(j.Niat) != l.niat || j.Jawaban != l.jawaban || str(k.Niat) != l.kNiat ||
			str(k.Dari) != l.kDari || str(k.Sampai) != l.kSampai || k.Gaya != l.kGaya {
			t.Errorf("%s %q:\n dapat niat=%s %q konteks=%s %s..%s %s\n mau   niat=%s %q konteks=%s %s..%s %s",
				nama, l.pesan, str(j.Niat), j.Jawaban, str(k.Niat), str(k.Dari), str(k.Sampai), k.Gaya,
				l.niat, l.jawaban, l.kNiat, l.kDari, l.kSampai, l.kGaya)
		}
	}
}

// BacaKonteksJSON meniru frontend: konteks respons dikirim balik apa adanya.
func BacaKonteksJSON(k models.KonteksChat) Konteks {
	b, _ := json.Marshal(k)
	return BacaKonteks(b)
}

// T-24 · Percakapan A dan B di kontrak, baris demi baris.
func TestT24PercakapanKontrak(t *testing.T) {
	const n = "<null>"
	jalankanPercakapan(t, "A", "petugas01", []langkahChat{
		{"halo bro", "sapaan", "Halo petugas01! Si Tangkap di sini. Mau cek rekap atau retribusi hari ini?", n, n, n, GayaSantai},
		{"retribusi hari ini dong", "retribusi", "Total retribusi pada 22/10/2026 adalah Rp 208.935 (tarif 3%).", "retribusi", "2026-10-22", "2026-10-22", GayaSantai},
		{"kalau kemarin?", "retribusi", "Total retribusi pada 21/10/2026 adalah Rp 138.375 (tarif 3%).", "retribusi", "2026-10-21", "2026-10-21", GayaSantai},
		{"makasih bro", "terima_kasih", "Sama-sama! Kalau butuh lagi, panggil aja.", "retribusi", "2026-10-21", "2026-10-21", GayaSantai},
	})
	jalankanPercakapan(t, "B", "dinas01", []langkahChat{
		{"Selamat pagi", "sapaan", `Selamat datang, dinas01. Silakan tanyakan data lelang, misalnya "rekap bulan ini".`, n, n, n, GayaFormal},
		{"Siapa kamu?", "siapa_kamu", "Nama saya Si Tangkap. Saya asisten berbasis aturan untuk data transaksi lelang TPI, jadi jawaban saya selalu diambil dari data yang tercatat.", n, n, n, GayaFormal},
		{"rekap bulan ini", "rekap", "Rekap pada 01/10/2026 sampai 22/10/2026: 5 transaksi, total berat 370,75 kg, nilai Rp 11.577.000, retribusi Rp 347.310 (tarif 3%).", "rekap", "2026-10-01", "2026-10-22", GayaFormal},
		{"yang retribusinya?", "retribusi", "Retribusi yang terkumpul pada 01/10/2026 sampai 22/10/2026 sebesar Rp 347.310 (tarif 3%).", "retribusi", "2026-10-01", "2026-10-22", GayaFormal},
		{"kalau bulan lalu?", "retribusi", "Retribusi yang terkumpul pada 01/09/2026 sampai 30/09/2026 sebesar Rp 135.000 (tarif 3%).", "retribusi", "2026-09-01", "2026-09-30", GayaFormal},
		{"terima kasih", "terima_kasih", "Dengan senang hati. Semoga membantu.", "retribusi", "2026-09-01", "2026-09-30", GayaFormal},
	})
}

// T-25 · Niat data selalu menang atas percakapan, dan jawaban deterministik.
func TestT25PrioritasDanDeterministik(t *testing.T) {
	for pesan, niat := range map[string]string{
		"halo, rekap bulan ini":           "rekap",
		"makasih, retribusi kemarin dong": "retribusi",
		"tidak ada transaksi hari ini?":   "jumlah",
		"halo apa kabar":                  "apa_kabar",
		"kamu bisa apa":                   "bantuan",
	} {
		if j := tanya(pesan, Konteks{Gaya: GayaNetral}, "petugas01"); str(j.Niat) != niat {
			t.Errorf("%q: niat %s, mau %s", pesan, str(j.Niat), niat)
		}
	}
	a := tanya("apa kabar bro", Konteks{Gaya: GayaNetral}, "petugas01")
	b := tanya("apa kabar bro", Konteks{Gaya: GayaNetral}, "petugas01")
	if a.Jawaban != b.Jawaban {
		t.Errorf("jawaban harus sama: %q vs %q", a.Jawaban, b.Jawaban)
	}
}
