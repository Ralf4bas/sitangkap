package main

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
	"time"
)

func hariUji(t *testing.T) (time.Time, *time.Location) {
	t.Helper()
	wita, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		t.Fatal(err)
	}
	return time.Date(2026, 10, 22, 10, 0, 0, 0, wita), wita
}

func bodyDari(t *testing.T, s string) transaksiBody {
	t.Helper()
	var b transaksiBody
	if err := json.Unmarshal([]byte(s), &b); err != nil {
		t.Fatal(err)
	}
	return b
}

func TestParseTransaksiValid(t *testing.T) {
	today, loc := hariUji(t)
	in, fields := parseTransaksi(bodyDari(t,
		`{"nama_nelayan":"  Pak Roni ","jenis_ikan":"Cakalang","berat_kg":120.5,"harga_per_kg":28000,"tanggal":"2026-10-22","dibuat_oleh":"diabaikan"}`),
		today, loc)
	if fields != nil {
		t.Fatalf("fields = %v", fields)
	}
	want := TransaksiInput{NamaNelayan: "Pak Roni", JenisIkan: "Cakalang", BeratCenti: 12050, HargaPerKg: 28000, Tanggal: "2026-10-22"}
	if in != want {
		t.Errorf("input = %+v, mau %+v", in, want)
	}
}

func TestParseTransaksiPesan(t *testing.T) {
	today, loc := hariUji(t)
	kasus := []struct {
		nama  string
		body  string
		field string
		pesan string
	}{
		{"nama kosong", `{"nama_nelayan":"   "}`, "nama_nelayan", msgNamaWajib},
		{"nama tidak ada", `{}`, "nama_nelayan", msgNamaWajib},
		{"nama 101 karakter", `{"nama_nelayan":"` + strings.Repeat("a", 101) + `"}`, "nama_nelayan", msgNamaPanjang},
		{"ikan angka", `{"jenis_ikan":5}`, "jenis_ikan", msgIkanWajib},
		{"ikan 61 karakter", `{"jenis_ikan":"` + strings.Repeat("é", 61) + `"}`, "jenis_ikan", msgIkanPanjang},
		{"berat string", `{"berat_kg":"120.5"}`, "berat_kg", msgBeratWajib},
		{"berat null", `{"berat_kg":null}`, "berat_kg", msgBeratWajib},
		{"berat negatif", `{"berat_kg":-1}`, "berat_kg", msgBeratPositif},
		{"berat nol", `{"berat_kg":0}`, "berat_kg", msgBeratPositif},
		{"berat terlalu besar", `{"berat_kg":100000.01}`, "berat_kg", msgBeratMaks},
		{"berat 3 desimal", `{"berat_kg":1.234}`, "berat_kg", msgBeratDesimal},
		{"harga pecahan", `{"harga_per_kg":28000.5}`, "harga_per_kg", msgHargaWajib},
		{"harga eksponen", `{"harga_per_kg":2.8e4}`, "harga_per_kg", msgHargaWajib},
		{"harga string", `{"harga_per_kg":"28000"}`, "harga_per_kg", msgHargaWajib},
		{"harga nol", `{"harga_per_kg":0}`, "harga_per_kg", msgHargaPositif},
		{"harga terlalu besar", `{"harga_per_kg":100000001}`, "harga_per_kg", msgHargaMaks},
		{"tanggal tidak nyata", `{"tanggal":"2026-02-30"}`, "tanggal", msgTanggalFormat},
		{"tanggal format lain", `{"tanggal":"22-10-2026"}`, "tanggal", msgTanggalFormat},
		{"tanggal besok", `{"tanggal":"2026-10-23"}`, "tanggal", msgTanggalMelewati},
	}
	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			_, fields := parseTransaksi(bodyDari(t, k.body), today, loc)
			if got := fields[k.field]; got != k.pesan {
				t.Errorf("fields[%s] = %q, mau %q", k.field, got, k.pesan)
			}
		})
	}
}

func TestParseTransaksiBatasValid(t *testing.T) {
	today, loc := hariUji(t)
	in, fields := parseTransaksi(bodyDari(t,
		`{"nama_nelayan":"`+strings.Repeat("a", 100)+`","jenis_ikan":"`+strings.Repeat("b", 60)+`","berat_kg":100000,"harga_per_kg":100000000,"tanggal":"2026-10-22"}`),
		today, loc)
	if fields != nil {
		t.Fatalf("nilai batas harus diterima, fields = %v", fields)
	}
	if in.BeratCenti != 10000000 {
		t.Errorf("BeratCenti = %d", in.BeratCenti)
	}
	_, fields = parseTransaksi(bodyDari(t, `{"nama_nelayan":"a","jenis_ikan":"b","berat_kg":0.29,"harga_per_kg":1,"tanggal":"2026-10-22"}`), today, loc)
	if fields != nil {
		t.Errorf("0.29 kg harus valid, fields = %v", fields)
	}
}

func TestParseRentang(t *testing.T) {
	kasus := []struct {
		query        string
		dari, sampai string
		ok           bool
	}{
		{"", "", "", true},
		{"dari=2026-10-01&sampai=2026-10-22", "2026-10-01", "2026-10-22", true},
		{"dari=2026-10-22&sampai=2026-10-22", "2026-10-22", "2026-10-22", true},
		{"dari=2026-10-01", "", "", false},
		{"sampai=2026-10-01", "", "", false},
		{"dari=2026-10-22&sampai=2026-10-01", "", "", false},
		{"dari=22-10-2026&sampai=2026-10-22", "", "", false},
		{"dari=2026-02-30&sampai=2026-03-01", "", "", false},
	}
	for _, k := range kasus {
		q, _ := url.ParseQuery(k.query)
		dari, sampai, ok := parseRentang(q)
		if dari != k.dari || sampai != k.sampai || ok != k.ok {
			t.Errorf("parseRentang(%q) = %q, %q, %v", k.query, dari, sampai, ok)
		}
	}
}
