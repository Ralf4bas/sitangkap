package routes

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"sitangkap/backend/models"
	"sitangkap/backend/repository"
)

// memStore adalah Store palsu di memori, khusus tes. Perilakunya meniru repository.MySQLStore.
type memStore struct {
	mu        sync.Mutex
	users     []models.UserWithHash
	transaksi []memTransaksi
	nextID    int64
	pingErr   error
	now       func() time.Time
}

type memTransaksi struct {
	models.TransaksiInput
	ID         int64
	DibuatOleh int64
	DibuatPada time.Time
}

// Hash bcrypt cost 12 dari db/seed.sql.
const (
	hashPetugas123 = "$2a$12$H5DHNiNE4NyHtRgV/prkP.nyKrq7AFUAMkne3Z1qOJbdGq4DLsXmW"
	hashDinas123   = "$2a$12$IxC8RS2LBIfYBOzw05AT4exQ2RWqXoVe8DX/RR5xVbQ67EQZrgv7C"
)

// newMemStore membuat store dengan dua akun demo dan tanpa transaksi.
func newMemStore() *memStore {
	return &memStore{
		users: []models.UserWithHash{
			{User: models.User{ID: 1, Username: "petugas01", Nama: "Petugas Contoh", Role: "petugas"}, PasswordHash: hashPetugas123, Aktif: true},
			{User: models.User{ID: 2, Username: "dinas01", Nama: "Dinas Contoh", Role: "dinas"}, PasswordHash: hashDinas123, Aktif: true},
		},
		nextID: 1,
		now:    func() time.Time { return time.Date(2026, 10, 22, 1, 15, 0, 0, time.UTC) },
	}
}

// tambah memasukkan transaksi langsung (tanpa validasi), untuk menyiapkan data tes.
func (s *memStore) tambah(nama, ikan string, beratCenti, harga int64, tanggal string) int64 {
	id, err := s.CreateTransaksi(context.Background(), models.TransaksiInput{
		NamaNelayan: nama, JenisIkan: ikan, BeratCenti: beratCenti, HargaPerKg: harga, Tanggal: tanggal,
	}, 1)
	if err != nil {
		panic(err)
	}
	return id
}

// isiContohKontrak memasukkan 5 baris contoh dari kontrak API bagian 3 (urutan sama dengan db/seed.sql).
func (s *memStore) isiContohKontrak() {
	s.tambah("Pak Roni", "Cakalang", 12050, 28000, "2026-10-22")
	s.tambah("Pak Yance", "Tuna", 4500, 65000, "2026-10-22")
	s.tambah("Ibu Merry", "Julung-julung", 3025, 22000, "2026-10-22")
	s.tambah("Pak Hendrik", "Tongkol", 8000, 25000, "2026-10-21")
	s.tambah("Pak Roni", "Cakalang", 9500, 27500, "2026-10-21")
}

// isiSeed memasukkan 8 baris db/seed.sql: 5 contoh kontrak lalu 3 baris bulan sebelumnya.
func (s *memStore) isiSeed() {
	s.isiContohKontrak()
	s.tambah("Pak Yance", "Tuna", 6000, 63000, "2026-09-15")
	s.tambah("Ibu Merry", "Layang", 4000, 18000, "2026-09-10")
	s.tambah("Pak Hendrik", "Cakalang", 11000, 26000, "2026-08-05")
}

func (s *memStore) Ping(ctx context.Context) error { return s.pingErr }

func (s *memStore) FindUserByUsername(ctx context.Context, username string) (models.UserWithHash, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.users {
		if u.Username == username {
			return u, nil
		}
	}
	return models.UserWithHash{}, repository.ErrNotFound
}

func (s *memStore) FindActiveUserByID(ctx context.Context, id int64) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range s.users {
		if u.ID == id && u.Aktif {
			return u.User, nil
		}
	}
	return models.User{}, repository.ErrNotFound
}

func (s *memStore) keBaris(t memTransaksi) models.TransaksiRow {
	tanggal, err := time.Parse("2006-01-02", t.Tanggal)
	if err != nil {
		panic(err)
	}
	username := ""
	for _, u := range s.users {
		if u.ID == t.DibuatOleh {
			username = u.Username
		}
	}
	return models.TransaksiRow{
		ID: t.ID, NamaNelayan: t.NamaNelayan, JenisIkan: t.JenisIkan, BeratCenti: t.BeratCenti,
		HargaPerKg: t.HargaPerKg, Tanggal: tanggal, DibuatOleh: username, DibuatPada: t.DibuatPada,
	}
}

func (s *memStore) ListTransaksi(ctx context.Context, dari, sampai string) ([]models.TransaksiRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	hasil := make([]models.TransaksiRow, 0)
	for _, t := range s.transaksi {
		if dari != "" && (t.Tanggal < dari || t.Tanggal > sampai) {
			continue
		}
		hasil = append(hasil, s.keBaris(t))
	}
	sort.Slice(hasil, func(i, j int) bool {
		if !hasil[i].Tanggal.Equal(hasil[j].Tanggal) {
			return hasil[i].Tanggal.After(hasil[j].Tanggal)
		}
		return hasil[i].ID > hasil[j].ID
	})
	return hasil, nil
}

func (s *memStore) cari(id int64) int {
	for i, t := range s.transaksi {
		if t.ID == id {
			return i
		}
	}
	return -1
}

func (s *memStore) GetTransaksi(ctx context.Context, id int64) (models.TransaksiRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i := s.cari(id)
	if i < 0 {
		return models.TransaksiRow{}, repository.ErrNotFound
	}
	return s.keBaris(s.transaksi[i]), nil
}

func (s *memStore) CreateTransaksi(ctx context.Context, in models.TransaksiInput, userID int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextID
	s.nextID++
	s.transaksi = append(s.transaksi, memTransaksi{TransaksiInput: in, ID: id, DibuatOleh: userID, DibuatPada: s.now().UTC()})
	return id, nil
}

func (s *memStore) UpdateTransaksi(ctx context.Context, id int64, in models.TransaksiInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	i := s.cari(id)
	if i < 0 {
		return repository.ErrNotFound
	}
	s.transaksi[i].TransaksiInput = in
	return nil
}

func (s *memStore) DeleteTransaksi(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	i := s.cari(id)
	if i < 0 {
		return repository.ErrNotFound
	}
	s.transaksi = append(s.transaksi[:i], s.transaksi[i+1:]...)
	return nil
}

var errPingGagal = errors.New("ping gagal")

var _ repository.Store = (*memStore)(nil)

func TestMemStoreUrutanDanNotFound(t *testing.T) {
	s := newMemStore()
	s.isiContohKontrak()
	ctx := context.Background()

	rows, err := s.ListTransaksi(ctx, "", "")
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, r := range rows {
		ids = append(ids, strconv.FormatInt(r.ID, 10))
	}
	if strings.Join(ids, ",") != "3,2,1,5,4" {
		t.Fatalf("ids = %v, mau 3,2,1,5,4", ids)
	}
	rentang, _ := s.ListTransaksi(ctx, "2026-10-21", "2026-10-21")
	if len(rentang) != 2 {
		t.Errorf("rentang 21 Okt = %d baris, mau 2", len(rentang))
	}
	if _, err := s.GetTransaksi(ctx, 99); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("GetTransaksi(99) err = %v", err)
	}
	if err := s.DeleteTransaksi(ctx, 99); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("DeleteTransaksi(99) err = %v", err)
	}
}
