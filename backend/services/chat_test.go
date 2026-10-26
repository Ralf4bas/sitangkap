package services

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
	"time"

	"sitangkap/backend/models"
)

// hariChat adalah "hari ini" semua tes chat: Kamis, 22 Oktober 2026 (WITA).
var hariChat = time.Date(2026, 10, 22, 10, 0, 0, 0, time.FixedZone("WITA", 8*3600))

// barisSeed adalah 8 transaksi db/seed.sql dengan id sesuai urutan insert.
func barisSeed() []models.TransaksiRow {
	data := []struct {
		nama, ikan   string
		centi, harga int64
		tanggal      string
	}{
		{"Pak Roni", "Cakalang", 12050, 28000, "2026-10-22"},
		{"Pak Yance", "Tuna", 4500, 65000, "2026-10-22"},
		{"Ibu Merry", "Julung-julung", 3025, 22000, "2026-10-22"},
		{"Pak Hendrik", "Tongkol", 8000, 25000, "2026-10-21"},
		{"Pak Roni", "Cakalang", 9500, 27500, "2026-10-21"},
		{"Pak Yance", "Tuna", 6000, 63000, "2026-09-15"},
		{"Ibu Merry", "Layang", 4000, 18000, "2026-09-10"},
		{"Pak Hendrik", "Cakalang", 11000, 26000, "2026-08-05"},
	}
	rows := make([]models.TransaksiRow, len(data))
	for i, d := range data {
		tgl, _ := time.Parse(FormatTanggal, d.tanggal)
		rows[i] = models.TransaksiRow{ID: int64(i + 1), NamaNelayan: d.nama, JenisIkan: d.ikan,
			BeratCenti: d.centi, HargaPerKg: d.harga, Tanggal: tgl, DibuatOleh: "petugas01"}
	}
	return rows
}

// ambil meniru Q-03/Q-04: filter rentang (kosong = semua), urut tanggal terbaru lalu id terbesar.
func ambil(dari, sampai string) []models.TransaksiRow {
	var out []models.TransaksiRow
	for _, r := range barisSeed() {
		t := r.Tanggal.Format(FormatTanggal)
		if dari == "" || (t >= dari && t <= sampai) {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(a, b int) bool {
		if !out[a].Tanggal.Equal(out[b].Tanggal) {
			return out[a].Tanggal.After(out[b].Tanggal)
		}
		return out[a].ID > out[b].ID
	})
	return out
}

// tanya menjalankan alur handler tanpa HTTP: pahami pesan, ambil data, susun jawaban.
func tanya(pesan string, k Konteks, username string) models.JawabanChat {
	p := PahamiPesan(pesan, k, hariChat)
	var rows []models.TransaksiRow
	switch {
	case p.TanggalTidakValid:
	case niatRentang[p.Niat]:
		rows = ambil(p.Dari, p.Sampai)
	case p.Niat == NiatTransaksiTerakhir:
		rows = ambil("", "")
	}
	return SusunJawaban(p, k, rows, 0.03, username)
}

func str(p *string) string {
	if p == nil {
		return "<null>"
	}
	return *p
}

// T-17 · Normalisasi, slang, salah ketik, dan gaya bahasa.
func TestT17Normalisasi(t *testing.T) {
	for masuk, mau := range map[string]string{
		"Berapa  Transaksi, HARI INI?": "berapa transaksi hari ini",
		"brp trx kmrn":                 "berapa transaksi kemarin",
		"total retribsi hr ini":        "total retribusi hari ini",
		"transkasi":                    "transaksi",
		"kgx":                          "kgx",
		"wkwkwk makasih":               "wkwk terima kasih",
		"rekap 01/10/2026":             "rekap 01/10/2026",
	} {
		if got := NormalisasiPesan(masuk); got != mau {
			t.Errorf("NormalisasiPesan(%q) = %q, mau %q", masuk, got, mau)
		}
	}
	if p := PahamiPesan("kgx", Konteks{}, hariChat); p.Niat != "" {
		t.Errorf("kgx tidak boleh cocok dengan metrik berat, dapat niat %q", p.Niat)
	}
	for _, k := range []struct {
		pesan, gayaKonteks, mau string
	}{
		{"halo bro", "", GayaSantai},
		{"Selamat pagi", "", GayaFormal},
		{"rekap", "", GayaNetral},
		{"rekap", GayaFormal, GayaFormal},
		{"mohon rekapnya dong bro", "", GayaSantai},
	} {
		if got := PahamiPesan(k.pesan, Konteks{Gaya: k.gayaKonteks}, hariChat).Gaya; got != k.mau {
			t.Errorf("gaya %q (konteks %q) = %q, mau %q", k.pesan, k.gayaKonteks, got, k.mau)
		}
	}
}

// T-18 · Rentang tanggal R1–R12 dengan hari ini Kamis 2026-10-22.
func TestT18Rentang(t *testing.T) {
	kasus := []struct {
		pesan        string
		k            Konteks
		dari, sampai string
	}{
		{"rekap hari ini", Konteks{}, "2026-10-22", "2026-10-22"},
		{"rekap kemarin", Konteks{}, "2026-10-21", "2026-10-21"},
		{"rekap minggu lalu", Konteks{}, "2026-10-12", "2026-10-18"},
		{"rekap minggu ini", Konteks{}, "2026-10-19", "2026-10-22"},
		{"rekap bulan lalu", Konteks{}, "2026-09-01", "2026-09-30"},
		{"rekap bulan ini", Konteks{}, "2026-10-01", "2026-10-22"},
		{"rekap tahun lalu", Konteks{}, "2025-01-01", "2025-12-31"},
		{"rekap tahun ini", Konteks{}, "2026-01-01", "2026-10-22"},
		{"rekap februari", Konteks{}, "2026-02-01", "2026-02-28"},
		{"rekap oktober 2025", Konteks{}, "2025-10-01", "2025-10-31"},
		{"rekap 15/9/2026", Konteks{}, "2026-09-15", "2026-09-15"},
		{"rekap 2026-09-01 sampai 2026-09-30", Konteks{}, "2026-09-01", "2026-09-30"},
		{"rekap", Konteks{}, "2026-10-01", "2026-10-22"},
		{"rekap", Konteks{Dari: "2026-09-01", Sampai: "2026-09-30"}, "2026-09-01", "2026-09-30"},
	}
	for _, k := range kasus {
		p := PahamiPesan(k.pesan, k.k, hariChat)
		if p.TanggalTidakValid || p.Dari != k.dari || p.Sampai != k.sampai {
			t.Errorf("%q: rentang %s..%s (tidak valid %v), mau %s..%s", k.pesan, p.Dari, p.Sampai, p.TanggalTidakValid, k.dari, k.sampai)
		}
	}
	for _, pesan := range []string{"rekap 2026-02-30", "rekap 31/04/2026", "rekap 22/10/2026 sampai 01/10/2026"} {
		if p := PahamiPesan(pesan, Konteks{}, hariChat); !p.TanggalTidakValid {
			t.Errorf("%q harus tidak valid, dapat %+v", pesan, p)
		}
	}
}

// T-19 · Urutan niat N1–N6.
func TestT19UrutanNiat(t *testing.T) {
	kasus := []struct {
		pesan  string
		k      Konteks
		niat   string
		metrik []string
	}{
		{"halo rekap", Konteks{}, NiatRekap, nil},
		{"total retribusi", Konteks{}, NiatRetribusi, nil},
		{"transaksi terakhir", Konteks{}, NiatTransaksiTerakhir, nil},
		{"ikan terbanyak", Konteks{}, NiatIkanTerbanyak, nil},
		{"kamu bisa apa", Konteks{}, NiatBantuan, nil},
		{"berat sama nilai", Konteks{}, NiatGabungan, []string{NiatBerat, NiatNilai}},
		{"berapa transaksi berat nilai retribusi", Konteks{}, NiatRekap, nil},
		{"ada transaksi hari ini", Konteks{}, NiatJumlah, nil},
		{"bulan ini", Konteks{}, NiatRekap, nil},
		{"kalau kemarin", Konteks{Niat: NiatRetribusi}, NiatRetribusi, nil},
		{"kalau kemarin", Konteks{Niat: NiatGabungan, Metrik: []string{NiatJumlah, NiatBerat}}, NiatGabungan, []string{NiatJumlah, NiatBerat}},
		{"cuaca besok", Konteks{}, "", nil},
	}
	for _, k := range kasus {
		p := PahamiPesan(k.pesan, k.k, hariChat)
		if p.Niat != k.niat || !reflect.DeepEqual(p.Metrik, k.metrik) {
			t.Errorf("%q: niat %q %v, mau %q %v", k.pesan, p.Niat, p.Metrik, k.niat, k.metrik)
		}
	}
}

// T-20 · Tabel "Contoh satu pertanyaan" di kontrak, karakter per karakter.
func TestT20ContohKontrak(t *testing.T) {
	kasus := []struct {
		pesan, niat  string
		metrik       []string
		dari, sampai string
		jawaban      string
	}{
		{"rekap dari 01/10/2026 sampai 22/10/2026", "rekap", nil, "2026-10-01", "2026-10-22",
			"Rekap pada 01/10/2026 sampai 22/10/2026: 5 transaksi, total berat 370,75 kg, nilai Rp 11.577.000, retribusi Rp 347.310 (tarif 3%)."},
		{"Berapa transaksi hari ini?", "jumlah", nil, "2026-10-22", "2026-10-22", "Ada 3 transaksi pada 22/10/2026."},
		{"total retribusi hari ini", "retribusi", nil, "2026-10-22", "2026-10-22", "Total retribusi pada 22/10/2026 adalah Rp 208.935 (tarif 3%)."},
		{"total retribsi hr ini", "retribusi", nil, "2026-10-22", "2026-10-22", "Total retribusi pada 22/10/2026 adalah Rp 208.935 (tarif 3%)."},
		{"brp trx kmrn?", "jumlah", nil, "2026-10-21", "2026-10-21", "Pada 21/10/2026 tercatat 2 transaksi."},
		{"berat ikan kemarin", "berat", nil, "2026-10-21", "2026-10-21", "Pada 21/10/2026 ikan yang terjual totalnya 175 kg."},
		{"nilai september 2026", "nilai", nil, "2026-09-01", "2026-09-30", "Pada 01/09/2026 sampai 30/09/2026 nilai transaksinya mencapai Rp 4.500.000."},
		{"berat sama nilai hari ini", "gabungan", []string{"berat", "nilai"}, "2026-10-22", "2026-10-22", "Pada 22/10/2026: total berat 195,75 kg dan nilai Rp 6.964.500."},
		{"ikan terbanyak bulan ini", "ikan_terbanyak", nil, "2026-10-01", "2026-10-22", "Pada 01/10/2026 sampai 22/10/2026 yang paling banyak adalah Cakalang, totalnya 215,5 kg dari 2 transaksi."},
		{"transaksi terakhir", "transaksi_terakhir", nil, "", "", "Transaksi terbaru: Ibu Merry, Julung-julung 30,25 kg dengan harga Rp 22.000/kg pada 22/10/2026, nilai Rp 665.500."},
		{"rekap januari 2025", "rekap", nil, "2025-01-01", "2025-01-31", "Belum ada transaksi yang tercatat pada 01/01/2025 sampai 31/01/2025."},
		{"rekap 2026-02-30", "", nil, "", "", "Tanggal di pertanyaan tidak valid. Tulis tanggal seperti 2026-10-22 atau 22/10/2026."},
		{"cuaca besok bagaimana", "", nil, "", "", `Pertanyaan itu di luar yang saya pahami. Coba tanyakan data lelang, misalnya "rekap bulan ini".`},
	}
	for _, k := range kasus {
		j := tanya(k.pesan, Konteks{Gaya: GayaNetral}, "petugas01")
		dari, sampai := str(j.Dari), str(j.Sampai)
		if k.dari == "" {
			k.dari, k.sampai = "<null>", "<null>"
		}
		niat := k.niat
		if niat == "" {
			niat = "<null>"
		}
		if j.Jawaban != k.jawaban || str(j.Niat) != niat || !reflect.DeepEqual(j.Metrik, k.metrik) ||
			dari != k.dari || sampai != k.sampai || j.Dikenali != (k.niat != "") {
			t.Errorf("%q:\n dapat %q niat=%s metrik=%v %s..%s dikenali=%v\n mau   %q niat=%s metrik=%v %s..%s",
				k.pesan, j.Jawaban, str(j.Niat), j.Metrik, dari, sampai, j.Dikenali, k.jawaban, niat, k.metrik, k.dari, k.sampai)
		}
	}
}

// Konteks respons: niat data mengganti konteks, tidak dikenali meneruskan konteks lama.
func TestKonteksRespons(t *testing.T) {
	j := tanya("retribusi hari ini dong", Konteks{Gaya: GayaNetral}, "petugas01")
	if str(j.Konteks.Niat) != "retribusi" || str(j.Konteks.Dari) != "2026-10-22" || j.Konteks.Gaya != GayaSantai {
		t.Errorf("konteks setelah retribusi = %+v", j.Konteks)
	}
	lama := Konteks{Niat: NiatRetribusi, Dari: "2026-10-21", Sampai: "2026-10-21", Gaya: GayaSantai}
	j = tanya("cuaca besok", lama, "petugas01")
	if j.Dikenali || str(j.Konteks.Niat) != "retribusi" || str(j.Konteks.Sampai) != "2026-10-21" || j.Konteks.Gaya != GayaSantai {
		t.Errorf("konteks harus diteruskan, dapat %+v", j.Konteks)
	}
	j = tanya("kalau kemarin?", Konteks{Niat: NiatRetribusi, Dari: "2026-10-22", Sampai: "2026-10-22", Gaya: GayaSantai}, "petugas01")
	if j.Jawaban != "Total retribusi pada 21/10/2026 adalah Rp 138.375 (tarif 3%)." {
		t.Errorf("lanjutan kemarin = %q", j.Jawaban)
	}
}

func TestBacaKonteks(t *testing.T) {
	kasus := map[string]Konteks{
		``:     {Gaya: GayaNetral},
		`null`: {Gaya: GayaNetral},
		`"x"`:  {Gaya: GayaNetral},
		`{"niat":5,"dari":"kemarin","gaya":"kasar"}`:                                     {Gaya: GayaNetral},
		`{"niat":"retribusi","dari":"2026-10-21","sampai":"2026-10-21","gaya":"santai"}`: {Niat: "retribusi", Dari: "2026-10-21", Sampai: "2026-10-21", Gaya: GayaSantai},
		`{"niat":"sapaan","dari":"2026-10-22","sampai":"2026-10-21"}`:                    {Gaya: GayaNetral},
		`{"niat":"gabungan","metrik":["berat","nilai"]}`:                                 {Niat: "gabungan", Metrik: []string{"berat", "nilai"}, Gaya: GayaNetral},
		`{"niat":"gabungan","metrik":["nilai","berat"]}`:                                 {Gaya: GayaNetral},
		`{"niat":"berat","metrik":["berat","nilai"],"dari":"2026-10-01"}`:                {Niat: "berat", Gaya: GayaNetral},
	}
	for masuk, mau := range kasus {
		if got := BacaKonteks(json.RawMessage(masuk)); !reflect.DeepEqual(got, mau) {
			t.Errorf("BacaKonteks(%s) = %+v, mau %+v", masuk, got, mau)
		}
	}
}

func TestFormatAngka(t *testing.T) {
	for _, k := range []struct{ got, mau string }{
		{formatRupiah(11577000), "Rp 11.577.000"},
		{formatRupiah(500), "Rp 500"},
		{formatBerat(37075), "370,75 kg"},
		{formatBerat(4500), "45 kg"},
		{formatBerat(12050), "120,5 kg"},
		{formatBerat(150005), "1.500,05 kg"},
		{formatPersen(0.03), "3"},
		{formatPersen(0.0285), "2,85"},
		{formatPersen(0.025), "2,5"},
	} {
		if k.got != k.mau {
			t.Errorf("dapat %q, mau %q", k.got, k.mau)
		}
	}
}
