package services

import "strings"

// Niat percakapan P1–P10 kontrak POST /api/chat. Diperiksa setelah niat data, tidak memakai data.

// aturanPercakapan berurutan: aturan pertama yang cocok yang dipakai.
var aturanPercakapan = []struct {
	niat  string
	frasa []string
}{
	{"siapa_kamu", []string{"siapa kamu", "kamu siapa", "nama kamu", "namamu", "kamu bot", "kamu robot", "kamu manusia"}},
	{"apa_kabar", []string{"apa kabar", "bagaimana kabar", "kabar kamu", "kabarmu"}},
	{"terima_kasih", []string{"terima kasih"}},
	{"maaf", []string{"maaf", "sorry", "sori"}},
	{"pamit", []string{"dadah", "bye", "sampai jumpa", "pamit", "selamat tinggal"}},
	{"capek", []string{"capek", "cape", "lelah", "pusing", "bosan", "ngantuk"}},
	{"tawa", []string{"wkwk", "haha", "hehe", "lol"}},
	{"oke", []string{"oke"}},
	{"tidak", nil}, // aturan khusus di niatPercakapan
	{"sapaan", []string{"halo", "hai", "hi", "permisi", "assalamualaikum", "selamat pagi", "selamat siang", "selamat sore", "selamat malam"}},
}

// niatPercakapan menjalankan P1–P10. Tidak cocok -> "" (tidak dikenali).
func niatPercakapan(c pencocok, tokens []string) string {
	for _, a := range aturanPercakapan {
		if a.niat == "tidak" {
			if len(tokens) > 0 && len(tokens) <= 3 && tokens[0] == "tidak" {
				return a.niat
			}
			continue
		}
		if c.ada(a.frasa...) {
			return a.niat
		}
	}
	return ""
}

// templatPercakapan: indeks 0 = Biasa (formal dan netral), 1 = Santai; masing-masing dua variasi.
var templatPercakapan = map[string][2][]string{
	"sapaan": {
		{`Halo, {username}! Saya Si Tangkap, asisten data TPI. Mau lihat rekap, jumlah transaksi, berat, nilai, atau retribusi?`,
			`Selamat datang, {username}. Silakan tanyakan data lelang, misalnya "rekap bulan ini".`},
		{`Halo {username}! Si Tangkap di sini. Mau cek rekap atau retribusi hari ini?`,
			`Hai {username}! Ada yang mau dicek? Coba aja "berapa transaksi hari ini".`},
	},
	"siapa_kamu": {
		{`Saya Si Tangkap, asisten yang membantu membaca data lelang ikan di TPI ini. Saya bisa menjawab rekap, jumlah transaksi, berat, nilai, dan retribusi.`,
			`Nama saya Si Tangkap. Saya asisten berbasis aturan untuk data transaksi lelang TPI, jadi jawaban saya selalu diambil dari data yang tercatat.`},
		{`Aku Si Tangkap, asisten data TPI. Tanya aja soal rekap, berat, nilai, atau retribusi.`,
			`Si Tangkap di sini! Kerjaanku bacain data lelang ikan biar kamu nggak perlu hitung manual.`},
	},
	"apa_kabar": {
		{`Kabar baik, terima kasih. Ada data lelang yang ingin Anda lihat?`,
			`Baik, terima kasih sudah bertanya. Mau saya bantu cek rekap hari ini?`},
		{`Baik dong! Kamu gimana? Mau cek data hari ini?`,
			`Aman! Siap bantu cek data lelang nih.`},
	},
	"terima_kasih": {
		{`Sama-sama. Kalau perlu data lain, tanyakan saja.`, `Dengan senang hati. Semoga membantu.`},
		{`Sama-sama! Kalau butuh lagi, panggil aja.`, `Siap, sama-sama!`},
	},
	"maaf": {
		{`Tidak apa-apa. Silakan lanjutkan pertanyaannya.`, `Tidak masalah. Ada yang bisa saya bantu?`},
		{`Santai aja, nggak apa-apa.`, `Aman kok. Lanjut aja.`},
	},
	"pamit": {
		{`Sampai jumpa. Semoga pekerjaannya lancar.`, `Terima kasih sudah memakai Si Tangkap. Sampai jumpa.`},
		{`Dadah! Semangat kerjanya.`, `Oke, sampai nanti!`},
	},
	"capek": {
		{`Semoga cepat pulih. Kalau mau, saya bisa merangkum data supaya pekerjaan Anda lebih ringan.`,
			`Istirahat sebentar tidak apa-apa. Saya siap merangkum data kapan saja.`},
		{`Waduh, istirahat dulu ya. Nanti biar aku yang rangkum datanya.`,
			`Semangat! Biar aku yang hitung, kamu cukup tanya.`},
	},
	"tawa": {
		{`Senang Anda terhibur. Ada data yang ingin dicek?`, `Hehe. Ada lagi yang bisa saya bantu?`},
		{`Wkwk. Lanjut cek data?`, `Haha, ada-ada aja. Mau cek apa lagi?`},
	},
	"oke": {
		{`Baik. Ada hal lain yang ingin Anda tanyakan?`, `Baik, silakan bila ada pertanyaan lain.`},
		{`Sip! Mau cek apa lagi?`, `Oke, gas. Ada lagi?`},
	},
	"tidak": {
		{`Baik, tidak masalah. Saya di sini kalau dibutuhkan.`, `Baik. Silakan tanyakan kapan saja.`},
		{`Oke deh. Kalau butuh, panggil aja.`, `Siap, santai aja.`},
	},
}

// jawabanPercakapan mengisi templat niat percakapan sesuai gaya.
func jawabanPercakapan(p Pemahaman, username string) string {
	g := 0
	if p.Gaya == GayaSantai {
		g = 1
	}
	v := templatPercakapan[p.Niat][g]
	return strings.ReplaceAll(v[indeksVariasi(p.PesanAkhir, len(v))], "{username}", username)
}
