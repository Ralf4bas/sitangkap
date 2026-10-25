# Query resmi SiTangkap (D-05)

Backend memakai query di bawah ini **persis**, dengan placeholder `?` (NFR-002). Kalau backend butuh query baru, tambahkan di sini dulu lewat pull request.

## DSN

```text
<DB_USER>:<DB_PASSWORD>@tcp(<DB_HOST>:<DB_PORT>)/<DB_NAME>?charset=utf8mb4&parseTime=true&loc=UTC&time_zone=%27%2B00%3A00%27&clientFoundRows=true
```

- `parseTime=true&loc=UTC&time_zone='+00:00'`: `DATE` dan `TIMESTAMP` dibaca sebagai `time.Time` dalam UTC. Backend mengubah `dibuat_pada` ke `APP_TIMEZONE` saat menampilkan, dan memformat `tanggal` dengan `t.Format("2006-01-02")` tanpa konversi zona.
- `clientFoundRows=true`: `UPDATE` melaporkan baris yang **cocok**, bukan yang berubah. Tanpa ini, menyimpan data yang sama persis terbaca "0 baris" dan backend salah menjawab 404.

Development: `DB_USER=sitangkap_app`, `DB_PASSWORD=sitangkap_dev` (dari `db/app_user.sql`). Password server tidak pernah ditulis di repo.

## Query

### Q-01 · Login

```sql
SELECT id, username, nama, password_hash, role, aktif FROM users WHERE username = ? LIMIT 1
```

### Q-02 · `/api/me`, cek sesi

```sql
SELECT id, username, nama, role FROM users WHERE id = ? AND aktif = 1
```

### Q-03 · Daftar semua transaksi

```sql
SELECT t.id, t.nama_nelayan, t.jenis_ikan, t.berat_kg, t.harga_per_kg, t.tanggal, u.username, t.dibuat_pada
FROM transaksi t JOIN users u ON u.id = t.dibuat_oleh
ORDER BY t.tanggal DESC, t.id DESC
```

### Q-04 · Daftar rentang tanggal, rekap

```sql
SELECT t.id, t.nama_nelayan, t.jenis_ikan, t.berat_kg, t.harga_per_kg, t.tanggal, u.username, t.dibuat_pada
FROM transaksi t JOIN users u ON u.id = t.dibuat_oleh
WHERE t.tanggal BETWEEN ? AND ?
ORDER BY t.tanggal DESC, t.id DESC
```

### Q-05 · Ambil satu transaksi

```sql
SELECT t.id, t.nama_nelayan, t.jenis_ikan, t.berat_kg, t.harga_per_kg, t.tanggal, u.username, t.dibuat_pada
FROM transaksi t JOIN users u ON u.id = t.dibuat_oleh
WHERE t.id = ?
```

### Q-06 · Tambah

```sql
INSERT INTO transaksi (nama_nelayan, jenis_ikan, berat_kg, harga_per_kg, tanggal, dibuat_oleh) VALUES (?, ?, ?, ?, ?, ?)
```

### Q-07 · Ubah

```sql
UPDATE transaksi SET nama_nelayan = ?, jenis_ikan = ?, berat_kg = ?, harga_per_kg = ?, tanggal = ? WHERE id = ?
```

### Q-08 · Hapus

```sql
DELETE FROM transaksi WHERE id = ?
```

### Q-09 · Health

```sql
SELECT 1
```

## Rekap

Rekap **tidak** punya query `SUM` sendiri. Backend menjumlahkan hasil Q-04 di Go: `nilai` dan `retribusi` per baris yang sudah dibulatkan (kontrak API bagian 3). Query `SUM` di `db/verify.sql` hanya untuk memeriksa.
