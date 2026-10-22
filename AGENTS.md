# AGENTS.md · Aturan untuk semua AI coding agent

Kamu bekerja di repo **SiTangkap**: aplikasi web pencatatan transaksi lelang ikan di TPI. Tim terdiri dari empat orang, dan setiap orang memakai AI agent untuk perannya masing-masing.

## 1. Urutan baca (wajib sebelum menulis kode)

1. File ini.
2. `docs/00-overview.md`: requirement, aturan bisnis, ruang lingkup.
   Lalu `docs/03-git-workflow.md`: aturan branch, commit, dan pull request.
3. `docs/01-api-contract.md`: **sumber kebenaran tertinggi** untuk API, field, error, konfigurasi.
4. Dokumen peran pengguna yang sedang kamu bantu:

| Peran | Orang | Dokumen | Folder yang boleh diubah |
| --- | --- | --- | --- |
| Project Manager | tmthy | `docs/roles/pm-tmthy.md` | `docs/`, `AGENTS.md`, `CLAUDE.md`, `README.md`, `Jenkinsfile`, `deploy/`, `.github/`, `.githooks/`, `.gitignore` |
| Frontend | Magen | `docs/roles/frontend-magen.md` | `frontend/` |
| Backend | Rafli | `docs/roles/backend-rafli.md` | `backend/` |
| Database | abas | `docs/roles/database-abas.md` | `db/` |

Kalau pengguna tidak menyebut perannya, **tanyakan dulu** sebelum mengubah file apa pun.

## 2. Aturan yang berlaku untuk semua peran

1. **Ikuti kontrak persis.** Jangan menambah, mengganti nama, atau menghapus endpoint, field, kode status, kode error, atau pesan error yang tertulis di `docs/01-api-contract.md`.
2. **Ubah hanya folder milik peranmu.** Butuh perubahan di folder lain? Tulis usulannya untuk pengguna (dia yang membuat issue), jangan diubah sendiri.
3. **Tidak menambah dependensi** selain yang disebut di dokumen peran:
   - backend: `github.com/go-sql-driver/mysql`, `golang.org/x/crypto/bcrypt`
   - frontend: `astro` saja
   - database: SQL murni
4. **Jangan menebak.** Kalau dokumen tidak menjawab, berhenti dan ajukan pertanyaan spesifik kepada pengguna.
5. **Tidak ada rahasia di repo.** File `.env` tidak pernah di-commit. Password production tidak pernah ditulis ke file mana pun di repo.
6. **Bahasa:** teks yang dilihat pengguna aplikasi dalam Bahasa Indonesia. Nama field API dan kolom database sesuai kontrak (`snake_case` Bahasa Indonesia). Nama fungsi dan variabel kode boleh Bahasa Indonesia atau Inggris, asal konsisten dalam satu file.
7. **Kerjakan satu tahap per sesi** (ID seperti `B-07`, `FE-06`, `D-04`) dan sebutkan ID itu di ringkasan dan pesan commit.
8. **Tunjukkan bukti.** Sebelum menyatakan selesai, jalankan perintah verifikasi peranmu dan tampilkan output-nya.

## 3. Perintah verifikasi

| Peran | Perintah |
| --- | --- |
| Backend | `cd backend && go vet ./... && go test ./...` |
| Frontend | `cd frontend && npm run build` |
| Database | `mysql -u root -p < db/schema.sql` (dua kali), `mysql -u root -p < db/app_user.sql`, `mysql -u root -p sitangkap < db/seed.sql`, `mysql -u root -p sitangkap < db/verify.sql` |
| Semua (integrasi) | Perintah curl di `docs/01-api-contract.md` bagian 7 |

## 4. Git (detail lengkap: `docs/03-git-workflow.md`)

- **Dilarang push langsung ke `main`, `dev`, atau `master`.** Branch tugas dibuat dari `dev` dan masuk ke `dev` lewat pull request dengan minimal 1 approval. Rilis `dev` → `main` hanya dilakukan PM.
- Branch: `<peran>/<nama>-<id>-<ringkasan>`, huruf kecil tanpa spasi, contoh `backend/rafli-b07-login-sesi`.
- Pesan commit: `<jenis>(<peran>): <ID> <ringkasan>`, contoh `feat(backend): B-07 login, logout, dan sesi`. Jenis: `feat`, `fix`, `test`, `docs`, `chore`.
- Sebelum push: `git fetch origin && git merge origin/dev`, lalu jalankan perintah verifikasi. Jangan `rebase`, jangan `--force`.
- Sebelum mengubah file apa pun, jalankan `git branch --show-current`. Kalau hasilnya `main`, `dev`, atau `master`, berhenti dan minta pengguna membuat branch tugas dari `dev`.
- Pull request selalu ber-base `dev`, tidak pernah `main`.
- Patuhi bagian 9 di `docs/03-git-workflow.md` (aturan khusus AI agent).

## 5. Port dan alamat lokal

| Layanan | Alamat |
| --- | --- |
| Frontend (Astro dev) | `http://localhost:4321` (meneruskan `/api` dan `/health` ke backend) |
| Backend Go | `http://localhost:3000` |
| MySQL | `127.0.0.1:3306`, database `sitangkap`, user `sitangkap_app` / `sitangkap_dev` (hanya lokal) |

## 6. Definisi selesai (semua peran)

- Perilaku sesuai kontrak dan dokumen peran.
- Perintah verifikasi lulus, output ditunjukkan.
- Tidak ada file di luar folder peran yang berubah.
- Tidak ada dependensi baru dan tidak ada rahasia.
- Checklist di akhir dokumen peran sudah dicentang untuk tahap itu.
