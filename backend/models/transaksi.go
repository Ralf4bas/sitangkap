package models

import "time"

// TransaksiInput adalah input yang sudah lolos validasi.
type TransaksiInput struct {
	NamaNelayan string
	JenisIkan   string
	BeratCenti  int64 // berat dalam 0,01 kg: 120.5 kg -> 12050
	HargaPerKg  int64
	Tanggal     string // "YYYY-MM-DD"
}

// TransaksiRow adalah satu baris dari database (Q-03/Q-04/Q-05).
type TransaksiRow struct {
	ID          int64
	NamaNelayan string
	JenisIkan   string
	BeratCenti  int64
	HargaPerKg  int64
	Tanggal     time.Time // DATE, dibaca UTC tengah malam
	DibuatOleh  string    // users.username
	DibuatPada  time.Time // UTC
}

// TransaksiJSON adalah objek Transaksi di kontrak API.
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

// RekapJSON adalah objek Rekap di kontrak API.
type RekapJSON struct {
	Dari            string  `json:"dari"`
	Sampai          string  `json:"sampai"`
	JumlahTransaksi int     `json:"jumlah_transaksi"`
	TotalBeratKg    float64 `json:"total_berat_kg"`
	TotalNilai      int64   `json:"total_nilai"`
	TotalRetribusi  int64   `json:"total_retribusi"`
	TarifRetribusi  float64 `json:"tarif_retribusi"`
}
