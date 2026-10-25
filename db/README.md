# Database SiTangkap

Panduan menyiapkan database lokal untuk semua anggota tim. Waktu: sekitar 5 menit.

## Prasyarat

MySQL 8.0.16+ atau MariaDB 10.6+ (keduanya menjalankan `CHECK` constraint). Cek:

```bash
mysql --version
```

Di Windows dengan XAMPP, klien ada di `C:\xampp\mysql\bin\mysql.exe` (sesuaikan huruf drive). Jalankan perintah di bawah dari root repo. Kalau user `root` tidak berpassword, tekan Enter saat diminta password.

## Langkah

```bash
mysql -u root -p < db/schema.sql
mysql -u root -p < db/app_user.sql
mysql -u root -p sitangkap < db/seed.sql
mysql -u root -p sitangkap < db/verify.sql
```

| File | Isi |
| --- | --- |
| `schema.sql` | Database `sitangkap`, tabel `users` dan `transaksi`. Aman dijalankan ulang |
| `app_user.sql` | User aplikasi `sitangkap_app` / `sitangkap_dev`, hanya SELECT, INSERT, UPDATE, DELETE. **Khusus development** |
| `seed.sql` | Akun demo dan 8 transaksi contoh. **Menghapus semua data lama** |
| `verify.sql` | Query pemeriksaan |
| `queries.md` | Query resmi yang dipakai backend, dan DSN |

Hasil `verify.sql` yang benar:

```text
panjang_hash = 60 untuk petugas01 dan dinas01
jumlah_transaksi | total_berat_kg | total_nilai | total_retribusi
5                | 370.75         | 11577000    | 347310
```

## Akun demo

| Username | Password | Role |
| --- | --- | --- |
| `petugas01` | `petugas123` | `petugas` |
| `dinas01` | `dinas123` | `dinas` |

Hanya untuk development dan demo. Di server, ganti password-nya (lihat `docs/02-cicd-jenkins.md` CI-02).

## Backend

Isi `backend/.env` (salin dari `backend/.env.example`):

```dotenv
DB_USER=sitangkap_app
DB_PASSWORD=sitangkap_dev
```

## Reset data

Jalankan ulang `db/seed.sql` kapan saja. Semua transaksi kembali ke 8 baris contoh.

## Masalah umum

| Gejala | Solusi |
| --- | --- |
| `Access denied for user 'sitangkap_app'` | `app_user.sql` belum dijalankan, atau `.env` salah ketik |
| `DROP command denied` saat memakai `sitangkap_app` | Memang disengaja: user aplikasi tidak boleh mengubah struktur tabel |
| `CHECK` tidak menolak berat negatif | Versi database terlalu lama; pakai MySQL 8.0.16+ atau MariaDB 10.6+ |
