package main

import (
	"math"
	"time"
)

// Semua perhitungan uang (BR-05) ada di file ini. Database hanya menyimpan input.

// hitungNilai = round(berat_kg × harga_per_kg) dalam rupiah.
// beratCenti dan harga selalu positif, jadi (x + 50) / 100 adalah pembulatan .5 ke atas.
func hitungNilai(beratCenti, hargaPerKg int64) int64 {
	return (beratCenti*hargaPerKg + 50) / 100
}

// hitungRetribusi = round(nilai × tarif). math.Round membulatkan .5 menjauhi nol.
func hitungRetribusi(nilai int64, tarif float64) int64 {
	return int64(math.Round(float64(nilai) * tarif))
}

type TransaksiJSON struct {
	ID          int64   `json:"id"`
	NamaNelayan string  `json:"nama_nelayan"`
	JenisIkan   string  `json:"jenis_ikan"`
	BeratKg     float64 `json:"berat_kg"`
	HargaPerKg  int64   `json:"harga_per_kg"`
	Tanggal     string  `json:"tanggal"`
	Nilai       int64   `json:"nilai"`
	Retribusi   int64   `json:"retribusi"`
	DibuatOleh  string  `json:"dibuat_oleh"`
	DibuatPada  string  `json:"dibuat_pada"`
}

type RekapJSON struct {
	Dari            string  `json:"dari"`
	Sampai          string  `json:"sampai"`
	JumlahTransaksi int     `json:"jumlah_transaksi"`
	TotalBeratKg    float64 `json:"total_berat_kg"`
	TotalNilai      int64   `json:"total_nilai"`
	TotalRetribusi  int64   `json:"total_retribusi"`
	TarifRetribusi  float64 `json:"tarif_retribusi"`
}

// keTransaksiJSON melengkapi satu baris dengan nilai dan retribusi.
// tanggal diformat tanpa konversi zona; dibuat_pada ditampilkan di zona APP_TIMEZONE.
func keTransaksiJSON(row TransaksiRow, tarif float64, loc *time.Location) TransaksiJSON {
	nilai := hitungNilai(row.BeratCenti, row.HargaPerKg)
	return TransaksiJSON{
		ID:          row.ID,
		NamaNelayan: row.NamaNelayan,
		JenisIkan:   row.JenisIkan,
		BeratKg:     float64(row.BeratCenti) / 100,
		HargaPerKg:  row.HargaPerKg,
		Tanggal:     row.Tanggal.Format("2006-01-02"),
		Nilai:       nilai,
		Retribusi:   hitungRetribusi(nilai, tarif),
		DibuatOleh:  row.DibuatOleh,
		DibuatPada:  row.DibuatPada.In(loc).Format(time.RFC3339),
	}
}

// hitungRekap menjumlahkan nilai dan retribusi per baris yang sudah dibulatkan,
// supaya total rekap selalu sama dengan jumlah kolom di tabel.
func hitungRekap(rows []TransaksiRow, dari, sampai string, tarif float64) RekapJSON {
	r := RekapJSON{Dari: dari, Sampai: sampai, JumlahTransaksi: len(rows), TarifRetribusi: tarif}
	var beratCenti int64
	for _, row := range rows {
		nilai := hitungNilai(row.BeratCenti, row.HargaPerKg)
		beratCenti += row.BeratCenti
		r.TotalNilai += nilai
		r.TotalRetribusi += hitungRetribusi(nilai, tarif)
	}
	r.TotalBeratKg = float64(beratCenti) / 100
	return r
}
