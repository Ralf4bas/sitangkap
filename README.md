# SiTangkap

Aplikasi web pencatatan transaksi lelang ikan di TPI (Tempat Pelelangan Ikan). Petugas TPI mencatat hasil lelang; dinas perikanan kabupaten/kota langsung melihat daftar dan rekapnya tanpa menunggu laporan fisik.

Proyek kelompok mata kuliah DevOps: development dulu, lalu CI/CD dengan Jenkins dan webhook GitHub.

| Peran | Nama | Panduan |
| --- | --- | --- |
| Project Manager | tmthy | [docs/roles/pm-tmthy.md](docs/roles/pm-tmthy.md) |
| Frontend | Magen | [docs/roles/frontend-magen.md](docs/roles/frontend-magen.md) |
| Backend | Rafli | [docs/roles/backend-rafli.md](docs/roles/backend-rafli.md) |
| Database | abas | [docs/roles/database-abas.md](docs/roles/database-abas.md) |

## Dokumen

| File | Isi |
| --- | --- |
| [AGENTS.md](AGENTS.md) | Aturan untuk AI coding agent (juga berguna untuk manusia) |
| [docs/00-overview.md](docs/00-overview.md) | Masalah, ruang lingkup, requirement, use case, RTM, arsitektur |
| [docs/01-api-contract.md](docs/01-api-contract.md) | **Kontrak API dan konfigurasi: sumber kebenaran** |
| [docs/02-cicd-jenkins.md](docs/02-cicd-jenkins.md) | Jenkins + webhook GitHub di VPS (fase 2) |
| [docs/03-git-workflow.md](docs/03-git-workflow.md) | **Branch, commit, pull request. Kerja di branch tugas → PR ke `dev` → rilis ke `main`. Dilarang push langsung ke `dev`/`main`** |
| Figma | <https://www.figma.com/design/DhWEhuRCLzcyPLhZuYiavO> |

## Sebelum commit pertama

Baca [docs/03-git-workflow.md](docs/03-git-workflow.md), lalu aktifkan pengaman lokal:

```bash
git config core.hooksPath .githooks
```

## Menjalankan di laptop

Prasyarat: Git, Go 1.26+, Node.js 22.23.4+ (syarat Astro 7; versi 22 yang lebih lama gagal saat `npm run build`), MySQL 8.0.16+ atau MariaDB 10.6+.

```bash
# 1. Database
mysql -u root -p < db/schema.sql
mysql -u root -p < db/app_user.sql
mysql -u root -p sitangkap < db/seed.sql

# 2. Backend (terminal 1)
cd backend
cp .env.example .env
go run .                      # http://localhost:3000/health

# 3. Frontend (terminal 2)
cd frontend
cp .env.example .env
npm install
npm run dev                   # buka http://localhost:4321
```

Akun demo: `petugas01` / `petugas123` dan `dinas01` / `dinas123`.

## Struktur repo

```text
backend/    Go (Rafli)
frontend/   Astro (Magen)
db/         SQL (abas)
docs/       dokumen proyek (tmthy)
deploy/     file server: systemd, sudoers (tmthy)
Jenkinsfile pipeline CI/CD (tmthy)
```
