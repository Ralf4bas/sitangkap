package main

import (
	"testing"
	"time"
)

// T-12 · Tabel contoh perhitungan di kontrak API bagian 3 (tarif 0.03).
func TestT12HitungContohKontrak(t *testing.T) {
	kasus := []struct {
		beratCenti, harga, nilai, retribusi int64
	}{
		{12050, 28000, 3374000, 101220},
		{4500, 65000, 2925000, 87750},
		{3025, 22000, 665500, 19965},
		{8000, 25000, 2000000, 60000},
		{9500, 27500, 2612500, 78375},
	}
	var rows []TransaksiRow
	for _, k := range kasus {
		nilai := hitungNilai(k.beratCenti, k.harga)
		if nilai != k.nilai {
			t.Errorf("hitungNilai(%d, %d) = %d, mau %d", k.beratCenti, k.harga, nilai, k.nilai)
		}
		if r := hitungRetribusi(nilai, 0.03); r != k.retribusi {
			t.Errorf("hitungRetribusi(%d) = %d, mau %d", nilai, r, k.retribusi)
		}
		rows = append(rows, TransaksiRow{BeratCenti: k.beratCenti, HargaPerKg: k.harga})
	}

	r := hitungRekap(rows, "2026-10-01", "2026-10-22", 0.03)
	if r.JumlahTransaksi != 5 || r.TotalBeratKg != 370.75 || r.TotalNilai != 11577000 || r.TotalRetribusi != 347310 {
		t.Errorf("rekap = %+v, mau 5 / 370.75 / 11577000 / 347310", r)
	}
	if r.Dari != "2026-10-01" || r.Sampai != "2026-10-22" || r.TarifRetribusi != 0.03 {
		t.Errorf("rentang/tarif rekap salah: %+v", r)
	}
}

func TestHitungPembulatan(t *testing.T) {
	// 0,01 kg × 50 = 0,5 rupiah -> 1 (setengah dibulatkan menjauhi nol); 0,01 × 49 = 0,49 -> 0
	if got := hitungNilai(1, 50); got != 1 {
		t.Errorf("hitungNilai(1, 50) = %d, mau 1", got)
	}
	if got := hitungNilai(1, 49); got != 0 {
		t.Errorf("hitungNilai(1, 49) = %d, mau 0", got)
	}
	// Batas atas: 100000 kg × Rp 100.000.000 tidak meluap di int64.
	if got := hitungNilai(10000000, 100000000); got != 10000000000000 {
		t.Errorf("hitungNilai batas atas = %d", got)
	}
}

func TestRekapKosong(t *testing.T) {
	r := hitungRekap(nil, "2026-10-01", "2026-10-22", 0.03)
	if r.JumlahTransaksi != 0 || r.TotalBeratKg != 0 || r.TotalNilai != 0 || r.TotalRetribusi != 0 {
		t.Errorf("rekap kosong = %+v", r)
	}
}

func TestKeTransaksiJSON(t *testing.T) {
	wita, _ := time.LoadLocation("Asia/Makassar")
	row := TransaksiRow{
		ID: 1, NamaNelayan: "Pak Roni", JenisIkan: "Cakalang", BeratCenti: 12050, HargaPerKg: 28000,
		Tanggal:    time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC),
		DibuatOleh: "petugas01",
		DibuatPada: time.Date(2026, 10, 22, 1, 15, 0, 0, time.UTC),
	}
	j := keTransaksiJSON(row, 0.03, wita)
	if j.BeratKg != 120.5 || j.Tanggal != "2026-10-22" || j.Nilai != 3374000 || j.Retribusi != 101220 {
		t.Errorf("transaksi = %+v", j)
	}
	if j.DibuatPada != "2026-10-22T09:15:00+08:00" {
		t.Errorf("dibuat_pada = %s, mau 2026-10-22T09:15:00+08:00", j.DibuatPada)
	}
}
