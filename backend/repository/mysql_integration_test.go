package repository

import (
	"context"
	"errors"
	"os"
	"testing"

	"sitangkap/backend/config"
	"sitangkap/backend/models"
)

// TestMySQLStoreIntegrasi hanya berjalan bila SITANGKAP_UJI_MYSQL=1, terhadap database dari backend/.env
// yang sudah di-seed (db/README.md). Tanpa variabel itu tes dilewati, jadi go test tetap tidak butuh MySQL.
// Tes menambah satu transaksi lalu menghapusnya lagi.
func TestMySQLStoreIntegrasi(t *testing.T) {
	if os.Getenv("SITANGKAP_UJI_MYSQL") != "1" {
		t.Skip("set SITANGKAP_UJI_MYSQL=1 untuk menguji terhadap MySQL lokal")
	}
	if err := config.LoadDotEnv("../.env"); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	db, err := config.OpenDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewMySQLStore(db)
	ctx := context.Background()

	if err := s.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	u, err := s.FindUserByUsername(ctx, "petugas01")
	if err != nil || !u.Aktif || len(u.PasswordHash) != 60 || u.Role != "petugas" {
		t.Fatalf("FindUserByUsername = %+v, %v", u, err)
	}
	if _, err := s.FindUserByUsername(ctx, "tidak-ada"); !errors.Is(err, ErrNotFound) {
		t.Errorf("username tidak ada: err = %v", err)
	}
	if _, err := s.FindActiveUserByID(ctx, u.ID); err != nil {
		t.Errorf("FindActiveUserByID: %v", err)
	}

	rows, err := s.ListTransaksi(ctx, "2026-10-01", "2026-10-22")
	if err != nil {
		t.Fatal(err)
	}
	var total int64
	for _, r := range rows {
		total += r.BeratCenti
	}
	if len(rows) != 5 || total != 37075 {
		t.Errorf("Q-04 Oktober = %d baris, %d centi; mau 5 baris, 37075", len(rows), total)
	}

	in := models.TransaksiInput{NamaNelayan: "Uji Integrasi", JenisIkan: "Tuna", BeratCenti: 150000, HargaPerKg: 1000, Tanggal: "2026-10-20"}
	id, err := s.CreateTransaksi(ctx, in, u.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer s.DeleteTransaksi(ctx, id)

	got, err := s.GetTransaksi(ctx, id)
	if err != nil || got.BeratCenti != 150000 || got.DibuatOleh != "petugas01" || got.Tanggal.Format("2006-01-02") != "2026-10-20" {
		t.Fatalf("GetTransaksi = %+v, %v", got, err)
	}
	// Simpan data yang sama persis dua kali: harus tetap sukses (clientFoundRows=true).
	for i := 0; i < 2; i++ {
		if err := s.UpdateTransaksi(ctx, id, in); err != nil {
			t.Fatalf("UpdateTransaksi ke-%d: %v", i+1, err)
		}
	}
	if err := s.DeleteTransaksi(ctx, id); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateTransaksi(ctx, id, in); !errors.Is(err, ErrNotFound) {
		t.Errorf("update setelah hapus: err = %v", err)
	}
}
