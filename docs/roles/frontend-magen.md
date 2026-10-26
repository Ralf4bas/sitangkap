# Panduan Peran: Frontend (Magen)

> **Untuk Magen dan AI agent yang membantu Magen.**
> Baca dulu, berurutan: [`../../AGENTS.md`](../../AGENTS.md), [`../00-overview.md`](../00-overview.md), [`../01-api-contract.md`](../01-api-contract.md) (**wajib seluruhnya**).
> Desain: Figma <https://www.figma.com/design/DhWEhuRCLzcyPLhZuYiavO> (minta akses ke PM).

## Ringkasan peran

| Hal | Isi |
| --- | --- |
| Tanggung jawab | Tiga halaman Astro (login, petugas, dinas), panggilan API, format angka dan tanggal, kondisi kosong dan error |
| Folder yang kamu miliki | `frontend/` |
| Jangan ubah | `backend/`, `db/`, `docs/` (minta lewat issue) |
| Kamu menerima dari | **PM:** akses Figma. **Backend (Rafli):** API sesuai kontrak di `localhost:3000`. **Database:** cara menyiapkan database lokal (untuk uji integrasi) |
| Kamu memberi ke | **Backend/PM:** folder `frontend/dist` hasil `npm run build` (dilayani backend saat production). **PM:** hasil cek manual F-01 s.d. F-13 |
| Bisa mulai kapan | **Segera.** Pakai mode mock (`PUBLIC_MOCK=1`) sampai backend siap |

## Aturan teknis yang tidak boleh dilanggar

1. Astro dengan output statis (default). **Tanpa** React, Vue, Svelte, Tailwind, jQuery, atau library UI lain. JavaScript biasa (TypeScript diperbolehkan) di `<script>` Astro.
2. Tanpa CDN dan tanpa font eksternal (NFR-006). Font memakai stack sistem (lihat token).
3. Data dari API **selalu** dimasukkan ke halaman dengan `textContent` atau `createElement`. **Dilarang** `innerHTML` untuk data (mencegah XSS).
4. Frontend memanggil path relatif `/api/...`. Jangan menulis `http://localhost:3000` di kode.
5. Jangan menyimpan token atau data user di `localStorage`. Login memakai cookie dari backend; frontend cukup memanggil `/api/me`.
6. Teks yang tampil di layar memakai Bahasa Indonesia, sesuai Figma.

## Struktur folder

```text
frontend/
  package.json
  package-lock.json        WAJIB di-commit (Jenkins memakai npm ci)
  astro.config.mjs         proxy /api dan /health ke backend
  .env.example             PUBLIC_MOCK=0
  src/
    styles/global.css      token warna dan komponen dasar dari Figma
    layouts/Layout.astro   <html>, <head>, import global.css
    components/
      Topbar.astro         nama aplikasi, nama TPI, pengguna, tombol Keluar
      StatCard.astro       kartu ringkasan (dinas)
    pages/
      index.astro          login                     (Figma 01, 02)
      petugas.astro        daftar + modal + konfirmasi (Figma 10–14)
      dinas.astro          filter + rekap + daftar   (Figma 20, 21)
    scripts/
      api.ts               semua fetch ke backend + penanganan error
      mock.ts              data dan perilaku palsu saat PUBLIC_MOCK=1
      format.ts            rupiah, tanggal, berat, parsing input angka
      hitung.ts            pratinjau nilai dan retribusi di form
      guard.ts             penjaga halaman: cek /api/me dan peran
      login.ts, petugas.ts, dinas.ts   logika per halaman
```

## FE-01 · Membuat proyek

```bash
npm create astro@latest frontend
# pilih template paling minimal/kosong, TypeScript "strict", install dependencies: ya, git init: TIDAK (repo sudah ada)
cd frontend
npm run dev        # http://localhost:4321
```

`astro.config.mjs`:

```js
import { defineConfig } from 'astro/config';

export default defineConfig({
  server: { port: 4321 },
  vite: {
    server: {
      proxy: {
        '/api': 'http://localhost:3000',
        '/health': 'http://localhost:3000',
      },
    },
  },
});
```

`create-astro` menolak folder yang tidak kosong, jadi folder `frontend/` **belum** ada di repo saat kamu mulai. Setelah proyek dibuat, tambahkan `.env.example` (di-commit) dan `.env` (lokal, tidak di-commit):

```dotenv
PUBLIC_MOCK=0
```

Selesai bila: `npm run dev` menampilkan halaman kosong di port 4321 dan `npm run build` membuat folder `dist/`.

## FE-02 · Token desain dari Figma (`src/styles/global.css`)

Nilai ini diambil dari file Figma. Pakai variabel, jangan menulis warna langsung di komponen.

```css
:root {
  --c-brand: #0b6e8a;        /* tombol utama, topbar, judul login */
  --c-ink: #12263a;          /* teks utama */
  --c-muted: #5b6b7a;        /* teks sekunder, header tabel */
  --c-line: #dbe3ea;         /* garis, border input dan kartu */
  --c-bg: #f4f7f9;           /* latar halaman, header tabel */
  --c-surface: #ffffff;      /* kartu, tabel, modal */
  --c-danger: #b3261e;       /* tombol hapus, teks error */
  --c-danger-bg: #fdedec;    /* kotak pesan error */
  --radius-sm: 8px;          /* tombol, input */
  --radius-md: 12px;         /* kartu, tabel */
  --radius-lg: 16px;         /* kartu login, modal */
  --font: Inter, system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
}
body { margin: 0; font: 14px/1.5 var(--font); color: var(--c-ink); background: var(--c-bg); }
```

Ukuran dari Figma:

| Elemen | Nilai |
| --- | --- |
| Kanvas desain | 1280 × 800 |
| Topbar | tinggi 64, padding 16/32, latar `--c-brand`, teks putih |
| Konten | padding 32, jarak antarblok 20 |
| Judul halaman | 24 px tebal; subjudul 14 px `--c-muted` |
| Tombol | padding 10/16, radius 8, teks 14 semibold. Varian: utama (`--c-brand`), sekunder (putih, border `--c-brand`), bahaya (`--c-danger`) |
| Input | label 13 medium; kotak padding 10/12, border `--c-line`, radius 8 |
| Kartu ringkasan | padding 18/20, radius 12, judul 13 `--c-muted`, angka 24 tebal |
| Tabel | header 12 semibold huruf besar `--c-muted` di atas `--c-bg`; baris 14, padding 12/16, garis bawah `--c-line`; kolom angka rata kanan |
| Modal | lebar 520 (form) atau 440 (konfirmasi), padding 28, radius 16, latar gelap 45% |

## FE-03 · Format dan input (`format.ts`, `hitung.ts`)

| Fungsi | Masukan → Keluaran |
| --- | --- |
| `rupiah(n)` | `3374000` → `Rp 3.374.000` (`Intl.NumberFormat('id-ID')`) |
| `berat(n)` | `120.5` → `120,5`; `45` → `45`; `30.25` → `30,25` |
| `tanggalID(s)` | `"2026-10-22"` → `22/10/2026` (potong string, **jangan** lewat `new Date()`) |
| `hariIniISO()` | Tanggal lokal browser `YYYY-MM-DD` dari `getFullYear()/getMonth()/getDate()`. **Jangan** `toISOString()` (itu UTC dan bisa mundur sehari) |
| `parseBerat(s)` | `"120,5"` atau `"120.5"` → `120.5`; tidak valid → `null` |
| `parseHarga(s)` | `"28.000"` atau `"28000"` → `28000`; tidak valid → `null` |

Pratinjau di form memakai rumus yang **sama** dengan backend (kontrak bagian 3):

```ts
export function hitungNilai(beratKg: number, hargaPerKg: number): number {
  const centi = Math.round(beratKg * 100);
  return Math.floor((centi * hargaPerKg + 50) / 100);
}
export function hitungRetribusi(nilai: number, tarif: number): number {
  return Math.round(nilai * tarif);
}
```

Tarif pratinjau: `meta.tarif_retribusi` dari `listTransaksi()` (jangan ditulis tetap). Isian berat atau harga yang kosong/tidak terbaca dikirim sebagai `null`; jangan memblokir form di browser, tampilkan pesan dari `fields` respons 422. Setelah disimpan, **selalu** tampilkan `nilai` dan `retribusi` dari respons API, bukan hasil hitungan sendiri. Di layar dinas, tarif dibaca dari `tarif_retribusi` respons rekap.

Selesai bila: semua baris tabel contoh di kontrak bagian 3 menghasilkan angka yang sama.

## FE-04 · Lapisan API (`api.ts`) dan mode mock (`mock.ts`)

```ts
export type Role = 'petugas' | 'dinas';
export interface User { id: number; username: string; nama: string; role: Role }
export interface Transaksi {
  id: number; nama_nelayan: string; jenis_ikan: string; berat_kg: number; harga_per_kg: number;
  tanggal: string; nilai: number; retribusi: number; dibuat_oleh: string; dibuat_pada: string;
}
export interface TransaksiInput {
  // null = isian kosong atau tidak bisa dibaca sebagai angka; backend akan menjawab 422 dengan pesan field
  nama_nelayan: string; jenis_ikan: string; berat_kg: number | null; harga_per_kg: number | null; tanggal: string;
}
export interface Rekap {
  dari: string; sampai: string; jumlah_transaksi: number; total_berat_kg: number;
  total_nilai: number; total_retribusi: number; tarif_retribusi: number;
}
export class ApiError extends Error {
  constructor(public status: number, public code: string, message: string,
              public fields?: Record<string, string>) { super(message); }
}

// Fungsi yang wajib ada (tanda tangan tetap; dipakai semua halaman):
// login(username, password): Promise<User>
// logout(): Promise<void>
// me(): Promise<User>
// listTransaksi(range?: { dari: string; sampai: string }): Promise<{ data: Transaksi[]; meta: { jumlah: number; dari: string | null; sampai: string | null; tarif_retribusi: number } }>
// createTransaksi(input: TransaksiInput): Promise<Transaksi>
// updateTransaksi(id: number, input: TransaksiInput): Promise<Transaksi>
// deleteTransaksi(id: number): Promise<void>
// rekap(range?: { dari: string; sampai: string }): Promise<Rekap>
```

Aturan `api.ts`:
- Respons bukan 2xx → baca `error.code`, `error.message`, `error.fields`, lalu `throw new ApiError(...)`.
- Body tidak bisa dibaca sebagai JSON → `ApiError(status, 'INTERNAL_ERROR', 'Terjadi kesalahan pada server.')`.
- Gagal jaringan (backend mati) → `ApiError(0, 'NETWORK', 'Tidak bisa terhubung ke server.')`.
- Respons `204` → selesai tanpa membaca body (logout, hapus). Respons 2xx lain → kembalikan `body.data` (dan `body.meta` untuk daftar).
- `api.ts` **tidak** melakukan redirect apa pun. Penanganan 401 dilakukan oleh `guard.ts` (FE-05) dan oleh halaman yang memanggil. Ini mencegah halaman login me-reload dirinya sendiri terus-menerus.

Mode mock (`PUBLIC_MOCK=1`):
- `api.ts` memeriksa `import.meta.env.PUBLIC_MOCK === '1'` lalu memanggil `mock.ts`.
- `mock.ts` menyimpan data di memori: akun demo (`petugas01`/`petugas123`, `dinas01`/`dinas123`) dan **8 transaksi dari `db/seed.sql`**, dengan `dibuat_oleh: "petugas01"`, `dibuat_pada` = `<tanggal>T09:00:00+08:00`, `tarif_retribusi: 0.03`. Bentuk respons, kode error, dan pesan persis seperti kontrak, termasuk 403 dan 422 dengan pesan validasi.
- "Sesi" mock disimpan di `sessionStorage` (khusus mock).
- Mode mock tidak boleh ikut ter-build dengan nilai `1`: Jenkins membangun dengan `PUBLIC_MOCK=0`.

## FE-05 · Penjaga halaman (`guard.ts`)

```ts
// Dipanggil paling awal di petugas.ts dan dinas.ts.
export async function requireRole(role: Role): Promise<User>
```

1. Panggil `me()`. Gagal 401 → `location.href = '/'`.
2. `user.role !== role` → arahkan ke halaman perannya sendiri (`/petugas` atau `/dinas`).
3. Kembalikan `user` untuk ditampilkan di topbar.

Halaman login (`index.astro`): saat dibuka, panggil `me()`. Kalau sudah login, langsung arahkan ke halaman sesuai peran.

Selama `requireRole` berjalan, sembunyikan isi halaman (misalnya atribut `hidden` pada `<main>`) supaya data tidak berkedip sebelum pemeriksaan selesai.

## FE-06 · Halaman login (`index.astro`) — Figma 01, 02

| Elemen | Perilaku |
| --- | --- |
| Input Username, Password | `<label>` + `<input>`; password `type="password"`; `autocomplete="username"` / `"current-password"` |
| Tombol Masuk | Lebar penuh; disabled + teks "Memproses..." selama request |
| Pesan error (Figma 02) | Kotak `--c-danger-bg` di atas input, `role="alert"`, isi `error.message` dari API |
| Sukses | `role === 'petugas'` → `/petugas`; `role === 'dinas'` → `/dinas` |
| Catatan bawah | "Akun dibuat oleh pengelola sistem. Hubungi dinas bila lupa password." |

Kirim form dengan tombol Enter (pakai `<form>` dan event `submit`).

## FE-07 · Halaman petugas (`petugas.astro`) — Figma 10–14

| Bagian | Isi dan perilaku |
| --- | --- |
| Topbar | "SiTangkap", "TPI Contoh · Dinas Perikanan Kabupaten/Kota", `"{username} · Petugas TPI"` (pakai `username`, bukan `nama`), tombol Keluar (`logout()` lalu ke `/`) |
| Judul | "Hasil Lelang" + "Transaksi lelang yang sudah selesai di TPI Contoh"; tombol utama "+ Tambah transaksi" di kanan |
| Tabel | Kolom: Tanggal, Nelayan, Jenis ikan, Berat (kg), Harga/kg, Nilai, Retribusi, Aksi (Ubah, Hapus). Data dari `listTransaksi()` tanpa filter |
| Footer | "Menampilkan {n} transaksi" |
| Kosong (Figma 14) | Di dalam tabel: "Belum ada transaksi" + "Catat hasil lelang pertama dengan tombol "+ Tambah transaksi"." |
| Modal tambah (Figma 11) | Pakai elemen `<dialog>`. Input: Nama nelayan, Jenis ikan, Berat (kg) dan Harga lelang per kg (Rp) berdampingan, Tanggal lelang (`type="date"`, default hari ini, `max` = hari ini). Kotak pratinjau "Nilai transaksi" dan "Retribusi (tarif {persen}%)" diperbarui saat mengetik; `persen` dihitung dari `meta.tarif_retribusi` seperti di FE-08. Tombol Batal dan Simpan |
| Modal ubah (Figma 12) | Sama, judul "Ubah transaksi", terisi data baris, tombol "Simpan perubahan" → `updateTransaksi` |
| Konfirmasi hapus (Figma 13) | `<dialog>` kecil: "Hapus transaksi ini?" + "{nama} · {ikan} · {berat} kg · {tanggal}. Data yang dihapus tidak bisa dikembalikan." Tombol Batal dan Hapus (merah) |
| Error validasi | Respons 422 → tampilkan `fields[nama_field]` di bawah input yang sesuai; modal tetap terbuka |
| Error lain | Pesan `error.message` di atas tabel atau di dalam modal, `role="alert"` |
| Setelah simpan/hapus | Tutup modal, muat ulang daftar |

## FE-08 · Halaman dinas (`dinas.astro`) — Figma 20, 21

| Bagian | Isi dan perilaku |
| --- | --- |
| Topbar | Sama, pengguna `"{username} · Dinas Kabupaten/Kota"` |
| Judul | "Rekap Hasil Lelang" + "Semua transaksi lelang dari TPI Contoh" |
| Filter | Input "Dari tanggal" dan "Sampai tanggal" (`type="date"`), tombol "Terapkan"; tombol sekunder "Hari ini", "Bulan ini", "Tahun ini" mengisi kedua input lalu langsung menerapkan |
| Default saat dibuka | Tanggal 1 bulan ini s.d. hari ini |
| Validasi filter | Salah satu tanggal kosong → "Isi kedua tanggal." `dari` > `sampai` → "Tanggal awal tidak boleh setelah tanggal akhir." Keduanya tanpa memanggil API |
| Kartu ringkasan | Total berat (`370,75 kg`), Nilai transaksi, Retribusi, Jumlah transaksi, dari `rekap(range)` |
| Keterangan | "Periode {dari} sampai {sampai}. Retribusi dihitung dengan tarif {persen}%." dengan `persen = (Math.round(tarif * 10000) / 100).toLocaleString('id-ID')` (0.03 → `3`, 0.0285 → `2,85`) |
| Tabel | Kolom sama dengan petugas **tanpa** Aksi; data dari `listTransaksi(range)` |
| Kosong (Figma 21) | Kartu bernilai 0; di tabel: "Tidak ada transaksi pada rentang tanggal ini" + "Coba ubah tanggal Dari dan Sampai, atau pilih Bulan ini." |

Panggil `rekap(range)` dan `listTransaksi(range)` bersamaan (`Promise.all`) dengan rentang yang sama.

## FE-09 · Responsif dan aksesibilitas

- Lebar 360 px: konten tetap terbaca; filter dan kartu bertumpuk vertikal; tabel dibungkus `<div style="overflow-x:auto">` sehingga hanya tabel yang bisa digeser (NFR-007).
- Setiap input punya `<label for>`; pesan error memakai `role="alert"`; `<dialog>` dibuka dengan `showModal()` supaya fokus tertahan di dalamnya dan Esc menutupnya.
- Kontras teks mengikuti token (warna sudah aman di atas latar putih/abu muda).

## FE-10 · Cek manual sebelum PR

Jalankan dengan backend sungguhan (`PUBLIC_MOCK=0`) dan database yang baru di-seed.

| ID | Langkah | Harapan |
| --- | --- | --- |
| F-01 | Login `petugas01`/`petugas123`, lalu `dinas01`/`dinas123` | Masing-masing masuk ke `/petugas` dan `/dinas` |
| F-02 | Login dengan password salah | Kotak merah "Username atau password salah." |
| F-03 | Klik Keluar, lalu buka `/petugas` langsung | Kembali ke login |
| F-04 | Buka `/petugas` | 8 baris, urut tanggal terbaru; baris pertama 22/10/2026 |
| F-05 | Tambah: Pak Roni, Cakalang, `120,5`, `28.000`, hari ini | Pratinjau `Rp 3.374.000` / `Rp 101.220`; setelah simpan muncul di tabel dengan angka yang sama |
| F-06 | Ubah baris itu ke berat 100 | Nilai `Rp 2.800.000`, retribusi `Rp 84.000` |
| F-07 | Hapus baris itu dengan konfirmasi | Hilang dari tabel; Batal tidak menghapus |
| F-08 | Dinas, rentang 01/10/2026–22/10/2026 (data seed) | `370,75 kg`, `Rp 11.577.000`, `Rp 347.310`, `5` |
| F-09 | Dinas, rentang 01/01/2025–31/01/2025 | Semua 0, tampilan kosong (Figma 21) |
| F-10 | Dinas, klik Hari ini / Bulan ini / Tahun ini | Input tanggal berubah sesuai, data ikut berubah |
| F-11 | Tambah dengan berat `-1` dan nama kosong | Pesan merah di bawah kedua input, modal tetap terbuka |
| F-12 | Tab Network di DevTools, halaman login | Total terkirim < 100 KB |
| F-13 | Lebar layar 360 px | Tidak ada scroll horizontal di halaman (tabel boleh) |

## FE-11 · API dan mock untuk chat (FR-012)

> **FE-11 dan FE-12 dikerjakan tmthy, Magen me-review dan meng-approve PR-nya.** Daftar file yang boleh diubah ada di `AGENTS.md` bagian 1.

Kerjakan setelah PR kontrak AI-01 digabung. Bentuk data ada di kontrak bagian "Jawaban chat" dan `POST /api/chat`.

```ts
export interface JawabanChat {
  jawaban: string; dikenali: boolean; niat: string | null; dari: string | null; sampai: string | null;
}
// chat(pesan: string): Promise<JawabanChat>
```

- `chat` di `api.ts` mengikuti aturan `api.ts` yang sudah ada (ApiError, gagal jaringan, tanpa redirect).
- Mock (`PUBLIC_MOCK=1`) **tidak** meniru aturan chat, karena aturannya tinggal di backend. Mock cukup:
  - Membalas 422 dengan `fields.pesan` persis kontrak untuk pesan kosong atau lebih dari 200 karakter.
  - Membalas 401 bila belum login.
  - Selain itu membalas templat "Tidak dikenali" (`dikenali: false`, `niat`, `dari`, `sampai` bernilai `null`).

## FE-12 · Panel "Tanya SiTangkap" (Figma 30)

Komponen `components/ChatPanel.astro` dengan logika di `scripts/chat.ts`, dipasang di `petugas.astro` dan `dinas.astro`. Kalau Figma 30 belum ada, ikuti tabel ini dan token desain yang sudah ada.

| Bagian | Isi dan perilaku |
| --- | --- |
| Pembuka | Tombol sekunder "Tanya SiTangkap" di bawah judul halaman. Klik → panel terbuka di bawahnya; klik lagi → tertutup |
| Petunjuk awal | Teks abu-abu: `Tanyakan rekap, jumlah transaksi, berat, nilai, atau retribusi. Contoh: "rekap bulan ini".` |
| Daftar pesan | Pertanyaan pengguna rata kanan, jawaban rata kiri. Wadah memakai `aria-live="polite"`. Isi dimasukkan dengan `textContent`, **tidak pernah** `innerHTML` |
| Input | `<label for>` "Pertanyaan", `<input type="text" maxlength="200">`, tombol "Kirim". Enter juga mengirim |
| Saat menunggu | Tombol "Kirim" nonaktif, muncul teks "Sedang mencari jawaban…" di daftar |
| 422 | Pesan `fields.pesan` merah di bawah input (`role="alert"`). Pertanyaan tidak masuk daftar |
| 401 | Arahkan ke halaman login, sama dengan penanganan 401 di halaman itu |
| Gagal jaringan atau 500 | `message` dari `ApiError` tampil sebagai jawaban berwarna merah |
| Riwayat | Hanya di memori halaman. Hilang saat halaman dimuat ulang. **Jangan** simpan di `localStorage` atau `sessionStorage` |
| 360 px | Panel selebar konten, input dan tombol bertumpuk bila tidak muat |

Cek manual tambahan (backend sungguhan, database baru di-seed):

| ID | Langkah | Harapan |
| --- | --- | --- |
| F-14 | Dinas, tanya `rekap dari 01/10/2026 sampai 22/10/2026` | Jawaban persis baris pertama tabel contoh di kontrak; angkanya sama dengan kartu di F-08 |
| F-15 | Petugas, tanya `transaksi terakhir`, lalu `cuaca besok bagaimana` | Jawaban persis tabel contoh di kontrak (Ibu Merry, Julung-julung …), lalu jawaban "Maaf, saya belum mengerti …" |
| F-16 | Kirim pertanyaan kosong; lalu ulangi F-14 di lebar 360 px | "Pertanyaan wajib diisi." di bawah input; panel tetap terbaca tanpa scroll horizontal |

## Koordinasi

| Kapan | Dengan siapa | Apa |
| --- | --- | --- |
| Hari 1 | — | Bangun semua halaman dalam mode mock |
| Setelah backend CP-2 | Backend (Rafli) | Matikan mock, uji login dan daftar dengan backend sungguhan |
| Menemukan respons berbeda dari kontrak | Backend + PM | Issue berlabel `kontrak` dengan contoh request/respons; jangan menyesuaikan frontend ke respons yang salah |
| Sebelum fase 2 | PM | Pastikan `npm ci && npm run build` lulus dari repo yang baru di-clone |

## Checklist selesai

- [ ] Tiga halaman sesuai Figma 01–02, 10–14, 20–21
- [ ] Mode mock dan mode backend sungguhan sama-sama jalan
- [ ] Tidak ada `innerHTML` untuk data, tidak ada `localStorage`
- [ ] Tidak ada library UI atau CDN; `package-lock.json` ter-commit
- [ ] F-01 s.d. F-13 lulus (dan F-14 s.d. F-16 setelah FE-11, FE-12)
- [ ] `npm run build` lulus

## Prompt siap pakai untuk AI agent

Kerjakan per tahap (satu sesi agent per tahap FE-xx).

```text
Kamu membantu peran FRONTEND di proyek SiTangkap (Astro statis, JavaScript/TypeScript biasa, tanpa framework UI).
Baca dulu, berurutan: AGENTS.md, docs/00-overview.md, docs/01-api-contract.md, docs/roles/frontend-magen.md.
Tugas sesi ini: kerjakan tahap <FE-xx> di docs/roles/frontend-magen.md.
Aturan:
- Hanya ubah file di folder frontend/.
- Ikuti docs/01-api-contract.md persis untuk nama field, bentuk respons, dan pesan error.
- Jangan pakai React/Vue/Svelte/Tailwind/CDN/font eksternal. Jangan pakai innerHTML untuk data. Jangan pakai localStorage.
- Panggil API dengan path relatif /api/... lewat fungsi di src/scripts/api.ts.
- Teks di layar Bahasa Indonesia, sesuai tabel di dokumen peran.
- Setelah selesai, jalankan `npm run build` dan tunjukkan hasilnya, lalu sebutkan cek manual F-xx yang relevan.
Kalau kontrak atau dokumen tidak menjawab sesuatu, berhenti dan tanyakan; jangan menebak.
```
