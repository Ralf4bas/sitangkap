package services

import (
	"encoding/json"
	"math"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"sitangkap/backend/models"
)

// Pesan validasi persis dari kontrak API (POST /api/transaksi).
const (
	MsgNamaWajib       = "Nama nelayan wajib diisi."
	MsgNamaPanjang     = "Nama nelayan paling banyak 100 karakter."
	MsgIkanWajib       = "Jenis ikan wajib diisi."
	MsgIkanPanjang     = "Jenis ikan paling banyak 60 karakter."
	MsgBeratWajib      = "Berat wajib diisi dengan angka."
	MsgBeratPositif    = "Berat harus lebih dari 0."
	MsgBeratMaks       = "Berat paling besar 100000 kg."
	MsgBeratDesimal    = "Berat paling banyak 2 angka di belakang koma."
	MsgHargaWajib      = "Harga per kg wajib diisi dengan bilangan bulat."
	MsgHargaPositif    = "Harga per kg harus lebih dari 0."
	MsgHargaMaks       = "Harga per kg paling besar Rp 100.000.000."
	MsgTanggalFormat   = "Tanggal harus berformat YYYY-MM-DD dan valid."
	MsgTanggalMelewati = "Tanggal tidak boleh melewati hari ini."
)

// FormatTanggal adalah format tanggal kontrak API (YYYY-MM-DD).
const FormatTanggal = "2006-01-02"

// TransaksiBody di-decode per field ke json.RawMessage, supaya tipe yang salah (misal berat berupa
// string) menjadi pesan validasi 422 per field, bukan 400.
type TransaksiBody struct {
	NamaNelayan json.RawMessage `json:"nama_nelayan"`
	JenisIkan   json.RawMessage `json:"jenis_ikan"`
	BeratKg     json.RawMessage `json:"berat_kg"`
	HargaPerKg  json.RawMessage `json:"harga_per_kg"`
	Tanggal     json.RawMessage `json:"tanggal"`
}

func isMissing(raw json.RawMessage) bool { return len(raw) == 0 || string(raw) == "null" }

// parseTeks memeriksa field teks wajib (BR-09). Panjang dihitung dalam karakter, setelah trim.
func parseTeks(raw json.RawMessage, maks int, msgWajib, msgPanjang string) (string, string) {
	var s string
	if isMissing(raw) || json.Unmarshal(raw, &s) != nil {
		return "", msgWajib
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return "", msgWajib
	}
	if utf8.RuneCountInString(s) > maks {
		return "", msgPanjang
	}
	return s, ""
}

// parseBerat memeriksa berat_kg (BR-02) dan mengembalikannya dalam 0,01 kg.
func parseBerat(raw json.RawMessage) (int64, string) {
	var f float64
	if isMissing(raw) || json.Unmarshal(raw, &f) != nil {
		return 0, MsgBeratWajib
	}
	switch {
	case f <= 0:
		return 0, MsgBeratPositif
	case f > 100000:
		return 0, MsgBeratMaks
	case math.Abs(f*100-math.Round(f*100)) > 1e-6:
		return 0, MsgBeratDesimal
	}
	return int64(math.Round(f * 100)), ""
}

// parseHarga memeriksa harga_per_kg (BR-03). 28000.5, "28000", dan 2.8e4 ditolak sebagai bukan bilangan bulat.
func parseHarga(raw json.RawMessage) (int64, string) {
	var h int64
	if isMissing(raw) || json.Unmarshal(raw, &h) != nil {
		return 0, MsgHargaWajib
	}
	switch {
	case h <= 0:
		return 0, MsgHargaPositif
	case h > 100000000:
		return 0, MsgHargaMaks
	}
	return h, ""
}

// parseTanggalInput memeriksa tanggal (BR-04): tanggal nyata dan tidak setelah today.
func parseTanggalInput(raw json.RawMessage, today time.Time, loc *time.Location) (string, string) {
	var s string
	if isMissing(raw) || json.Unmarshal(raw, &s) != nil {
		return "", MsgTanggalFormat
	}
	t, err := time.ParseInLocation(FormatTanggal, s, loc)
	if err != nil || t.Format(FormatTanggal) != s {
		return "", MsgTanggalFormat
	}
	if s > today.In(loc).Format(FormatTanggal) {
		return "", MsgTanggalMelewati
	}
	return s, ""
}

// ParseTransaksi mengembalikan input bersih, atau map field -> pesan berisi SEMUA field yang salah.
func ParseTransaksi(b TransaksiBody, today time.Time, loc *time.Location) (models.TransaksiInput, map[string]string) {
	var in models.TransaksiInput
	fields := map[string]string{}
	var msg string

	if in.NamaNelayan, msg = parseTeks(b.NamaNelayan, 100, MsgNamaWajib, MsgNamaPanjang); msg != "" {
		fields["nama_nelayan"] = msg
	}
	if in.JenisIkan, msg = parseTeks(b.JenisIkan, 60, MsgIkanWajib, MsgIkanPanjang); msg != "" {
		fields["jenis_ikan"] = msg
	}
	if in.BeratCenti, msg = parseBerat(b.BeratKg); msg != "" {
		fields["berat_kg"] = msg
	}
	if in.HargaPerKg, msg = parseHarga(b.HargaPerKg); msg != "" {
		fields["harga_per_kg"] = msg
	}
	if in.Tanggal, msg = parseTanggalInput(b.Tanggal, today, loc); msg != "" {
		fields["tanggal"] = msg
	}
	if len(fields) > 0 {
		return models.TransaksiInput{}, fields
	}
	return in, nil
}

// ParseRentang membaca query dari dan sampai (BR-07). Keduanya kosong -> ("", "", true).
// Hanya salah satu, format salah, atau dari > sampai -> ok = false (400 BAD_REQUEST).
func ParseRentang(q url.Values) (dari, sampai string, ok bool) {
	dari, sampai = q.Get("dari"), q.Get("sampai")
	if dari == "" && sampai == "" {
		return "", "", true
	}
	if !tanggalValid(dari) || !tanggalValid(sampai) || dari > sampai {
		return "", "", false
	}
	return dari, sampai, true
}

func tanggalValid(s string) bool {
	t, err := time.Parse(FormatTanggal, s)
	return err == nil && t.Format(FormatTanggal) == s
}
