package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"

	"sitangkap/backend/models"
)

// Query dari db/queries.md, persis, dengan placeholder ?.
const (
	q01 = `SELECT id, username, nama, password_hash, role, aktif FROM users WHERE username = ? LIMIT 1`
	q02 = `SELECT id, username, nama, role FROM users WHERE id = ? AND aktif = 1`

	qTransaksiSelect = `SELECT t.id, t.nama_nelayan, t.jenis_ikan, t.berat_kg, t.harga_per_kg, t.tanggal, u.username, t.dibuat_pada FROM transaksi t JOIN users u ON u.id = t.dibuat_oleh`
	qTransaksiUrut   = ` ORDER BY t.tanggal DESC, t.id DESC`
	q03              = qTransaksiSelect + qTransaksiUrut
	q04              = qTransaksiSelect + ` WHERE t.tanggal BETWEEN ? AND ?` + qTransaksiUrut
	q05              = qTransaksiSelect + ` WHERE t.id = ?`

	q06 = `INSERT INTO transaksi (nama_nelayan, jenis_ikan, berat_kg, harga_per_kg, tanggal, dibuat_oleh) VALUES (?, ?, ?, ?, ?, ?)`
	q07 = `UPDATE transaksi SET nama_nelayan = ?, jenis_ikan = ?, berat_kg = ?, harga_per_kg = ?, tanggal = ? WHERE id = ?`
	q08 = `DELETE FROM transaksi WHERE id = ?`
	q09 = `SELECT 1`
)

// MySQLStore adalah implementasi Store dengan MySQL/MariaDB.
type MySQLStore struct {
	db *sql.DB
}

func NewMySQLStore(db *sql.DB) *MySQLStore {
	return &MySQLStore{db: db}
}

// beratKeCenti mengubah DECIMAL(10,2) dari database ("120.50") ke 0,01 kg (12050).
func beratKeCenti(s string) (int64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("berat_kg %q tidak terbaca: %w", s, err)
	}
	return int64(math.Round(f * 100)), nil
}

// centiKeDecimal mengubah 12050 ke "120.50" untuk ditulis ke kolom DECIMAL(10,2).
func centiKeDecimal(c int64) string {
	return fmt.Sprintf("%d.%02d", c/100, c%100)
}

type pemindai interface {
	Scan(dest ...any) error
}

func pindaiTransaksi(p pemindai) (models.TransaksiRow, error) {
	var (
		row   models.TransaksiRow
		berat string
	)
	err := p.Scan(&row.ID, &row.NamaNelayan, &row.JenisIkan, &berat, &row.HargaPerKg, &row.Tanggal, &row.DibuatOleh, &row.DibuatPada)
	if err != nil {
		return row, err
	}
	row.BeratCenti, err = beratKeCenti(berat)
	return row, err
}

func (s *MySQLStore) Ping(ctx context.Context) error {
	var satu int
	return s.db.QueryRowContext(ctx, q09).Scan(&satu)
}

func (s *MySQLStore) FindUserByUsername(ctx context.Context, username string) (models.UserWithHash, error) {
	var u models.UserWithHash
	err := s.db.QueryRowContext(ctx, q01, username).Scan(&u.ID, &u.Username, &u.Nama, &u.PasswordHash, &u.Role, &u.Aktif)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func (s *MySQLStore) FindActiveUserByID(ctx context.Context, id int64) (models.User, error) {
	var u models.User
	err := s.db.QueryRowContext(ctx, q02, id).Scan(&u.ID, &u.Username, &u.Nama, &u.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func (s *MySQLStore) ListTransaksi(ctx context.Context, dari, sampai string) ([]models.TransaksiRow, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if dari == "" {
		rows, err = s.db.QueryContext(ctx, q03)
	} else {
		rows, err = s.db.QueryContext(ctx, q04, dari, sampai)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hasil := make([]models.TransaksiRow, 0)
	for rows.Next() {
		row, err := pindaiTransaksi(rows)
		if err != nil {
			return nil, err
		}
		hasil = append(hasil, row)
	}
	return hasil, rows.Err()
}

func (s *MySQLStore) GetTransaksi(ctx context.Context, id int64) (models.TransaksiRow, error) {
	row, err := pindaiTransaksi(s.db.QueryRowContext(ctx, q05, id))
	if errors.Is(err, sql.ErrNoRows) {
		return row, ErrNotFound
	}
	return row, err
}

func (s *MySQLStore) CreateTransaksi(ctx context.Context, in models.TransaksiInput, userID int64) (int64, error) {
	res, err := s.db.ExecContext(ctx, q06, in.NamaNelayan, in.JenisIkan, centiKeDecimal(in.BeratCenti), in.HargaPerKg, in.Tanggal, userID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *MySQLStore) UpdateTransaksi(ctx context.Context, id int64, in models.TransaksiInput) error {
	res, err := s.db.ExecContext(ctx, q07, in.NamaNelayan, in.JenisIkan, centiKeDecimal(in.BeratCenti), in.HargaPerKg, in.Tanggal, id)
	if err != nil {
		return err
	}
	return cekBarisTerkena(res)
}

func (s *MySQLStore) DeleteTransaksi(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, q08, id)
	if err != nil {
		return err
	}
	return cekBarisTerkena(res)
}

// cekBarisTerkena mengubah "0 baris" menjadi ErrNotFound. Untuk UPDATE ini benar karena DSN memakai
// clientFoundRows=true (baris yang cocok, bukan yang berubah).
func cekBarisTerkena(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

var _ Store = (*MySQLStore)(nil)
