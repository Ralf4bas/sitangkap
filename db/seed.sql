-- db/seed.sql
-- DATA DEMO. MENGHAPUS SEMUA DATA LAMA. Jangan dijalankan di server production.
-- Jalankan: mysql -u root -p sitangkap < db/seed.sql

SET FOREIGN_KEY_CHECKS = 0;
TRUNCATE TABLE transaksi;
TRUNCATE TABLE users;
SET FOREIGN_KEY_CHECKS = 1;

INSERT INTO users (username, nama, password_hash, role) VALUES
  ('petugas01', 'Petugas Contoh', '$2a$12$H5DHNiNE4NyHtRgV/prkP.nyKrq7AFUAMkne3Z1qOJbdGq4DLsXmW', 'petugas'),
  ('dinas01',   'Dinas Contoh',   '$2a$12$IxC8RS2LBIfYBOzw05AT4exQ2RWqXoVe8DX/RR5xVbQ67EQZrgv7C', 'dinas');

SET @petugas := (SELECT id FROM users WHERE username = 'petugas01');

INSERT INTO transaksi (nama_nelayan, jenis_ikan, berat_kg, harga_per_kg, tanggal, dibuat_oleh) VALUES
  -- 5 baris yang sama dengan contoh di 01-api-contract.md bagian 3
  ('Pak Roni',    'Cakalang',      120.50, 28000, '2026-10-22', @petugas),
  ('Pak Yance',   'Tuna',           45.00, 65000, '2026-10-22', @petugas),
  ('Ibu Merry',   'Julung-julung',  30.25, 22000, '2026-10-22', @petugas),
  ('Pak Hendrik', 'Tongkol',        80.00, 25000, '2026-10-21', @petugas),
  ('Pak Roni',    'Cakalang',       95.00, 27500, '2026-10-21', @petugas),
  -- data bulan sebelumnya, untuk menguji filter "Bulan ini" dan "Tahun ini"
  ('Pak Yance',   'Tuna',           60.00, 63000, '2026-09-15', @petugas),
  ('Ibu Merry',   'Layang',         40.00, 18000, '2026-09-10', @petugas),
  ('Pak Hendrik', 'Cakalang',      110.00, 26000, '2026-08-05', @petugas);
