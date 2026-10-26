# 01 · Kontrak API dan Konfigurasi Bersama

> **Sumber kebenaran tertinggi proyek ini.** Frontend, backend, dan database **wajib** mengikuti dokumen ini persis: nama field, tipe, kode status, dan format error.
> AI agent: **jangan menambah, mengganti nama, atau menghapus endpoint/field** yang tidak tertulis di sini. Kalau perlu perubahan, berhenti dan minta PM (tmthy) memperbarui dokumen ini lebih dulu.

Pembaca utama: Backend (implementasi), Frontend (pemakai), Database (bentuk data), PM (peninjau).

## 1. Konvensi umum

| Hal | Aturan |
| --- | --- |
| Base URL development | Frontend memanggil path relatif (`/api/...`). Astro dev server (port 4321) meneruskannya ke backend `http://localhost:3000`. |
| Format | JSON, UTF-8. Header `Content-Type: application/json` untuk request dan respons yang punya body. |
| Nama field | `snake_case`, Bahasa Indonesia, sama persis dengan kolom database bila ada padanannya. **Pengecualian disengaja:** `dibuat_oleh` di API berisi `username` (string), sedangkan kolom `transaksi.dibuat_oleh` di database berisi `users.id` (integer). Backend menerjemahkannya lewat JOIN. Jangan "menyamakan" salah satunya. |
| Tanggal | String `YYYY-MM-DD` (contoh `2026-10-22`). Tanggal-waktu: RFC 3339 dengan offset (contoh `2026-10-22T09:15:00+08:00`). |
| Uang | Integer rupiah, tanpa desimal (contoh `3374000`). |
| Berat | Number JSON dengan paling banyak 2 desimal (contoh `120.5`). Bukan string. |
| Zona waktu "hari ini" | `APP_TIMEZONE`, default `Asia/Makassar` (WITA). |
| Autentikasi | Cookie sesi `sitangkap_session`. Browser mengirimnya otomatis; frontend tidak menyimpan token sendiri. |
| CORS | Tidak ada. Semua akses dari origin yang sama (lewat proxy saat development, satu server saat production). |

### Bentuk respons sukses

```json
{ "data": <objek atau array> }
```

Respons daftar menambahkan `meta`:

```json
{ "data": [ ... ], "meta": { "jumlah": 5, "dari": "2026-10-01", "sampai": "2026-10-22", "tarif_retribusi": 0.03 } }
```

`204 No Content` tidak punya body sama sekali.

### Bentuk respons error (semua endpoint)

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Data yang dikirim belum benar.",
    "fields": { "berat_kg": "Berat harus lebih dari 0." }
  }
}
```

`fields` hanya ada pada `VALIDATION_ERROR`. `message` selalu Bahasa Indonesia dan aman ditampilkan ke pengguna.

| HTTP | `code` | `message` default | Kapan |
| --- | --- | --- | --- |
| 400 | `BAD_REQUEST` | `Permintaan tidak valid.` | JSON rusak, body > 10 KB, parameter query salah format, `dari` > `sampai`, id bukan angka |
| 401 | `UNAUTHORIZED` | `Silakan login terlebih dahulu.` | Tidak ada cookie sesi, sesi tidak dikenal atau kedaluwarsa |
| 401 | `INVALID_CREDENTIALS` | `Username atau password salah.` | Login gagal (username tidak ada, password salah, atau akun nonaktif: pesannya **sama**) |
| 403 | `FORBIDDEN` | `Anda tidak memiliki akses untuk tindakan ini.` | Peran tidak diizinkan (misalnya `dinas` memanggil `POST /api/transaksi`) |
| 404 | `NOT_FOUND` | `Data tidak ditemukan.` | Transaksi dengan id itu tidak ada, atau path `/api/...` tidak dikenal |
| 422 | `VALIDATION_ERROR` | `Data yang dikirim belum benar.` | Body valid JSON tapi melanggar aturan validasi |
| 500 | `INTERNAL_ERROR` | `Terjadi kesalahan pada server.` | Error tak terduga. Detail hanya ke log server |
| 503 | — | — | Hanya `/health` saat database tidak terjangkau |

## 2. Objek data

### User

```json
{ "id": 1, "username": "petugas01", "nama": "Petugas Contoh", "role": "petugas" }
```

| Field | Tipe | Keterangan |
| --- | --- | --- |
| `id` | integer | |
| `username` | string | |
| `nama` | string | Nama lengkap pengguna. **Topbar menampilkan `username`, bukan `nama`** (sesuai Figma) |
| `role` | string | Hanya `"petugas"` atau `"dinas"` |

`password_hash` **tidak pernah** dikirim ke klien.

### Transaksi

```json
{
  "id": 1,
  "nama_nelayan": "Pak Roni",
  "jenis_ikan": "Cakalang",
  "berat_kg": 120.5,
  "harga_per_kg": 28000,
  "tanggal": "2026-10-22",
  "nilai": 3374000,
  "retribusi": 101220,
  "dibuat_oleh": "petugas01",
  "dibuat_pada": "2026-10-22T09:15:00+08:00"
}
```

| Field | Tipe | Sumber | Keterangan |
| --- | --- | --- | --- |
| `id` | integer | database | |
| `nama_nelayan` | string | input | 1–100 karakter setelah trim |
| `jenis_ikan` | string | input | 1–60 karakter setelah trim |
| `berat_kg` | number | input | > 0, ≤ 100000, maks 2 desimal |
| `harga_per_kg` | integer | input | > 0, ≤ 100000000 |
| `tanggal` | string `YYYY-MM-DD` | input | Tanggal nyata, tidak setelah hari ini |
| `nilai` | integer | **dihitung backend** | `round(berat_kg × harga_per_kg)` |
| `retribusi` | integer | **dihitung backend** | `round(nilai × RETRIBUSI_TARIF)` |
| `dibuat_oleh` | string | database (join `users.username`) | Username pencatat. Di database kolom ini berisi id user (lihat pengecualian di bagian 1) |
| `dibuat_pada` | string RFC 3339 | database | Waktu dicatat |

### Rekap

```json
{
  "dari": "2026-10-01",
  "sampai": "2026-10-22",
  "jumlah_transaksi": 5,
  "total_berat_kg": 370.75,
  "total_nilai": 11577000,
  "total_retribusi": 347310,
  "tarif_retribusi": 0.03
}
```

### Jawaban chat

```json
{
  "jawaban": "Ada 3 transaksi pada 22/10/2026.",
  "dikenali": true,
  "niat": "jumlah",
  "dari": "2026-10-22",
  "sampai": "2026-10-22"
}
```

| Field | Tipe | Keterangan |
| --- | --- | --- |
| `jawaban` | string | Kalimat Bahasa Indonesia yang siap ditampilkan apa adanya |
| `dikenali` | boolean | `false` bila pertanyaan tidak cocok dengan aturan mana pun atau tanggalnya tidak valid |
| `niat` | string atau `null` | Kode niat dari tabel di `POST /api/chat`. `null` bila `dikenali = false` |
| `dari`, `sampai` | string `YYYY-MM-DD` atau `null` | Rentang yang dipakai untuk menjawab. `null` untuk niat yang tidak memakai rentang (`sapaan`, `bantuan`, `transaksi_terakhir`) dan bila `dikenali = false` |

## 3. Aturan perhitungan (BR-05)

Semua perhitungan dilakukan **di backend, di satu tempat** (fungsi `hitungNilai` dan `hitungRetribusi`). Database hanya menyimpan input. Frontend hanya menampilkan pratinjau di form (rumus sama, tarif dari `meta.tarif_retribusi` di `GET /api/transaksi`) dan selalu menampilkan angka dari API setelah disimpan.

```text
nilai      = round(berat_kg * harga_per_kg)          // rupiah, integer
retribusi  = round(nilai * RETRIBUSI_TARIF)          // rupiah, integer
round      = pembulatan ke bilangan bulat terdekat, .5 menjauhi nol (Go: math.Round)
```

Rekap dihitung dengan **menjumlahkan nilai per transaksi yang sudah dibulatkan**, supaya total di rekap selalu sama dengan jumlah kolom di tabel:

```text
total_nilai      = Σ nilai per transaksi
total_retribusi  = Σ retribusi per transaksi
total_berat_kg   = Σ berat_kg, dibulatkan ke 2 desimal
```

Contoh uji (tarif 0.03), dipakai di tes backend dan cek frontend:

| berat_kg | harga_per_kg | nilai | retribusi |
| --- | --- | --- | --- |
| 120.5 | 28000 | 3374000 | 101220 |
| 45 | 65000 | 2925000 | 87750 |
| 30.25 | 22000 | 665500 | 19965 |
| 80 | 25000 | 2000000 | 60000 |
| 95 | 27500 | 2612500 | 78375 |
| **Total 370.75** | | **11577000** | **347310** |

## 4. Endpoint

Ringkasan:

| Method | Path | Login? | Peran | Ref |
| --- | --- | --- | --- | --- |
| GET | `/health` | Tidak | — | FR-011 |
| POST | `/api/login` | Tidak | — | FR-001 |
| POST | `/api/logout` | Tidak wajib | — | FR-002 |
| GET | `/api/me` | Ya | petugas, dinas | FR-001 |
| GET | `/api/transaksi` | Ya | petugas, dinas | FR-003, FR-009 |
| POST | `/api/transaksi` | Ya | **petugas** | FR-004 |
| PUT | `/api/transaksi/{id}` | Ya | **petugas** | FR-005 |
| DELETE | `/api/transaksi/{id}` | Ya | **petugas** | FR-006 |
| GET | `/api/rekap` | Ya | **dinas** | FR-008 |
| POST | `/api/chat` | Ya | petugas, dinas | FR-012 |

Urutan pemeriksaan di setiap endpoint yang butuh login: **401 dulu (sesi), lalu 403 (peran), lalu 400 (format: JSON rusak, id bukan angka, query salah), lalu 422 (validasi isi), lalu 404 (data dengan id itu tidak ada).**

---

### GET /health

Untuk Jenkins dan pengecekan manual. Tidak butuh login.

- `200` → `{ "status": "ok", "db": "up" }`
- `503` → `{ "status": "error", "db": "down" }` (ping database gagal dalam 2 detik)

---

### POST /api/login

Request:

```json
{ "username": "petugas01", "password": "petugas123" }
```

| Respons | Body | Catatan |
| --- | --- | --- |
| `200` | `{ "data": User }` | Header `Set-Cookie: sitangkap_session=<token>; Path=/; HttpOnly; SameSite=Lax; Max-Age=28800` (+ `; Secure` bila `COOKIE_SECURE=true`) |
| `422` | `VALIDATION_ERROR` | `username` atau `password` kosong. `fields.username`: `Username wajib diisi.`, `fields.password`: `Password wajib diisi.` |
| `401` | `INVALID_CREDENTIALS` | Username tidak ada, password salah, atau `aktif = 0` |
| `400` | `BAD_REQUEST` | JSON rusak |

Token: 32 byte dari `crypto/rand`, dienkode hex (64 karakter). Disimpan di memori server (map + mutex) bersama `user_id`, `role`, dan waktu kedaluwarsa (`SESSION_TTL_HOURS`, default 8). **Konsekuensi yang disengaja:** restart backend = semua pengguna harus login ulang.

---

### POST /api/logout

Tanpa body. Selalu `204`, baik ada sesi maupun tidak. Menghapus sesi di server dan mengirim `Set-Cookie: sitangkap_session=; Path=/; HttpOnly; SameSite=Lax; Max-Age=0`.

---

### GET /api/me

- `200` → `{ "data": User }`
- `401` → `UNAUTHORIZED`

Frontend memanggil ini setiap halaman `petugas` atau `dinas` dibuka (lihat panduan frontend, bagian penjaga halaman).

---

### GET /api/transaksi

Query (opsional, **harus berpasangan**):

| Param | Format | Contoh |
| --- | --- | --- |
| `dari` | `YYYY-MM-DD` | `2026-10-01` |
| `sampai` | `YYYY-MM-DD` | `2026-10-22` |

Aturan:
- Tanpa keduanya → semua transaksi; `meta.dari` dan `meta.sampai` bernilai `null`.
- Hanya salah satu, format salah, atau `dari` > `sampai` → `400 BAD_REQUEST`.
- Urutan: `tanggal` terbaru dulu, lalu `id` terbesar dulu.
- Tidak ada paginasi (target data ≤ 1.000 baris).

Respons `200`:

```json
{
  "data": [ Transaksi, ... ],
  "meta": { "jumlah": 5, "dari": "2026-10-01", "sampai": "2026-10-22", "tarif_retribusi": 0.03 }
}
```

`meta.tarif_retribusi` = nilai `RETRIBUSI_TARIF` yang sedang dipakai backend. Frontend petugas memakainya untuk pratinjau dan label tarif di form, sehingga pratinjau selalu sama dengan hasil simpan.

Daftar kosong → `"data": []`, **bukan** `null`.

---

### POST /api/transaksi

Peran: `petugas` (selain itu `403`). Request:

```json
{
  "nama_nelayan": "Pak Roni",
  "jenis_ikan": "Cakalang",
  "berat_kg": 120.5,
  "harga_per_kg": 28000,
  "tanggal": "2026-10-22"
}
```

- `201` → `{ "data": Transaksi }` (lengkap dengan `id`, `nilai`, `retribusi`, `dibuat_oleh`, `dibuat_pada`)
- `422` → `VALIDATION_ERROR` dengan **semua** field yang salah sekaligus di `fields`
- `400`, `401`, `403` sesuai tabel error

Field tak dikenal diabaikan. `dibuat_oleh` diambil dari sesi, **bukan** dari body.

Pesan validasi (dipakai persis oleh backend, ditampilkan apa adanya oleh frontend):

| Field | Kondisi | Pesan |
| --- | --- | --- |
| `nama_nelayan` | kosong / tidak ada | `Nama nelayan wajib diisi.` |
| `nama_nelayan` | > 100 karakter | `Nama nelayan paling banyak 100 karakter.` |
| `jenis_ikan` | kosong / tidak ada | `Jenis ikan wajib diisi.` |
| `jenis_ikan` | > 60 karakter | `Jenis ikan paling banyak 60 karakter.` |
| `berat_kg` | tidak ada / bukan angka | `Berat wajib diisi dengan angka.` |
| `berat_kg` | ≤ 0 | `Berat harus lebih dari 0.` |
| `berat_kg` | > 100000 | `Berat paling besar 100000 kg.` |
| `berat_kg` | > 2 desimal | `Berat paling banyak 2 angka di belakang koma.` |
| `harga_per_kg` | tidak ada / bukan bilangan bulat | `Harga per kg wajib diisi dengan bilangan bulat.` |
| `harga_per_kg` | ≤ 0 | `Harga per kg harus lebih dari 0.` |
| `harga_per_kg` | > 100000000 | `Harga per kg paling besar Rp 100.000.000.` |
| `tanggal` | tidak ada / format salah / tidak nyata | `Tanggal harus berformat YYYY-MM-DD dan valid.` |
| `tanggal` | setelah hari ini | `Tanggal tidak boleh melewati hari ini.` |

---

### PUT /api/transaksi/{id}

Peran: `petugas`. Body sama persis dengan POST (penggantian penuh, semua field wajib).

- `200` → `{ "data": Transaksi }` (data setelah diubah; `dibuat_oleh` dan `dibuat_pada` tidak berubah)
- `400` → `id` bukan bilangan bulat positif
- `404` → id tidak ada
- `422`, `401`, `403` sesuai tabel

---

### DELETE /api/transaksi/{id}

Peran: `petugas`.

- `204` → terhapus permanen
- `400` → `id` bukan bilangan bulat positif
- `404` → id tidak ada
- `401`, `403` sesuai tabel

---

### GET /api/rekap

Peran: `dinas`. Query `dari` dan `sampai` sama seperti `GET /api/transaksi`, dengan satu perbedaan: **tanpa keduanya → default dari tanggal 1 bulan ini sampai hari ini** (zona `APP_TIMEZONE`). Respons selalu menyebut rentang yang dipakai.

- `200` → `{ "data": Rekap }`. Tidak ada transaksi → semua total `0`, `jumlah_transaksi: 0`.
- `400`, `401`, `403` sesuai tabel.

Frontend dinas memanggil `/api/rekap` **dan** `/api/transaksi` dengan `dari`/`sampai` yang sama.

---

### POST /api/chat

Asisten tanya-jawab berbasis aturan, mirip AIML (FR-012, BR-10 s.d. BR-12). Backend mencocokkan kata kunci di pertanyaan, mengambil data dengan query yang sudah ada (Q-03 atau Q-04), lalu menyusun jawaban dari templat di bawah. **Tidak ada layanan AI luar, tidak ada dependensi baru, dan riwayat percakapan tidak disimpan.** Endpoint ini hanya membaca data.

Peran: `petugas` dan `dinas`. Request:

```json
{ "pesan": "berapa transaksi hari ini?" }
```

| Respons | Body | Catatan |
| --- | --- | --- |
| `200` | `{ "data": JawabanChat }` | Juga untuk pertanyaan yang tidak dikenali (`dikenali: false`). Itu bukan error |
| `422` | `VALIDATION_ERROR` | `fields.pesan`: `Pertanyaan wajib diisi.` (tidak ada, bukan string, atau kosong setelah trim) atau `Pertanyaan paling banyak 200 karakter.` |
| `400`, `401` | sesuai tabel error | JSON rusak, body > 10 KB, tidak ada sesi |

Isi `pesan` **tidak** dicatat ke log server (sama dengan aturan log body). Log cukup mencatat `niat`.

#### Langkah 1 · Normalisasi pesan

1. Trim, lalu ubah ke huruf kecil.
2. Ganti setiap karakter selain huruf `a-z`, angka `0-9`, `/`, dan `-` dengan spasi.
3. Rapatkan spasi berturut-turut menjadi satu.

Contoh: `"Berapa  Transaksi, HARI INI?"` menjadi `"berapa transaksi hari ini"`.

Kata kunci dicocokkan sebagai **kata utuh** atau **frasa utuh** (dibatasi spasi atau awal/akhir pesan), bukan potongan kata. Contoh: `kg` cocok di `"berat dalam kg"` tetapi tidak di `"kgx"`.

#### Langkah 2 · Rentang tanggal

"Hari ini" mengikuti `APP_TIMEZONE`. Aturan diperiksa dari atas, yang pertama cocok yang dipakai:

| Urutan | Pola di pesan | `dari` | `sampai` |
| --- | --- | --- | --- |
| R1 | Dua tanggal (`YYYY-MM-DD` atau `DD/MM/YYYY`) | tanggal pertama | tanggal kedua |
| R2 | Satu tanggal | tanggal itu | tanggal itu |
| R3 | `hari ini` | hari ini | hari ini |
| R4 | `kemarin` | kemarin | kemarin |
| R5 | `minggu ini` | Senin minggu ini | hari ini |
| R6 | `bulan lalu` | tanggal 1 bulan lalu | tanggal terakhir bulan lalu |
| R7 | `bulan ini` | tanggal 1 bulan ini | hari ini |
| R8 | `tahun ini` | 1 Januari tahun ini | hari ini |
| R9 | Nama bulan (`januari` … `desember`), boleh diikuti tahun 4 angka | tanggal 1 bulan itu | tanggal terakhir bulan itu. Tanpa tahun = tahun ini |
| R10 | Tidak ada pola di atas | tanggal 1 bulan ini | hari ini (sama dengan default `GET /api/rekap`) |

Tanggal yang ditulis tetapi tidak nyata (misalnya `2026-02-30` atau `31/04/2026`), atau tanggal pertama setelah tanggal kedua di R1, menghasilkan `dikenali: false` dengan `jawaban` dari baris "Tanggal tidak valid" di tabel templat.

#### Langkah 3 · Niat (intent)

Diperiksa dari atas, yang pertama cocok yang dipakai. Satu pesan hanya menghasilkan satu niat.

| Urutan | `niat` | Cocok bila pesan mengandung | Memakai rentang? |
| --- | --- | --- | --- |
| N1 | `bantuan` | `bantuan`, `help`, `bisa apa`, `contoh` | Tidak |
| N2 | `ikan_terbanyak` | `ikan` **dan** salah satu dari `terbanyak`, `paling banyak`, `terlaris` | Ya |
| N3 | `transaksi_terakhir` | `terakhir` atau `terbaru` | Tidak, selalu seluruh data (Q-03, baris pertama) |
| N4 | `retribusi` | `retribusi` | Ya |
| N5 | `nilai` | `nilai`, `pendapatan`, `omzet`, `uang` | Ya |
| N6 | `berat` | `berat`, `kg`, `kilo` | Ya |
| N7 | `jumlah` | `transaksi` **dan** salah satu dari `berapa`, `jumlah`, `banyak` | Ya |
| N8 | `rekap` | `rekap`, `ringkasan`, `total`, `laporan`, **atau** salah satu pola R1–R9 ditemukan tanpa kata kunci niat lain | Ya |
| N9 | `sapaan` | `halo`, `hai`, `hi`, `selamat pagi`, `selamat siang`, `selamat sore`, `selamat malam` | Tidak |
| — | tidak dikenali | Tidak ada yang cocok | — |

Angka untuk niat yang memakai rentang dihitung dengan fungsi yang **sama** dengan `GET /api/rekap` (aturan bagian 3), sehingga jawaban chat selalu sama dengan kartu rekap untuk rentang yang sama.

`ikan_terbanyak`: transaksi di rentang dikelompokkan per `jenis_ikan` (persis sama, peka huruf besar/kecil), lalu dipilih kelompok dengan total `berat_kg` terbesar. Kalau seri, pilih `jenis_ikan` yang urutan abjadnya paling awal.

#### Langkah 4 · Templat jawaban

Format angka di jawaban sama dengan tampilan frontend:

- Rupiah: `Rp 11.577.000`
- Berat: `370,75 kg`. Tanpa desimal bila bulat (`45 kg`), satu atau dua desimal sesuai nilainya (`120,5 kg`).
- Tanggal: `DD/MM/YYYY`
- Persen tarif: `3` untuk 0.03, `2,85` untuk 0.0285

`{periode}` berisi `pada 22/10/2026` bila `dari = sampai`, selain itu `pada 01/10/2026 sampai 22/10/2026`.

| Kondisi | `jawaban` |
| --- | --- |
| Niat memakai rentang dan tidak ada transaksi di rentang itu | `Tidak ada transaksi {periode}.` |
| `rekap` | `Rekap {periode}: {jumlah} transaksi, total berat {berat}, nilai {nilai}, retribusi {retribusi} (tarif {persen}%).` |
| `jumlah` | `Ada {jumlah} transaksi {periode}.` |
| `berat` | `Total berat ikan {periode} adalah {berat}.` |
| `nilai` | `Total nilai transaksi {periode} adalah {nilai}.` |
| `retribusi` | `Total retribusi {periode} adalah {retribusi} (tarif {persen}%).` |
| `ikan_terbanyak` | `Ikan dengan berat terbanyak {periode} adalah {jenis_ikan}: {berat} dari {jumlah} transaksi.` |
| `transaksi_terakhir` | `Transaksi terbaru: {nama_nelayan}, {jenis_ikan} {berat} dengan harga {harga_per_kg}/kg pada {tanggal}, nilai {nilai}.` |
| `transaksi_terakhir`, data kosong | `Belum ada transaksi yang dicatat.` |
| `sapaan` | `Halo, {username}! Tanyakan rekap, jumlah transaksi, berat, nilai, atau retribusi. Contoh: "rekap bulan ini".` |
| `bantuan` | `Contoh pertanyaan: "rekap bulan ini", "berapa transaksi hari ini", "total retribusi oktober 2026", "ikan terbanyak minggu ini", "transaksi terakhir".` |
| Tanggal tidak valid | `Tanggal di pertanyaan tidak valid. Tulis tanggal seperti 2026-10-22 atau 22/10/2026.` |
| Tidak dikenali | `Maaf, saya belum mengerti pertanyaan itu. Ketik "bantuan" untuk melihat contoh pertanyaan.` |

#### Contoh dengan data `db/seed.sql` (hari ini = 2026-10-22, tarif 0.03)

| `pesan` | `niat` | `jawaban` |
| --- | --- | --- |
| `rekap dari 01/10/2026 sampai 22/10/2026` | `rekap` | `Rekap pada 01/10/2026 sampai 22/10/2026: 5 transaksi, total berat 370,75 kg, nilai Rp 11.577.000, retribusi Rp 347.310 (tarif 3%).` |
| `Berapa transaksi hari ini?` | `jumlah` | `Ada 3 transaksi pada 22/10/2026.` |
| `total retribusi hari ini` | `retribusi` | `Total retribusi pada 22/10/2026 adalah Rp 208.935 (tarif 3%).` |
| `berat ikan kemarin` | `berat` | `Total berat ikan pada 21/10/2026 adalah 175 kg.` |
| `nilai september 2026` | `nilai` | `Total nilai transaksi pada 01/09/2026 sampai 30/09/2026 adalah Rp 4.500.000.` |
| `ikan terbanyak bulan ini` | `ikan_terbanyak` | `Ikan dengan berat terbanyak pada 01/10/2026 sampai 22/10/2026 adalah Cakalang: 215,5 kg dari 2 transaksi.` |
| `transaksi terakhir` | `transaksi_terakhir` | `Transaksi terbaru: Ibu Merry, Julung-julung 30,25 kg dengan harga Rp 22.000/kg pada 22/10/2026, nilai Rp 665.500.` |
| `rekap januari 2025` | `rekap` | `Tidak ada transaksi pada 01/01/2025 sampai 31/01/2025.` |
| `rekap 2026-02-30` | — (`dikenali: false`) | `Tanggal di pertanyaan tidak valid. Tulis tanggal seperti 2026-10-22 atau 22/10/2026.` |
| `cuaca besok bagaimana` | — (`dikenali: false`) | `Maaf, saya belum mengerti pertanyaan itu. Ketik "bantuan" untuk melihat contoh pertanyaan.` |

## 5. Konfigurasi bersama (environment variables)

### Backend (`backend/.env`, tidak di-commit; contoh di `backend/.env.example`)

| Variabel | Default | Keterangan |
| --- | --- | --- |
| `PORT` | `3000` | Port HTTP backend |
| `DB_HOST` | `127.0.0.1` | |
| `DB_PORT` | `3306` | |
| `DB_USER` | `sitangkap_app` | User aplikasi dari `db/app_user.sql` (lokal) atau CI-02 (server) |
| `DB_PASSWORD` | *(wajib)* | Lokal: `sitangkap_dev` dari `db/app_user.sql`. Server: password kuat dari CI-02. Pakai huruf dan angka saja |
| `DB_NAME` | `sitangkap` | |
| `RETRIBUSI_TARIF` | `0.03` | Desimal, 0 < tarif < 1. **Contoh**, konfirmasi ke perda |
| `APP_TIMEZONE` | `Asia/Makassar` | Penentu "hari ini" dan default rekap |
| `SESSION_TTL_HOURS` | `8` | Masa berlaku sesi |
| `COOKIE_SECURE` | `false` | `true` hanya bila sudah HTTPS |
| `STATIC_DIR` | *(kosong)* | Kosong saat development. Saat production diisi path hasil build frontend; backend lalu melayani file statis di luar `/api` dan `/health` |

### Frontend (`frontend/.env`, tidak di-commit)

| Variabel | Default | Keterangan |
| --- | --- | --- |
| `PUBLIC_MOCK` | `0` | `1` = frontend memakai data palsu tanpa backend (untuk bekerja paralel). Lihat panduan frontend |

### Port

| Layanan | Port |
| --- | --- |
| Backend Go | 3000 |
| Astro dev server | 4321 |
| MySQL | 3306 |
| Jenkins (server, fase 2) | 8080 |

## 6. Akun demo (dari `db/seed.sql`)

| Username | Password | Role |
| --- | --- | --- |
| `petugas01` | `petugas123` | `petugas` |
| `dinas01` | `dinas123` | `dinas` |

Hanya untuk development dan demo. Ganti sebelum dipakai sungguhan.

## 7. Uji cepat dengan curl

```bash
# login, simpan cookie ke file
curl -i -c cookie.txt -H "Content-Type: application/json" \
  -d '{"username":"petugas01","password":"petugas123"}' http://localhost:3000/api/login

curl -b cookie.txt http://localhost:3000/api/me
curl -b cookie.txt "http://localhost:3000/api/transaksi?dari=2026-10-01&sampai=2026-10-22"

curl -i -b cookie.txt -H "Content-Type: application/json" \
  -d '{"nama_nelayan":"Pak Roni","jenis_ikan":"Cakalang","berat_kg":120.5,"harga_per_kg":28000,"tanggal":"2026-10-22"}' \
  http://localhost:3000/api/transaksi

# login sebagai dinas di file cookie lain, lalu rekap
curl -s -c dinas.txt -H "Content-Type: application/json" \
  -d '{"username":"dinas01","password":"dinas123"}' http://localhost:3000/api/login
curl -b dinas.txt "http://localhost:3000/api/rekap?dari=2026-10-01&sampai=2026-10-22"

# chat (petugas maupun dinas)
curl -b dinas.txt -H "Content-Type: application/json" \
  -d '{"pesan":"rekap dari 01/10/2026 sampai 22/10/2026"}' http://localhost:3000/api/chat
```

## 8. Mengubah kontrak

1. Siapa pun yang butuh perubahan membuat issue berlabel `kontrak` berisi: apa yang diubah, alasannya, dan siapa yang terdampak.
2. PM memperbarui dokumen ini lewat pull request dan menandai (`@`) pemilik peran yang terdampak.
3. Baru setelah PR itu digabung, kode boleh diubah. Kode yang menyimpang dari dokumen ini dianggap bug.
