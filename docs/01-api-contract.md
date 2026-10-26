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
  "metrik": null,
  "dari": "2026-10-22",
  "sampai": "2026-10-22",
  "konteks": { "niat": "jumlah", "metrik": null, "dari": "2026-10-22", "sampai": "2026-10-22", "gaya": "netral" }
}
```

| Field | Tipe | Keterangan |
| --- | --- | --- |
| `jawaban` | string | Kalimat Bahasa Indonesia yang siap ditampilkan apa adanya |
| `dikenali` | boolean | `false` bila pertanyaan tidak cocok dengan aturan mana pun atau tanggalnya tidak valid |
| `niat` | string atau `null` | Kode niat dari tabel niat di `POST /api/chat`. `null` bila `dikenali = false` |
| `metrik` | array string atau `null` | Hanya untuk niat `gabungan`: metrik yang dijawab, urut `jumlah`, `berat`, `nilai`, `retribusi`. Selain itu `null` |
| `dari`, `sampai` | string `YYYY-MM-DD` atau `null` | Rentang yang dipakai untuk menjawab. `null` untuk niat yang tidak memakai rentang dan bila `dikenali = false` |
| `konteks` | objek Konteks chat | Selalu ada. Frontend menyimpannya di memori halaman dan mengirimnya kembali di pertanyaan berikutnya |

### Konteks chat

```json
{ "niat": "retribusi", "metrik": null, "dari": "2026-10-22", "sampai": "2026-10-22", "gaya": "santai" }
```

| Field | Tipe | Keterangan |
| --- | --- | --- |
| `niat` | string atau `null` | Niat data terakhir yang memakai rentang: `rekap`, `jumlah`, `berat`, `nilai`, `retribusi`, `gabungan`, atau `ikan_terbanyak` |
| `metrik` | array string atau `null` | Diisi hanya bila `niat = "gabungan"` |
| `dari`, `sampai` | string `YYYY-MM-DD` atau `null` | Rentang dari niat data terakhir |
| `gaya` | string | `santai`, `formal`, atau `netral` |

Konteks hanya hidup di memori halaman frontend. Server tidak menyimpannya (BR-12).

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

Asisten tanya-jawab "Si Tangkap", berbasis aturan (FR-012, BR-10 s.d. BR-12). Backend mencocokkan kata kunci di pertanyaan, mengambil data dengan query yang sudah ada (Q-03 atau Q-04), lalu menyusun jawaban dari templat di bawah. Selain pertanyaan data, Si Tangkap juga bisa diajak bercakap ringan (sapaan, terima kasih, dan sejenisnya), mengikuti gaya bahasa pengguna, mengerti slang dan salah ketik, dan mengerti pertanyaan lanjutan lewat `konteks`. **Tidak ada layanan AI luar, tidak ada dependensi baru, dan server tidak menyimpan riwayat percakapan.** Endpoint ini hanya membaca data.

Peran: `petugas` dan `dinas`. Request:

```json
{ "pesan": "kalau kemarin?", "konteks": { "niat": "retribusi", "metrik": null, "dari": "2026-10-22", "sampai": "2026-10-22", "gaya": "santai" } }
```

`konteks` opsional. Isinya adalah `konteks` dari respons sebelumnya, dikirim apa adanya. Field `konteks` yang tidak valid diabaikan satu per satu, **bukan** error:

- `niat` bukan salah satu niat data di tabel Konteks chat → dianggap `null`.
- `metrik` bukan array metrik yang valid, atau `niat` bukan `gabungan` → dianggap `null`. Bila `niat = "gabungan"` tetapi `metrik` tidak valid, `niat` juga dianggap `null`.
- `dari`/`sampai` bukan tanggal nyata, hanya salah satu, atau `dari` > `sampai` → keduanya dianggap `null`.
- `gaya` bukan `santai`/`formal`/`netral` → dianggap `netral`.

| Respons | Body | Catatan |
| --- | --- | --- |
| `200` | `{ "data": JawabanChat }` | Juga untuk pertanyaan yang tidak dikenali (`dikenali: false`). Itu bukan error |
| `422` | `VALIDATION_ERROR` | `fields.pesan`: `Pertanyaan wajib diisi.` (tidak ada, bukan string, atau kosong setelah trim) atau `Pertanyaan paling banyak 200 karakter.` (dihitung per karakter, bukan byte) |
| `400`, `401` | sesuai tabel error | JSON rusak, body > 10 KB, tidak ada sesi |

Isi `pesan` **tidak** dicatat ke log server (sama dengan aturan log body). Log cukup mencatat `niat`.

#### Langkah 1 · Normalisasi pesan

1. Trim, lalu ubah ke huruf kecil.
2. Ganti setiap karakter selain huruf `a-z`, angka `0-9`, `/`, dan `-` dengan spasi.
3. Rapatkan spasi berturut-turut menjadi satu. Hasilnya dipecah menjadi token per spasi.

Contoh: `"Berapa  Transaksi, HARI INI?"` menjadi `"berapa transaksi hari ini"`.

#### Langkah 2 · Gaya bahasa

Dihitung dari token hasil langkah 1, **sebelum** slang diganti.

| Gaya | Token penanda |
| --- | --- |
| `santai` | `bro`, `gan`, `sis`, `bang`, `bos`, `cuy`, `coy`, `dong`, `sih`, `nih`, `deh`, `gue`, `gw`, `lu`, `lo`, `aja`, `banget`, `gimana`, `gak`, `ga`, `nggak`, `kok`, `yuk`, `makasih`, `thx`, `mantap`, `sip`, `brp`, `trx`, `kmrn`, serta token tawa (lihat langkah 3) |
| `formal` | `anda`, `bapak`, `ibu`, `saudara`, `mohon`, `silakan`, `apakah`, `bagaimanakah`, `permisi`, `selamat`, `terima` |

Lebih banyak penanda `santai` → `santai`. Lebih banyak penanda `formal` → `formal`. Seri (termasuk sama-sama nol) → `konteks.gaya` bila ada, selain itu `netral`.

#### Langkah 3 · Slang dan singkatan

Setiap token yang **persis sama** dengan kolom kiri diganti dengan kolom kanan (boleh lebih dari satu kata).

| Token | Diganti menjadi |
| --- | --- |
| `brp`, `brapa`, `berapakah` | `berapa` |
| `trx`, `trans`, `transaksinya`, `transaksi2` | `transaksi` |
| `hr`, `hri` | `hari` |
| `kmrn`, `kmarin`, `kemaren`, `kmaren` | `kemarin` |
| `skrg`, `skarang`, `sekarang` | `hari ini` |
| `bln` | `bulan` |
| `thn` | `tahun` |
| `mgg`, `mgu` | `minggu` |
| `tgl` | `tanggal` |
| `rekapan`, `rekapnya` | `rekap` |
| `retribusinya` | `retribusi` |
| `beratnya` | `berat` |
| `nilainya` | `nilai` |
| `jumlahnya` | `jumlah` |
| `ikannya` | `ikan` |
| `duit` | `uang` |
| `gmn`, `gimana` | `bagaimana` |
| `makasih`, `makasi`, `makacih`, `mksh`, `thx`, `thanks`, `tq`, `trims`, `tengkyu` | `terima kasih` |
| `ok`, `okay`, `okey`, `okee`, `okeh`, `sip`, `siap`, `mantap`, `mantab` | `oke` |
| `gk`, `ga`, `gak`, `nggak`, `ngga`, `enggak`, `engga`, `tdk` | `tidak` |
| `hallo`, `haloo`, `hello`, `helo`, `hey`, `hei`, `hay` | `halo` |
| Token tawa: pola `wk` berulang (`wkwk`, `wkwkwk`, `wkwkw`), `ha` berulang minimal dua kali (`haha`, `hahaha`), `he` berulang minimal dua kali (`hehe`) | `wkwk`, `haha`, `hehe` |

#### Langkah 4 · Salah ketik

Token yang panjangnya 5 huruf atau lebih, hanya berisi huruf, dan **tidak** persis sama dengan kata kunci mana pun, diganti dengan kata kunci pertama di daftar ini yang jaraknya 1. Jarak dihitung dengan *optimal string alignment* (Levenshtein ditambah pertukaran dua huruf bersebelahan), jadi `retribsi`, `transkasi`, dan `kemarn` tetap dikenali.

Daftar kata kunci, berurutan: `transaksi`, `retribusi`, `berapa`, `kemarin`, `ringkasan`, `laporan`, `terbanyak`, `terakhir`, `terbaru`, `pendapatan`, `penjualan`, `bantuan`, `jumlah`, `rekap`, `nilai`, `berat`, `bulan`, `tahun`, `minggu`, `januari`, `februari`, `maret`, `april`, `agustus`, `september`, `oktober`, `november`, `desember`, `terima`, `kasih`, `siapa`, `kabar`, `sampai`.

Hasil langkah 1, 3, dan 4 disebut **pesan akhir**, misalnya `"brp trx kmrn"` menjadi `"berapa transaksi kemarin"`. Semua pencocokan di bawah memakai pesan akhir, per **kata utuh** atau **frasa utuh** (dibatasi spasi atau awal/akhir pesan).

#### Langkah 5 · Rentang tanggal

"Hari ini" mengikuti `APP_TIMEZONE`. Aturan diperiksa dari atas, yang pertama cocok yang dipakai:

| Urutan | Pola di pesan | `dari` | `sampai` |
| --- | --- | --- | --- |
| R1 | Dua tanggal (`YYYY-MM-DD` atau `DD/MM/YYYY`) | tanggal pertama | tanggal kedua |
| R2 | Satu tanggal | tanggal itu | tanggal itu |
| R3 | `hari ini` | hari ini | hari ini |
| R4 | `kemarin` | kemarin | kemarin |
| R5 | `minggu lalu` | Senin minggu lalu | Minggu minggu lalu |
| R6 | `minggu ini` | Senin minggu ini | hari ini |
| R7 | `bulan lalu` | tanggal 1 bulan lalu | tanggal terakhir bulan lalu |
| R8 | `bulan ini` | tanggal 1 bulan ini | hari ini |
| R9 | `tahun lalu` | 1 Januari tahun lalu | 31 Desember tahun lalu |
| R10 | `tahun ini` | 1 Januari tahun ini | hari ini |
| R11 | Nama bulan (`januari` … `desember`), boleh diikuti tahun 4 angka | tanggal 1 bulan itu | tanggal terakhir bulan itu. Tanpa tahun = tahun ini |
| R12 | Tidak ada pola di atas | `konteks.dari` | `konteks.sampai`. Bila konteks tidak punya rentang: tanggal 1 bulan ini sampai hari ini (sama dengan default `GET /api/rekap`) |

Tanggal yang ditulis tetapi tidak nyata (misalnya `2026-02-30` atau `31/04/2026`), atau tanggal pertama setelah tanggal kedua di R1, menghasilkan `dikenali: false` dengan `jawaban` "Tanggal tidak valid".

#### Langkah 6 · Niat

Diperiksa dari atas, yang pertama cocok yang dipakai. Satu pesan menghasilkan satu niat.

**Niat data** (memakai data transaksi):

| Urutan | `niat` | Cocok bila pesan akhir mengandung | Memakai rentang? |
| --- | --- | --- | --- |
| N1 | `bantuan` | `bantuan`, `help`, `bisa apa`, `contoh`, `menu` | Tidak |
| N2 | `ikan_terbanyak` | `ikan` **dan** salah satu dari `terbanyak`, `paling banyak`, `terlaris` | Ya |
| N3 | `transaksi_terakhir` | `terakhir` atau `terbaru` | Tidak, selalu seluruh data (Q-03, baris pertama) |
| N4 | metrik | Satu atau lebih metrik di bawah. Satu metrik → niat bernama metrik itu. Dua atau tiga → `gabungan` (field `metrik` berisi daftarnya). Keempatnya → `rekap` | Ya |
| N5 | `rekap` | `rekap`, `ringkasan`, `total`, `laporan` | Ya |
| N6 | lanjutan | Tidak ada niat N1–N5, tetapi ada pola R1–R11 | Ya. Niat (dan `metrik`) diambil dari `konteks`. Konteks tanpa niat → `rekap` |

Metrik untuk N4:

| Metrik | Cocok bila pesan akhir mengandung |
| --- | --- |
| `jumlah` | `transaksi` **dan** salah satu dari `berapa`, `jumlah`, `banyak`, `ada` |
| `berat` | `berat`, `kg`, `kilo`, `kilogram` |
| `nilai` | `nilai`, `pendapatan`, `penjualan`, `omzet`, `uang` |
| `retribusi` | `retribusi`, `pungutan` |

**Niat percakapan** (diperiksa setelah N6, tidak memakai data):

| Urutan | `niat` | Cocok bila pesan akhir mengandung |
| --- | --- | --- |
| P1 | `siapa_kamu` | `siapa kamu`, `kamu siapa`, `nama kamu`, `namamu`, `kamu bot`, `kamu robot`, `kamu manusia` |
| P2 | `apa_kabar` | `apa kabar`, `bagaimana kabar`, `kabar kamu`, `kabarmu` |
| P3 | `terima_kasih` | `terima kasih` |
| P4 | `maaf` | `maaf`, `sorry`, `sori` |
| P5 | `pamit` | `dadah`, `bye`, `sampai jumpa`, `pamit`, `selamat tinggal` |
| P6 | `capek` | `capek`, `cape`, `lelah`, `pusing`, `bosan`, `ngantuk` |
| P7 | `tawa` | `wkwk`, `haha`, `hehe`, `lol` |
| P8 | `oke` | `oke` |
| P9 | `tidak` | Token pertama `tidak` dan pesan akhir paling banyak 3 token |
| P10 | `sapaan` | `halo`, `hai`, `hi`, `permisi`, `assalamualaikum`, `selamat pagi`, `selamat siang`, `selamat sore`, `selamat malam` |
| — | tidak dikenali | Tidak ada yang cocok |

Karena niat data diperiksa lebih dulu, `"halo, rekap bulan ini"` dijawab sebagai rekap dan `"makasih, retribusi kemarin dong"` dijawab sebagai retribusi.

Angka untuk niat data yang memakai rentang dihitung dengan fungsi yang **sama** dengan `GET /api/rekap` (aturan bagian 3), sehingga jawaban chat selalu sama dengan kartu rekap untuk rentang yang sama.

`ikan_terbanyak`: transaksi di rentang dikelompokkan per `jenis_ikan` (persis sama, peka huruf besar/kecil), lalu dipilih kelompok dengan total `berat_kg` terbesar. Kalau seri, pilih `jenis_ikan` yang urutan abjadnya paling awal.

#### Langkah 7 · Konteks di respons

| Kondisi | `konteks` di respons |
| --- | --- |
| Niat data yang memakai rentang (`rekap`, `jumlah`, `berat`, `nilai`, `retribusi`, `gabungan`, `ikan_terbanyak`) | `niat`, `metrik`, `dari`, `sampai` dari jawaban ini; `gaya` dari langkah 2 |
| Niat lain, tidak dikenali, atau tanggal tidak valid | `niat`, `metrik`, `dari`, `sampai` dari konteks request (setelah validasi; `null` bila tidak ada); `gaya` dari langkah 2 |

#### Langkah 8 · Templat jawaban

Format angka di jawaban sama dengan tampilan frontend:

- Rupiah: `Rp 11.577.000`
- Berat: `370,75 kg`. Tanpa desimal bila bulat (`45 kg`), satu atau dua desimal sesuai nilainya (`120,5 kg`).
- Tanggal: `DD/MM/YYYY`
- Persen tarif: `3` untuk 0.03, `2,85` untuk 0.0285

`{periode}` berisi `pada 22/10/2026` bila `dari = sampai`, selain itu `pada 01/10/2026 sampai 22/10/2026`. `{Periode}` sama dengan huruf awal kapital (`Pada …`). `{username}` adalah username pengguna yang login.

**Pemilihan variasi.** Kalau satu kondisi punya beberapa variasi, indeks variasi = FNV-1a 32-bit dari pesan akhir (Go: `hash/fnv`, `New32a`) modulo jumlah variasi. Variasi dihitung mulai dari 0. Pertanyaan yang sama selalu mendapat jawaban yang sama, sehingga bisa dites persis.

Templat niat data (sama untuk semua gaya):

| Kondisi | Variasi 0 | Variasi 1 |
| --- | --- | --- |
| Niat memakai rentang dan tidak ada transaksi di rentang itu | `Tidak ada transaksi {periode}.` | `Belum ada transaksi yang tercatat {periode}.` |
| `rekap` | `Rekap {periode}: {jumlah} transaksi, total berat {berat}, nilai {nilai}, retribusi {retribusi} (tarif {persen}%).` | `{Periode} tercatat {jumlah} transaksi dengan total berat {berat}, nilai {nilai}, dan retribusi {retribusi} (tarif {persen}%).` |
| `jumlah` | `Ada {jumlah} transaksi {periode}.` | `{Periode} tercatat {jumlah} transaksi.` |
| `berat` | `Total berat ikan {periode} adalah {berat}.` | `{Periode} ikan yang terjual totalnya {berat}.` |
| `nilai` | `Total nilai transaksi {periode} adalah {nilai}.` | `{Periode} nilai transaksinya mencapai {nilai}.` |
| `retribusi` | `Total retribusi {periode} adalah {retribusi} (tarif {persen}%).` | `Retribusi yang terkumpul {periode} sebesar {retribusi} (tarif {persen}%).` |
| `ikan_terbanyak` | `Ikan dengan berat terbanyak {periode} adalah {jenis_ikan}: {berat} dari {jumlah} transaksi.` | `{Periode} yang paling banyak adalah {jenis_ikan}, totalnya {berat} dari {jumlah} transaksi.` |
| `gabungan` | `{Periode}: {bagian}.` (satu variasi) | — |
| `transaksi_terakhir` | `Transaksi terbaru: {nama_nelayan}, {jenis_ikan} {berat} dengan harga {harga_per_kg}/kg pada {tanggal}, nilai {nilai}.` (satu variasi) | — |
| `transaksi_terakhir`, data kosong | `Belum ada transaksi yang dicatat.` (satu variasi) | — |
| Tanggal tidak valid | `Tanggal di pertanyaan tidak valid. Tulis tanggal seperti 2026-10-22 atau 22/10/2026.` (satu variasi) | — |

`{bagian}` untuk `gabungan` adalah potongan per metrik, urut `jumlah`, `berat`, `nilai`, `retribusi`: `{jumlah} transaksi`, `total berat {berat}`, `nilai {nilai}`, `retribusi {retribusi} (tarif {persen}%)`. Dua potongan digabung dengan ` dan `. Tiga potongan menjadi `A, B, dan C`.

Templat percakapan, bantuan, dan tidak dikenali dibedakan menurut gaya. Kolom "Biasa" dipakai untuk gaya `formal` dan `netral`.

| `niat` | Biasa, variasi 0 | Biasa, variasi 1 | Santai, variasi 0 | Santai, variasi 1 |
| --- | --- | --- | --- | --- |
| `sapaan` | `Halo, {username}! Saya Si Tangkap, asisten data TPI. Mau lihat rekap, jumlah transaksi, berat, nilai, atau retribusi?` | `Selamat datang, {username}. Silakan tanyakan data lelang, misalnya "rekap bulan ini".` | `Halo {username}! Si Tangkap di sini. Mau cek rekap atau retribusi hari ini?` | `Hai {username}! Ada yang mau dicek? Coba aja "berapa transaksi hari ini".` |
| `siapa_kamu` | `Saya Si Tangkap, asisten yang membantu membaca data lelang ikan di TPI ini. Saya bisa menjawab rekap, jumlah transaksi, berat, nilai, dan retribusi.` | `Nama saya Si Tangkap. Saya asisten berbasis aturan untuk data transaksi lelang TPI, jadi jawaban saya selalu diambil dari data yang tercatat.` | `Aku Si Tangkap, asisten data TPI. Tanya aja soal rekap, berat, nilai, atau retribusi.` | `Si Tangkap di sini! Kerjaanku bacain data lelang ikan biar kamu nggak perlu hitung manual.` |
| `apa_kabar` | `Kabar baik, terima kasih. Ada data lelang yang ingin Anda lihat?` | `Baik, terima kasih sudah bertanya. Mau saya bantu cek rekap hari ini?` | `Baik dong! Kamu gimana? Mau cek data hari ini?` | `Aman! Siap bantu cek data lelang nih.` |
| `terima_kasih` | `Sama-sama. Kalau perlu data lain, tanyakan saja.` | `Dengan senang hati. Semoga membantu.` | `Sama-sama! Kalau butuh lagi, panggil aja.` | `Siap, sama-sama!` |
| `maaf` | `Tidak apa-apa. Silakan lanjutkan pertanyaannya.` | `Tidak masalah. Ada yang bisa saya bantu?` | `Santai aja, nggak apa-apa.` | `Aman kok. Lanjut aja.` |
| `pamit` | `Sampai jumpa. Semoga pekerjaannya lancar.` | `Terima kasih sudah memakai Si Tangkap. Sampai jumpa.` | `Dadah! Semangat kerjanya.` | `Oke, sampai nanti!` |
| `capek` | `Semoga cepat pulih. Kalau mau, saya bisa merangkum data supaya pekerjaan Anda lebih ringan.` | `Istirahat sebentar tidak apa-apa. Saya siap merangkum data kapan saja.` | `Waduh, istirahat dulu ya. Nanti biar aku yang rangkum datanya.` | `Semangat! Biar aku yang hitung, kamu cukup tanya.` |
| `tawa` | `Senang Anda terhibur. Ada data yang ingin dicek?` | `Hehe. Ada lagi yang bisa saya bantu?` | `Wkwk. Lanjut cek data?` | `Haha, ada-ada aja. Mau cek apa lagi?` |
| `oke` | `Baik. Ada hal lain yang ingin Anda tanyakan?` | `Baik, silakan bila ada pertanyaan lain.` | `Sip! Mau cek apa lagi?` | `Oke, gas. Ada lagi?` |
| `tidak` | `Baik, tidak masalah. Saya di sini kalau dibutuhkan.` | `Baik. Silakan tanyakan kapan saja.` | `Oke deh. Kalau butuh, panggil aja.` | `Siap, santai aja.` |
| `bantuan` | `Contoh pertanyaan: "rekap bulan ini", "berapa transaksi hari ini", "total retribusi oktober 2026", "ikan terbanyak minggu ini", "transaksi terakhir".` (satu variasi) | — | `Coba tanya kayak gini: "rekap bulan ini", "brp trx hari ini", "retribusi kemarin", "ikan terbanyak minggu ini", atau "transaksi terakhir".` (satu variasi) | — |
| tidak dikenali | `Maaf, saya belum mengerti pertanyaan itu. Saya bisa membantu soal rekap, jumlah transaksi, berat, nilai, dan retribusi. Ketik "bantuan" untuk contoh.` | `Pertanyaan itu di luar yang saya pahami. Coba tanyakan data lelang, misalnya "rekap bulan ini".` | `Hmm, aku belum nangkep maksudnya. Coba tanya soal rekap, berat, nilai, atau retribusi ya. Ketik "bantuan" buat contoh.` | `Waduh, itu di luar keahlianku. Kalau soal data lelang, aku jago. Coba "rekap bulan ini".` |

#### Contoh satu pertanyaan (data `db/seed.sql`, hari ini = 2026-10-22, tarif 0.03, tanpa konteks)

| `pesan` | `niat` | `jawaban` |
| --- | --- | --- |
| `rekap dari 01/10/2026 sampai 22/10/2026` | `rekap` | `Rekap pada 01/10/2026 sampai 22/10/2026: 5 transaksi, total berat 370,75 kg, nilai Rp 11.577.000, retribusi Rp 347.310 (tarif 3%).` |
| `Berapa transaksi hari ini?` | `jumlah` | `Ada 3 transaksi pada 22/10/2026.` |
| `total retribusi hari ini` | `retribusi` | `Total retribusi pada 22/10/2026 adalah Rp 208.935 (tarif 3%).` |
| `total retribsi hr ini` | `retribusi` | Sama persis dengan baris di atas (salah ketik dan singkatan dikenali, pesan akhirnya sama) |
| `brp trx kmrn?` | `jumlah` | `Pada 21/10/2026 tercatat 2 transaksi.` |
| `berat ikan kemarin` | `berat` | `Pada 21/10/2026 ikan yang terjual totalnya 175 kg.` |
| `nilai september 2026` | `nilai` | `Pada 01/09/2026 sampai 30/09/2026 nilai transaksinya mencapai Rp 4.500.000.` |
| `berat sama nilai hari ini` | `gabungan` (`metrik: ["berat","nilai"]`) | `Pada 22/10/2026: total berat 195,75 kg dan nilai Rp 6.964.500.` |
| `ikan terbanyak bulan ini` | `ikan_terbanyak` | `Pada 01/10/2026 sampai 22/10/2026 yang paling banyak adalah Cakalang, totalnya 215,5 kg dari 2 transaksi.` |
| `transaksi terakhir` | `transaksi_terakhir` | `Transaksi terbaru: Ibu Merry, Julung-julung 30,25 kg dengan harga Rp 22.000/kg pada 22/10/2026, nilai Rp 665.500.` |
| `rekap januari 2025` | `rekap` | `Belum ada transaksi yang tercatat pada 01/01/2025 sampai 31/01/2025.` |
| `rekap 2026-02-30` | — (`dikenali: false`) | `Tanggal di pertanyaan tidak valid. Tulis tanggal seperti 2026-10-22 atau 22/10/2026.` |
| `cuaca besok bagaimana` | — (`dikenali: false`) | `Pertanyaan itu di luar yang saya pahami. Coba tanyakan data lelang, misalnya "rekap bulan ini".` |

#### Contoh percakapan berlanjut

Setiap baris mengirim `konteks` dari respons baris sebelumnya.

Percakapan A, login `petugas01`:

| `pesan` | `niat` | `jawaban` | `konteks` di respons |
| --- | --- | --- | --- |
| `halo bro` | `sapaan` | `Halo petugas01! Si Tangkap di sini. Mau cek rekap atau retribusi hari ini?` | niat `null`, rentang `null`, gaya `santai` |
| `retribusi hari ini dong` | `retribusi` | `Total retribusi pada 22/10/2026 adalah Rp 208.935 (tarif 3%).` | `retribusi`, 2026-10-22–2026-10-22, `santai` |
| `kalau kemarin?` | `retribusi` | `Total retribusi pada 21/10/2026 adalah Rp 138.375 (tarif 3%).` | `retribusi`, 2026-10-21–2026-10-21, `santai` |
| `makasih bro` | `terima_kasih` | `Sama-sama! Kalau butuh lagi, panggil aja.` | `retribusi`, 2026-10-21–2026-10-21, `santai` |

Percakapan B, login `dinas01`:

| `pesan` | `niat` | `jawaban` | `konteks` di respons |
| --- | --- | --- | --- |
| `Selamat pagi` | `sapaan` | `Selamat datang, dinas01. Silakan tanyakan data lelang, misalnya "rekap bulan ini".` | niat `null`, rentang `null`, gaya `formal` |
| `Siapa kamu?` | `siapa_kamu` | `Nama saya Si Tangkap. Saya asisten berbasis aturan untuk data transaksi lelang TPI, jadi jawaban saya selalu diambil dari data yang tercatat.` | niat `null`, rentang `null`, `formal` |
| `rekap bulan ini` | `rekap` | `Rekap pada 01/10/2026 sampai 22/10/2026: 5 transaksi, total berat 370,75 kg, nilai Rp 11.577.000, retribusi Rp 347.310 (tarif 3%).` | `rekap`, 2026-10-01–2026-10-22, `formal` |
| `yang retribusinya?` | `retribusi` | `Retribusi yang terkumpul pada 01/10/2026 sampai 22/10/2026 sebesar Rp 347.310 (tarif 3%).` | `retribusi`, 2026-10-01–2026-10-22, `formal` |
| `kalau bulan lalu?` | `retribusi` | `Retribusi yang terkumpul pada 01/09/2026 sampai 30/09/2026 sebesar Rp 135.000 (tarif 3%).` | `retribusi`, 2026-09-01–2026-09-30, `formal` |
| `terima kasih` | `terima_kasih` | `Dengan senang hati. Semoga membantu.` | `retribusi`, 2026-09-01–2026-09-30, `formal` |

#### Batas yang disengaja

Si Tangkap bukan model AI. Pertanyaan di luar aturan di atas selalu dijawab dengan templat "tidak dikenali" dan diarahkan kembali ke data lelang. Untuk laporan, uji kewajarannya memakai cek manual F-17 (uji ala Turing yang terukur), bukan klaim lulus Turing test.

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
