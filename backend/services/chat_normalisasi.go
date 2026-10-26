package services

import "strings"

// Langkah 1–4 kontrak POST /api/chat: normalisasi, gaya bahasa, slang, dan salah ketik.

// Gaya bahasa pengguna.
const (
	GayaSantai = "santai"
	GayaFormal = "formal"
	GayaNetral = "netral"
)

// tokenDasar menjalankan langkah 1: trim, huruf kecil, selain a-z 0-9 / - menjadi spasi, lalu dipecah per spasi.
func tokenDasar(pesan string) []string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(pesan)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '/' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}
	return strings.Fields(b.String())
}

var penandaSantai = himpunan("bro", "gan", "sis", "bang", "bos", "cuy", "coy", "dong", "sih", "nih", "deh",
	"gue", "gw", "lu", "lo", "aja", "banget", "gimana", "gak", "ga", "nggak", "kok", "yuk", "makasih", "thx",
	"mantap", "sip", "brp", "trx", "kmrn")

var penandaFormal = himpunan("anda", "bapak", "ibu", "saudara", "mohon", "silakan", "apakah", "bagaimanakah",
	"permisi", "selamat", "terima")

// deteksiGaya menjalankan langkah 2 pada token sebelum slang diganti. Seri -> gaya konteks, atau netral.
func deteksiGaya(tokens []string, gayaKonteks string) string {
	santai, formal := 0, 0
	for _, t := range tokens {
		switch {
		case penandaSantai[t] || tokenTawa(t) != "":
			santai++
		case penandaFormal[t]:
			formal++
		}
	}
	switch {
	case santai > formal:
		return GayaSantai
	case formal > santai:
		return GayaFormal
	case gayaKonteks != "":
		return gayaKonteks
	}
	return GayaNetral
}

// kamusSlang adalah tabel langkah 3. Penggantinya boleh lebih dari satu kata.
var kamusSlang = map[string]string{}

func init() {
	for ganti, daftar := range map[string][]string{
		"berapa":       {"brp", "brapa", "berapakah"},
		"transaksi":    {"trx", "trans", "transaksinya", "transaksi2"},
		"hari":         {"hr", "hri"},
		"kemarin":      {"kmrn", "kmarin", "kemaren", "kmaren"},
		"hari ini":     {"skrg", "skarang", "sekarang"},
		"bulan":        {"bln"},
		"tahun":        {"thn"},
		"minggu":       {"mgg", "mgu"},
		"tanggal":      {"tgl"},
		"rekap":        {"rekapan", "rekapnya"},
		"retribusi":    {"retribusinya"},
		"berat":        {"beratnya"},
		"nilai":        {"nilainya"},
		"jumlah":       {"jumlahnya"},
		"ikan":         {"ikannya"},
		"uang":         {"duit"},
		"bagaimana":    {"gmn", "gimana"},
		"terima kasih": {"makasih", "makasi", "makacih", "mksh", "thx", "thanks", "tq", "trims", "tengkyu"},
		"oke":          {"ok", "okay", "okey", "okee", "okeh", "sip", "siap", "mantap", "mantab"},
		"tidak":        {"gk", "ga", "gak", "nggak", "ngga", "enggak", "engga", "tdk"},
		"halo":         {"hallo", "haloo", "hello", "helo", "hey", "hei", "hay"},
	} {
		for _, s := range daftar {
			kamusSlang[s] = ganti
		}
	}
}

// tokenTawa mengenali wkwk, haha, dan hehe beserta variasi panjangnya. Bukan tawa -> "".
func tokenTawa(t string) string {
	switch {
	case ulangan(t, "wk", 2):
		return "wkwk"
	case ulangan(t, "ha", 2):
		return "haha"
	case ulangan(t, "he", 2):
		return "hehe"
	}
	return ""
}

// ulangan: t adalah unit yang diulang minimal min kali, boleh diakhiri huruf pertama unit (wkwkw, hahah).
func ulangan(t, unit string, min int) bool {
	if strings.HasSuffix(t, unit[:1]) && len(t)%len(unit) == 1 {
		t = t[:len(t)-1]
	}
	n := len(t) / len(unit)
	return n >= min && strings.Repeat(unit, n) == t
}

// gantiSlang menjalankan langkah 3.
func gantiSlang(tokens []string) []string {
	var out []string
	for _, t := range tokens {
		if g, ok := kamusSlang[t]; ok {
			out = append(out, strings.Fields(g)...)
		} else if tawa := tokenTawa(t); tawa != "" {
			out = append(out, tawa)
		} else {
			out = append(out, t)
		}
	}
	return out
}

// kataKunciTypo adalah daftar langkah 4, berurutan: kata kunci pertama yang berjarak 1 yang dipakai.
var kataKunciTypo = []string{"transaksi", "retribusi", "berapa", "kemarin", "ringkasan", "laporan", "terbanyak",
	"terakhir", "terbaru", "pendapatan", "penjualan", "bantuan", "jumlah", "rekap", "nilai", "berat", "bulan",
	"tahun", "minggu", "januari", "februari", "maret", "april", "agustus", "september", "oktober", "november",
	"desember", "terima", "kasih", "siapa", "kabar", "sampai"}

var himpunanTypo = himpunan(kataKunciTypo...)

// perbaikiTypo menjalankan langkah 4 pada token huruf yang panjangnya 5 atau lebih.
func perbaikiTypo(tokens []string) []string {
	out := make([]string, len(tokens))
	for i, t := range tokens {
		out[i] = t
		if len(t) < 5 || himpunanTypo[t] || !hanyaHuruf(t) {
			continue
		}
		for _, k := range kataKunciTypo {
			if jarakOSA(t, k) == 1 {
				out[i] = k
				break
			}
		}
	}
	return out
}

// NormalisasiPesan mengembalikan pesan akhir (langkah 1, 3, dan 4): "brp trx kmrn" -> "berapa transaksi kemarin".
func NormalisasiPesan(pesan string) string {
	return strings.Join(perbaikiTypo(gantiSlang(tokenDasar(pesan))), " ")
}

// jarakOSA adalah jarak optimal string alignment: Levenshtein ditambah pertukaran dua huruf bersebelahan.
func jarakOSA(a, b string) int {
	d := make([][]int, len(a)+1)
	for i := range d {
		d[i] = make([]int, len(b)+1)
		d[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		d[0][j] = j
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			biaya := 1
			if a[i-1] == b[j-1] {
				biaya = 0
			}
			d[i][j] = min(d[i-1][j]+1, d[i][j-1]+1, d[i-1][j-1]+biaya)
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				d[i][j] = min(d[i][j], d[i-2][j-2]+1)
			}
		}
	}
	return d[len(a)][len(b)]
}

func hanyaHuruf(s string) bool {
	for _, r := range s {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	return true
}

func himpunan(kata ...string) map[string]bool {
	m := make(map[string]bool, len(kata))
	for _, k := range kata {
		m[k] = true
	}
	return m
}
