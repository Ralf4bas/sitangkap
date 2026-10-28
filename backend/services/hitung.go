// Package services berisi logika bisnis tanpa HTTP dan tanpa SQL: perhitungan (BR-05),
// validasi (BR-02..04, BR-07, BR-09), dan sesi login.
package services

import (
	"math"
	"time"

	"sitangkap/backend/models"
)

// HitungNilai = round(berat_kg × harga_per_kg) dalam rupiah.
// beratCenti dan harga selalu positif, jadi (x + 50) / 100 adalah pembulatan .5 ke atas.
func HitungNilai(beratCenti, hargaPerKg int64) int64 {
	return (beratCenti*hargaPerKg + 50) / 100
}

// HitungRetribusi = round(nilai × tarif). math.Round membulatkan .5 menjauhi nol.
func HitungRetribusi(nilai int64, tarif float64) int64 {
	return int64(math.Round(float64(nilai) * tarif))
}

// KeTransaksiJSON melengkapi satu baris dengan nilai dan retribusi.
// tanggal diformat tanpa konversi zona; dibuat_pada ditampilkan di zona APP_TIMEZONE.
func KeTransaksiJSON(row models.TransaksiRow, tarif float64, loc *time.Location) models.TransaksiJSON {
	nilai := HitungNilai(row.BeratCenti, row.HargaPerKg)
	return models.TransaksiJSON{
		ID:          row.ID,
		NamaNelayan: row.NamaNelayan,
		JenisIkan:   row.JenisIkan,
		BeratKg:     float64(row.BeratCenti) / 100,
		HargaPerKg:  row.HargaPerKg,
		Tanggal:     row.Tanggal.Format(FormatTanggal),
		Nilai:       nilai,
		Retribusi:   HitungRetribusi(nilai, tarif),
		DibuatOleh:  row.DibuatOleh,
		DibuatPada:  row.DibuatPada.In(loc).Format(time.RFC3339),
	}
}

// HitungRekap menjumlahkan nilai dan retribusi per baris yang sudah dibulatkan,
// supaya total rekap selalu sama dengan jumlah kolom di tabel.
func HitungRekap(rows []models.TransaksiRow, dari, sampai string, tarif float64) models.RekapJSON {
	r := models.RekapJSON{Dari: dari, Sampai: sampai, JumlahTransaksi: len(rows), TarifRetribusi: tarif}
	var beratCenti int64
	for _, row := range rows {
		nilai := HitungNilai(row.BeratCenti, row.HargaPerKg)
		beratCenti += row.BeratCenti
		r.TotalNilai += nilai
		r.TotalRetribusi += HitungRetribusi(nilai, tarif)
	}
	r.TotalBeratKg = float64(beratCenti) / 100
	return r
}
