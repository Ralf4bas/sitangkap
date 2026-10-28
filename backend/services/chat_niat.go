package services

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// Langkah 5–7 kontrak POST /api/chat: konteks, rentang tanggal, dan niat.

// Niat data dan metrik dari kontrak.
const (
	NiatBantuan           = "bantuan"
	NiatIkanTerbanyak     = "ikan_terbanyak"
	NiatTransaksiTerakhir = "transaksi_terakhir"
	NiatRekap             = "rekap"
	NiatJumlah            = "jumlah"
	NiatBerat             = "berat"
	NiatNilai             = "nilai"
	NiatRetribusi         = "retribusi"
	NiatGabungan          = "gabungan"
)

// urutanMetrik juga urutan potongan jawaban gabungan.
var urutanMetrik = []string{NiatJumlah, NiatBerat, NiatNilai, NiatRetribusi}

// niatRentang adalah niat data yang memakai rentang; hanya niat ini yang boleh tersimpan di konteks.
var niatRentang = himpunan(NiatRekap, NiatJumlah, NiatBerat, NiatNilai, NiatRetribusi, NiatGabungan, NiatIkanTerbanyak)

// Konteks adalah konteks chat yang sudah divalidasi. String kosong / slice nil berarti null.
type Konteks struct {
	Niat   string
	Metrik []string
	Dari   string
	Sampai string
	Gaya   string
}

// BacaKonteks membaca field konteks dari request. Field yang tidak valid diabaikan satu per satu, tidak pernah error.
func BacaKonteks(raw json.RawMessage) Konteks {
	k := Konteks{Gaya: GayaNetral}
	var m map[string]json.RawMessage
	if isMissing(raw) || json.Unmarshal(raw, &m) != nil {
		return k
	}
	var gaya, niat, dari, sampai string
	if json.Unmarshal(m["gaya"], &gaya) == nil && (gaya == GayaSantai || gaya == GayaFormal || gaya == GayaNetral) {
		k.Gaya = gaya
	}
	if json.Unmarshal(m["niat"], &niat) == nil && niatRentang[niat] {
		k.Niat = niat
	}
	var metrik []string
	if k.Niat == NiatGabungan {
		if json.Unmarshal(m["metrik"], &metrik) == nil && metrikValid(metrik) {
			k.Metrik = metrik
		} else {
			k.Niat = ""
		}
	}
	if json.Unmarshal(m["dari"], &dari) == nil && json.Unmarshal(m["sampai"], &sampai) == nil &&
		tanggalValid(dari) && tanggalValid(sampai) && dari <= sampai {
		k.Dari, k.Sampai = dari, sampai
	}
	return k
}

// metrikValid: dua atau tiga metrik berbeda, urut sesuai urutanMetrik (bentuk yang dikirim server sendiri).
func metrikValid(metrik []string) bool {
	if len(metrik) < 2 || len(metrik) > 3 {
		return false
	}
	i := 0
	for _, m := range urutanMetrik {
		if i < len(metrik) && metrik[i] == m {
			i++
		}
	}
	return i == len(metrik)
}

// Pemahaman adalah hasil membaca satu pesan.
type Pemahaman struct {
	PesanAkhir        string
	Gaya              string
	Niat              string   // "" = tidak dikenali
	Metrik            []string // hanya untuk gabungan
	Dari, Sampai      string   // hanya untuk niat yang memakai rentang
	TanggalTidakValid bool
}

// PahamiPesan menjalankan langkah 1–6. today adalah "hari ini" di zona APP_TIMEZONE.
func PahamiPesan(pesan string, k Konteks, today time.Time) Pemahaman {
	dasar := tokenDasar(pesan)
	tokens := perbaikiTypo(gantiSlang(dasar))
	p := Pemahaman{PesanAkhir: strings.Join(tokens, " "), Gaya: deteksiGaya(dasar, k.Gaya)}
	c := pencocok(tokens)

	dari, sampai, adaPola, valid := polaRentang(tokens, today)
	if !valid {
		p.TanggalTidakValid = true
		return p
	}

	p.Niat, p.Metrik = niatData(c)
	if p.Niat == "" {
		p.Niat = niatPercakapan(c, tokens)
	}
	if p.Niat == "" && adaPola {
		// N6: pertanyaan lanjutan seperti "kalau kemarin?" memakai niat dari konteks.
		p.Niat, p.Metrik = NiatRekap, nil
		if k.Niat != "" {
			p.Niat, p.Metrik = k.Niat, k.Metrik
		}
	}
	if niatRentang[p.Niat] {
		switch {
		case adaPola:
			p.Dari, p.Sampai = dari, sampai
		case k.Dari != "":
			p.Dari, p.Sampai = k.Dari, k.Sampai
		default:
			p.Dari, p.Sampai = awalBulan(today), today.Format(FormatTanggal)
		}
	}
	return p
}

// pencocok mencocokkan kata atau frasa utuh di pesan akhir.
type pencocok []string

func (c pencocok) ada(frasa ...string) bool {
	s := " " + strings.Join(c, " ") + " "
	for _, f := range frasa {
		if strings.Contains(s, " "+f+" ") {
			return true
		}
	}
	return false
}

// niatData menjalankan N1–N5. Urutan lengkap: N1–N5, percakapan P1–P10, lalu N6.
func niatData(c pencocok) (string, []string) {
	switch {
	case c.ada("bantuan", "help", "bisa apa", "contoh", "menu"):
		return NiatBantuan, nil
	case c.ada("ikan") && c.ada("terbanyak", "paling banyak", "terlaris", "laris"):
		return NiatIkanTerbanyak, nil
	case c.ada("terakhir", "terbaru"):
		return NiatTransaksiTerakhir, nil
	}
	cocok := map[string]bool{
		NiatJumlah:    c.ada("transaksi") && c.ada("berapa", "jumlah", "banyak", "ada"),
		NiatBerat:     c.ada("berat", "kg", "kilo", "kilogram"),
		NiatNilai:     c.ada("nilai", "pendapatan", "penjualan", "omzet", "uang"),
		NiatRetribusi: c.ada("retribusi", "pungutan"),
	}
	var metrik []string
	for _, m := range urutanMetrik {
		if cocok[m] {
			metrik = append(metrik, m)
		}
	}
	switch len(metrik) {
	case 0:
		if c.ada("rekap", "ringkasan", "total", "laporan") {
			return NiatRekap, nil
		}
		return "", nil
	case 1:
		return metrik[0], nil
	case 4:
		return NiatRekap, nil
	}
	return NiatGabungan, metrik
}

// niatPercakapan diisi tahap B-15 (P1–P10) lewat init di chat_percakapan.go.
// Tanpa itu, pesan percakapan tidak dikenali.
var niatPercakapan = func(c pencocok, tokens []string) string { return "" }

var namaBulan = map[string]time.Month{"januari": 1, "februari": 2, "maret": 3, "april": 4, "mei": 5, "juni": 6,
	"juli": 7, "agustus": 8, "september": 9, "oktober": 10, "november": 11, "desember": 12}

// polaRentang menjalankan R1–R11 (nama bulan R3 didahulukan dari frasa hari/minggu/bulan/tahun). adaPola = salah satu pola ditemukan; valid = false untuk tanggal tidak nyata
// atau dua tanggal terbalik.
func polaRentang(tokens []string, today time.Time) (dari, sampai string, adaPola, valid bool) {
	var tanggal []string
	for _, t := range tokens {
		if !bentukTanggal(t) {
			continue
		}
		iso, ok := keISO(t)
		if !ok {
			return "", "", true, false
		}
		tanggal = append(tanggal, iso)
	}
	switch len(tanggal) {
	case 0:
	case 1:
		return tanggal[0], tanggal[0], true, true
	default:
		if tanggal[0] > tanggal[1] {
			return "", "", true, false
		}
		return tanggal[0], tanggal[1], true, true
	}

	c := pencocok(tokens)
	f := func(a, b time.Time) (string, string, bool, bool) {
		return a.Format(FormatTanggal), b.Format(FormatTanggal), true, true
	}
	y, m, d := today.Date()
	hari := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	senin := hari.AddDate(0, 0, -((int(hari.Weekday()) + 6) % 7))
	awal := time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
	for i, t := range tokens {
		bln, ok := namaBulan[t]
		if !ok {
			continue
		}
		tahun := y
		if i+1 < len(tokens) && len(tokens[i+1]) == 4 {
			if n, err := strconv.Atoi(tokens[i+1]); err == nil {
				tahun = n
			}
		}
		a := time.Date(tahun, bln, 1, 0, 0, 0, 0, time.UTC)
		return f(a, a.AddDate(0, 1, -1))
	}
	switch {
	case c.ada("hari ini"):
		return f(hari, hari)
	case c.ada("kemarin"):
		return f(hari.AddDate(0, 0, -1), hari.AddDate(0, 0, -1))
	case c.ada("minggu lalu"):
		return f(senin.AddDate(0, 0, -7), senin.AddDate(0, 0, -1))
	case c.ada("minggu ini"):
		return f(senin, hari)
	case c.ada("bulan lalu"):
		return f(awal.AddDate(0, -1, 0), awal.AddDate(0, 0, -1))
	case c.ada("bulan ini"):
		return f(awal, hari)
	case c.ada("tahun lalu"):
		return f(time.Date(y-1, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(y-1, 12, 31, 0, 0, 0, 0, time.UTC))
	case c.ada("tahun ini"):
		return f(time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC), hari)
	}
	return "", "", false, true
}

// bentukTanggal: token berbentuk YYYY-MM-DD atau D/M/YYYY (tanpa memeriksa apakah tanggalnya nyata).
func bentukTanggal(t string) bool {
	if p := strings.Split(t, "-"); len(p) == 3 {
		return angka(p[0], 4, 4) && angka(p[1], 2, 2) && angka(p[2], 2, 2)
	}
	if p := strings.Split(t, "/"); len(p) == 3 {
		return angka(p[0], 1, 2) && angka(p[1], 1, 2) && angka(p[2], 4, 4)
	}
	return false
}

func angka(s string, minLen, maxLen int) bool {
	if len(s) < minLen || len(s) > maxLen {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// keISO mengubah token tanggal ke YYYY-MM-DD; ok = false bila tanggalnya tidak nyata.
func keISO(t string) (string, bool) {
	if p := strings.Split(t, "/"); len(p) == 3 {
		d, _ := strconv.Atoi(p[0])
		m, _ := strconv.Atoi(p[1])
		t = p[2] + "-" + dua(m) + "-" + dua(d)
	}
	return t, tanggalValid(t)
}

func dua(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

func awalBulan(today time.Time) string {
	return today.Format("2006-01") + "-01"
}
