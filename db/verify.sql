-- db/verify.sql  — jalankan setelah seed: mysql -u root -p sitangkap < db/verify.sql
SELECT username, role, aktif, LENGTH(password_hash) AS panjang_hash FROM users ORDER BY id;

-- Rekap 2026-10-01 s.d. 2026-10-22 dengan tarif 0.03.
-- Harus: 5 transaksi, 370.75 kg, nilai 11577000, retribusi 347310
SELECT COUNT(*)                                       AS jumlah_transaksi,
       SUM(berat_kg)                                  AS total_berat_kg,
       SUM(ROUND(berat_kg * harga_per_kg))            AS total_nilai,
       SUM(ROUND(ROUND(berat_kg * harga_per_kg) * 0.03)) AS total_retribusi
FROM transaksi
WHERE tanggal BETWEEN '2026-10-01' AND '2026-10-22';

-- Constraint harus MENOLAK baris ini. Pesan error MySQL: "Check constraint ... is violated";
-- MariaDB: "CONSTRAINT `chk_transaksi_berat` failed".
-- INSERT INTO transaksi (nama_nelayan, jenis_ikan, berat_kg, harga_per_kg, tanggal, dibuat_oleh)
-- VALUES ('Uji', 'Uji', -1, 1000, '2026-10-01', 1);
