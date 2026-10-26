package services

import (
	"hash/fnv"
	"math"
	"sort"
	"strconv"
	"strings"

	"sitangkap/backend/models"
)

// Langkah 7–8 kontrak POST /api/chat: konteks respons dan templat jawaban.

// templatData: variasi jawaban niat data, sama untuk semua gaya.
var templatData = map[string][]string{
	"kosong":          {"Tidak ada transaksi {periode}.", "Belum ada transaksi yang tercatat {periode}."},
	NiatRekap:         {"Rekap {periode}: {jumlah} transaksi, total berat {berat}, nilai {nilai}, retribusi {retribusi} (tarif {persen}%).", "{Periode} tercatat {jumlah} transaksi dengan total berat {berat}, nilai {nilai}, dan retribusi {retribusi} (tarif {persen}%)."},
	NiatJumlah:        {"Ada {jumlah} transaksi {periode}.", "{Periode} tercatat {jumlah} transaksi."},
	NiatBerat:         {"Total berat ikan {periode} adalah {berat}.", "{Periode} ikan yang terjual totalnya {berat}."},
	NiatNilai:         {"Total nilai transaksi {periode} adalah {nilai}.", "{Periode} nilai transaksinya mencapai {nilai}."},
	NiatRetribusi:     {"Total retribusi {periode} adalah {retribusi} (tarif {persen}%).", "Retribusi yang terkumpul {periode} sebesar {retribusi} (tarif {persen}%)."},
	NiatIkanTerbanyak: {"Ikan dengan berat terbanyak {periode} adalah {jenis_ikan}: {berat} dari {jumlah} transaksi.", "{Periode} yang paling banyak adalah {jenis_ikan}, totalnya {berat} dari {jumlah} transaksi."},
	NiatGabungan:      {"{Periode}: {bagian}."},
}

const (
	jawabanTerakhir       = "Transaksi terbaru: {nama_nelayan}, {jenis_ikan} {berat} dengan harga {harga_per_kg}/kg pada {tanggal}, nilai {nilai}."
	jawabanBelumAda       = "Belum ada transaksi yang dicatat."
	jawabanTanggalInvalid = "Tanggal di pertanyaan tidak valid. Tulis tanggal seperti 2026-10-22 atau 22/10/2026."
)

// templatGaya: variasi jawaban yang dibedakan gaya. Indeks 0 = Biasa (formal dan netral), 1 = Santai.
var templatGaya = map[string][2][]string{
	NiatBantuan: {
		{`Contoh pertanyaan: "rekap bulan ini", "berapa transaksi hari ini", "total retribusi oktober 2026", "ikan terbanyak minggu ini", "transaksi terakhir".`},
		{`Coba tanya kayak gini: "rekap bulan ini", "brp trx hari ini", "retribusi kemarin", "ikan terbanyak minggu ini", atau "transaksi terakhir".`},
	},
	"tidak_dikenali": {
		{`Maaf, saya belum mengerti pertanyaan itu. Saya bisa membantu soal rekap, jumlah transaksi, berat, nilai, dan retribusi. Ketik "bantuan" untuk contoh.`,
			`Pertanyaan itu di luar yang saya pahami. Coba tanyakan data lelang, misalnya "rekap bulan ini".`},
		{`Hmm, aku belum nangkep maksudnya. Coba tanya soal rekap, berat, nilai, atau retribusi ya. Ketik "bantuan" buat contoh.`,
			`Waduh, itu di luar keahlianku. Kalau soal data lelang, aku jago. Coba "rekap bulan ini".`},
	},
}

// SusunJawaban menjalankan langkah 7–8. rows adalah transaksi di rentang p (niat dengan rentang),
// atau seluruh transaksi urut terbaru (transaksi_terakhir); niat lain mengabaikannya.
func SusunJawaban(p Pemahaman, k Konteks, rows []models.TransaksiRow, tarif float64, username string) models.JawabanChat {
	j := models.JawabanChat{Dikenali: p.Niat != "" && !p.TanggalTidakValid}
	kr := Konteks{Niat: k.Niat, Metrik: k.Metrik, Dari: k.Dari, Sampai: k.Sampai, Gaya: p.Gaya}

	switch {
	case p.TanggalTidakValid:
		j.Jawaban = jawabanTanggalInvalid
	case !j.Dikenali:
		j.Jawaban = pilihGaya("tidak_dikenali", p)
	case niatRentang[p.Niat]:
		j.Jawaban = jawabanRentang(p, rows, tarif)
		j.Metrik = p.Metrik
		j.Dari, j.Sampai = &p.Dari, &p.Sampai
		kr = Konteks{Niat: p.Niat, Metrik: p.Metrik, Dari: p.Dari, Sampai: p.Sampai, Gaya: p.Gaya}
	case p.Niat == NiatTransaksiTerakhir:
		j.Jawaban = jawabanBelumAda
		if len(rows) > 0 {
			j.Jawaban = jawabanTransaksi(rows[0], tarif)
		}
	case p.Niat == NiatBantuan:
		j.Jawaban = pilihGaya(NiatBantuan, p)
	default:
		j.Jawaban = jawabanPercakapan(p, username)
	}
	if j.Dikenali {
		niat := p.Niat
		j.Niat = &niat
	}
	j.Konteks = kr.keJSON()
	return j
}

// jawabanPercakapan diisi tahap B-15 lewat init di chat_percakapan.go.
var jawabanPercakapan = func(p Pemahaman, username string) string { return pilihGaya("tidak_dikenali", p) }

func (k Konteks) keJSON() models.KonteksChat {
	kj := models.KonteksChat{Metrik: k.Metrik, Gaya: k.Gaya}
	if k.Niat != "" {
		kj.Niat = &k.Niat
	}
	if k.Dari != "" {
		kj.Dari, kj.Sampai = &k.Dari, &k.Sampai
	}
	return kj
}

// indeksVariasi = FNV-1a 32-bit pesan akhir modulo jumlah variasi.
func indeksVariasi(pesanAkhir string, n int) int {
	h := fnv.New32a()
	h.Write([]byte(pesanAkhir))
	return int(h.Sum32() % uint32(n))
}

func pilihGaya(kunci string, p Pemahaman) string {
	g := 0
	if p.Gaya == GayaSantai {
		g = 1
	}
	v := templatGaya[kunci][g]
	return v[indeksVariasi(p.PesanAkhir, len(v))]
}

func jawabanRentang(p Pemahaman, rows []models.TransaksiRow, tarif float64) string {
	r := HitungRekap(rows, p.Dari, p.Sampai, tarif)
	isi := map[string]string{
		"jumlah":    strconv.Itoa(r.JumlahTransaksi),
		"berat":     formatBerat(int64(math.Round(r.TotalBeratKg * 100))),
		"nilai":     formatRupiah(r.TotalNilai),
		"retribusi": formatRupiah(r.TotalRetribusi),
		"persen":    formatPersen(tarif),
	}
	kunci := p.Niat
	switch {
	case len(rows) == 0:
		kunci = "kosong"
	case p.Niat == NiatIkanTerbanyak:
		jenis, beratCenti, jumlah := ikanTerbanyak(rows)
		isi["jenis_ikan"], isi["berat"], isi["jumlah"] = jenis, formatBerat(beratCenti), strconv.Itoa(jumlah)
	case p.Niat == NiatGabungan:
		isi["bagian"] = bagianGabungan(p.Metrik, isi)
	}
	v := templatData[kunci]
	return isiTemplat(v[indeksVariasi(p.PesanAkhir, len(v))], periode(p.Dari, p.Sampai), isi)
}

// bagianGabungan: "A dan B" atau "A, B, dan C", urut jumlah, berat, nilai, retribusi.
func bagianGabungan(metrik []string, isi map[string]string) string {
	potong := map[string]string{
		NiatJumlah:    isi["jumlah"] + " transaksi",
		NiatBerat:     "total berat " + isi["berat"],
		NiatNilai:     "nilai " + isi["nilai"],
		NiatRetribusi: "retribusi " + isi["retribusi"] + " (tarif " + isi["persen"] + "%)",
	}
	var b []string
	for _, m := range metrik {
		b = append(b, potong[m])
	}
	if len(b) == 2 {
		return b[0] + " dan " + b[1]
	}
	return strings.Join(b[:len(b)-1], ", ") + ", dan " + b[len(b)-1]
}

// ikanTerbanyak: jenis ikan dengan total berat terbesar; seri -> abjad paling awal.
func ikanTerbanyak(rows []models.TransaksiRow) (string, int64, int) {
	berat := map[string]int64{}
	jumlah := map[string]int{}
	for _, r := range rows {
		berat[r.JenisIkan] += r.BeratCenti
		jumlah[r.JenisIkan]++
	}
	jenis := make([]string, 0, len(berat))
	for j := range berat {
		jenis = append(jenis, j)
	}
	sort.Slice(jenis, func(a, b int) bool {
		if berat[jenis[a]] != berat[jenis[b]] {
			return berat[jenis[a]] > berat[jenis[b]]
		}
		return jenis[a] < jenis[b]
	})
	return jenis[0], berat[jenis[0]], jumlah[jenis[0]]
}

func jawabanTransaksi(r models.TransaksiRow, tarif float64) string {
	return isiTemplat(jawabanTerakhir, "", map[string]string{
		"nama_nelayan": r.NamaNelayan,
		"jenis_ikan":   r.JenisIkan,
		"berat":        formatBerat(r.BeratCenti),
		"harga_per_kg": formatRupiah(r.HargaPerKg),
		"tanggal":      formatTanggal(r.Tanggal.Format(FormatTanggal)),
		"nilai":        formatRupiah(HitungNilai(r.BeratCenti, r.HargaPerKg)),
	})
}

// periode: "pada 22/10/2026" atau "pada 01/10/2026 sampai 22/10/2026".
func periode(dari, sampai string) string {
	if dari == sampai {
		return "pada " + formatTanggal(dari)
	}
	return "pada " + formatTanggal(dari) + " sampai " + formatTanggal(sampai)
}

func isiTemplat(t, periode string, isi map[string]string) string {
	pasangan := []string{"{periode}", periode, "{Periode}", "P" + strings.TrimPrefix(periode, "p")}
	for k, v := range isi {
		pasangan = append(pasangan, "{"+k+"}", v)
	}
	return strings.NewReplacer(pasangan...).Replace(t)
}

// formatRupiah: 11577000 -> "Rp 11.577.000".
func formatRupiah(n int64) string {
	return "Rp " + ribuan(n)
}

// formatBerat dari 0,01 kg: 37075 -> "370,75 kg", 4500 -> "45 kg", 150000 -> "1.500 kg".
func formatBerat(centi int64) string {
	s := ribuan(centi / 100)
	if sisa := centi % 100; sisa != 0 {
		s += "," + strings.TrimRight(dua(int(sisa)), "0")
	}
	return s + " kg"
}

// formatPersen sama dengan frontend: 0.03 -> "3", 0.0285 -> "2,85".
func formatPersen(tarif float64) string {
	bp := int64(math.Round(tarif * 10000))
	s := strconv.FormatInt(bp/100, 10)
	if sisa := bp % 100; sisa != 0 {
		s += "," + strings.TrimRight(dua(int(sisa)), "0")
	}
	return s
}

// formatTanggal: "2026-10-22" -> "22/10/2026".
func formatTanggal(iso string) string {
	return iso[8:10] + "/" + iso[5:7] + "/" + iso[0:4]
}

func ribuan(n int64) string {
	s := strconv.FormatInt(n, 10)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return s
}
