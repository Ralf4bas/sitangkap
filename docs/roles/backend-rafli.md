# Panduan Peran: Backend (Rafli)

> **Untuk Rafli dan AI agent yang membantu Rafli.**
> Baca dulu, berurutan: [`../../AGENTS.md`](../../AGENTS.md), [`../00-overview.md`](../00-overview.md), [`../01-api-contract.md`](../01-api-contract.md) (**wajib seluruhnya**), lalu bagian D-05 di [`database-abas.md`](database-abas.md) (query dan DSN).
> Kontrak API adalah hukum. Kalau kode dan kontrak berbeda, kodenya yang salah.

## Ringkasan peran

| Hal | Isi |
| --- | --- |
| Tanggung jawab | API Go: login dan sesi, CRUD transaksi, perhitungan nilai dan retribusi, rekap, health, tes otomatis |
| Folder yang kamu miliki | `backend/` |
| Jangan ubah | `frontend/`, `db/`, `docs/` (minta lewat issue) |
| Kamu menerima dari | **Database (abas):** `db/schema.sql`, `db/seed.sql`, `db/queries.md` (query Q-01 s.d. Q-09, DSN). **PM:** tarif retribusi |
| Kamu memberi ke | **Frontend (Magen):** API yang berjalan di `localhost:3000` sesuai kontrak. **PM/Jenkins:** `go vet` dan `go test` yang lulus, binary yang bisa di-build |
| Tenggat | Login, `/api/me`, dan `GET /api/transaksi` berjalan dengan database sungguhan **di akhir hari pertama** (checkpoint CP-2) |

## Aturan teknis yang tidak boleh dilanggar

1. Go 1.26 atau lebih baru (`golang.org/x/crypto` terbaru membutuhkan go 1.26). Router memakai `net/http` standar dengan pola `"GET /api/transaksi/{id}"`. **Tanpa framework** (Gin, Echo, Fiber, Chi, dll.).
2. Dependensi hanya dua: `github.com/go-sql-driver/mysql` dan `golang.org/x/crypto/bcrypt`. Jangan menambah library lain (termasuk `godotenv`, `sqlx`, `gorm`, `testify`).
3. Semua query memakai `?` (NFR-002) dan diambil dari `db/queries.md`.
4. Semua angka uang dihitung di satu tempat (`hitung.go`) dengan aritmetika integer (lihat B-05).
5. Pesan error ke klien hanya yang tertulis di kontrak. Error asli dicatat ke log, tidak pernah dikirim ke klien (NFR-008).
6. Password tidak pernah dicatat di log dan tidak pernah dikirim di respons.

## Struktur folder

Satu paket `main`, datar. Pembagian file mengikuti pola controller, service, dan repository dari materi Class Diagram, supaya mudah dijelaskan di laporan.

```text
backend/
  go.mod                 module sitangkap/backend, go 1.26.0
  go.sum
  .env.example           contoh konfigurasi (di-commit)
  .env                   konfigurasi lokal (TIDAK di-commit)
  main.go                baca config, buka DB, rakit App, daftarkan rute, jalankan server
  config.go              baca .env + environment, validasi nilai
  respond.go             helper JSON: writeJSON, writeError, decodeBody
  store.go               interface Store + tipe data (repository: kontrak)
  store_mysql.go         implementasi Store dengan MySQL (repository)
  auth.go                sesi, cookie, middleware withAuth, handler login/logout/me (controller)
  transaksi.go           handler list/create/update/delete (controller)
  rekap.go               handler rekap (controller)
  hitung.go              hitungNilai, hitungRetribusi, rekap (service: business rule BR-05)
  validasi.go            parse + validasi body transaksi dan query tanggal (service: BR-02..04, BR-07, BR-09)
  health.go              handler /health
  memstore_test.go       Store palsu di memori, khusus tes
  *_test.go              tes T-01 s.d. T-16
```

| Lapisan (materi SAD) | File | Tanggung jawab |
| --- | --- | --- |
| Controller | `auth.go`, `transaksi.go`, `rekap.go`, `health.go` | Terima request, cek sesi dan peran, panggil service dan repository, tulis respons |
| Service | `hitung.go`, `validasi.go` | Aturan bisnis, tanpa HTTP dan tanpa SQL |
| Repository | `store.go`, `store_mysql.go` | Akses database |

## B-01 · Inisialisasi modul

```bash
cd backend
go mod init sitangkap/backend
go get github.com/go-sql-driver/mysql golang.org/x/crypto/bcrypt
```

`backend/.env.example` (di-commit):

```dotenv
PORT=3000
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=sitangkap_app
DB_PASSWORD=sitangkap_dev
DB_NAME=sitangkap
RETRIBUSI_TARIF=0.03
APP_TIMEZONE=Asia/Makassar
SESSION_TTL_HOURS=8
COOKIE_SECURE=false
STATIC_DIR=
```

Selesai bila: `go build ./...` berhasil dengan `main.go` kosong yang hanya mencetak "ok".

## B-02 · `config.go`

- Muat `.env` dari direktori kerja **bila ada**, dengan pembaca kecil buatan sendiri (tanpa library). Variabel yang sudah ada di environment **tidak ditimpa**, supaya systemd di server bisa memakai cara yang sama.
- Validasi: `DB_PASSWORD` wajib ada; `RETRIBUSI_TARIF` harus 0 < tarif < 1; `APP_TIMEZONE` harus bisa dimuat; `SESSION_TTL_HOURS` bilangan bulat > 0. Kalau salah, cetak pesan jelas lalu `os.Exit(1)`.
- Tambahkan `import _ "time/tzdata"` di `main.go` supaya `time.LoadLocation("Asia/Makassar")` jalan juga di Windows.

```go
// Pembaca .env minimal. Format: KEY=VALUE, baris kosong dan '#' diabaikan.
func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		if _, exists := os.LookupEnv(k); !exists {
			os.Setenv(k, v)
		}
	}
	return sc.Err()
}
```

DSN: susun dengan `mysql.Config` (dari dependensi yang sama) supaya password berisi karakter khusus tetap aman. Hasilnya setara dengan DSN di `db/queries.md`:

```go
mc := mysql.NewConfig()
mc.User = cfg.DBUser
mc.Passwd = cfg.DBPassword
mc.Net = "tcp"
mc.Addr = net.JoinHostPort(cfg.DBHost, cfg.DBPort)
mc.DBName = cfg.DBName
mc.ParseTime = true
mc.Loc = time.UTC
mc.ClientFoundRows = true
mc.Params = map[string]string{"charset": "utf8mb4", "time_zone": "'+00:00'"}
db, err := sql.Open("mysql", mc.FormatDSN())
```

Setelah `sql.Open`: `db.SetMaxOpenConns(10)`, `db.SetConnMaxLifetime(5 * time.Minute)`, lalu `db.PingContext` dengan timeout 5 detik. Kalau gagal, log error dan keluar.

## B-03 · `respond.go`

Tiga helper yang dipakai semua handler:

```go
func writeJSON(w http.ResponseWriter, status int, body any)        // set Content-Type, encode JSON
func writeError(w http.ResponseWriter, status int, code, message string, fields map[string]string)
func decodeBody(w http.ResponseWriter, r *http.Request, dst any) bool // MaxBytesReader 10 KB; gagal -> 400 BAD_REQUEST, return false
```

- `writeError` menghasilkan persis `{"error":{"code":...,"message":...,"fields":...}}`; `fields` dihilangkan bila kosong (`omitempty`).
- Siapkan konstanta pesan default dari tabel error di kontrak supaya tidak salah ketik.
- Slice kosong harus ter-encode sebagai `[]`, bukan `null`: inisialisasi dengan `make([]T, 0)`.

## B-04 · `store.go` dan `store_mysql.go`

```go
var ErrNotFound = errors.New("not found")

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nama     string `json:"nama"`
	Role     string `json:"role"`
}

type UserWithHash struct {
	User
	PasswordHash string
	Aktif        bool
}

// TransaksiInput adalah input yang sudah lolos validasi.
type TransaksiInput struct {
	NamaNelayan string
	JenisIkan   string
	BeratCenti  int64  // berat dalam 0,01 kg: 120.5 kg -> 12050
	HargaPerKg  int64
	Tanggal     string // "YYYY-MM-DD"
}

// TransaksiRow adalah satu baris dari database (Q-03/Q-04/Q-05).
type TransaksiRow struct {
	ID          int64
	NamaNelayan string
	JenisIkan   string
	BeratCenti  int64
	HargaPerKg  int64
	Tanggal     time.Time // DATE, dibaca UTC tengah malam
	DibuatOleh  string    // users.username
	DibuatPada  time.Time // UTC
}

type Store interface {
	Ping(ctx context.Context) error                                              // Q-09
	FindUserByUsername(ctx context.Context, username string) (UserWithHash, error) // Q-01, ErrNotFound
	FindActiveUserByID(ctx context.Context, id int64) (User, error)             // Q-02, ErrNotFound
	ListTransaksi(ctx context.Context, dari, sampai string) ([]TransaksiRow, error) // Q-03 bila dari=="" , Q-04 bila terisi
	GetTransaksi(ctx context.Context, id int64) (TransaksiRow, error)            // Q-05, ErrNotFound
	CreateTransaksi(ctx context.Context, in TransaksiInput, userID int64) (int64, error) // Q-06, id baru
	UpdateTransaksi(ctx context.Context, id int64, in TransaksiInput) error      // Q-07, ErrNotFound bila 0 baris cocok
	DeleteTransaksi(ctx context.Context, id int64) error                         // Q-08, ErrNotFound bila 0 baris
}
```

Catatan implementasi MySQL:
- `berat_kg` (`DECIMAL`) dibaca ke `string`, lalu dikonversi ke centi dengan `int64(math.Round(f * 100))` setelah `strconv.ParseFloat`. Saat menulis, kirim `fmt.Sprintf("%d.%02d", c/100, c%100)`.
- `tanggal` dibaca sebagai `time.Time` dan diformat dengan `t.Format("2006-01-02")` **tanpa** `.In(loc)`.
- `dibuat_pada` dibaca sebagai `time.Time` UTC dan ditampilkan dengan `t.In(loc).Format(time.RFC3339)`.
- Gunakan `context` dari request (`r.Context()`) di setiap query.

## B-05 · `hitung.go` (BR-05)

Pakai integer supaya hasilnya persis sama dengan tabel contoh di kontrak, tanpa kesalahan pecahan desimal.

```go
// hitungNilai = round(berat_kg × harga_per_kg) dalam rupiah.
// beratCenti dan harga selalu positif, jadi (x + 50) / 100 adalah pembulatan .5 ke atas.
func hitungNilai(beratCenti, hargaPerKg int64) int64 {
	return (beratCenti*hargaPerKg + 50) / 100
}

// hitungRetribusi = round(nilai × tarif). math.Round membulatkan .5 menjauhi nol.
func hitungRetribusi(nilai int64, tarif float64) int64 {
	return int64(math.Round(float64(nilai) * tarif))
}
```

Rekap: jumlahkan `nilai` dan `retribusi` **per baris** yang sudah dibulatkan; `total_berat_kg = float64(ΣberatCenti) / 100`.

Tipe respons JSON:

```go
type TransaksiJSON struct {
	ID          int64   `json:"id"`
	NamaNelayan string  `json:"nama_nelayan"`
	JenisIkan   string  `json:"jenis_ikan"`
	BeratKg     float64 `json:"berat_kg"`
	HargaPerKg  int64   `json:"harga_per_kg"`
	Tanggal     string  `json:"tanggal"`
	Nilai       int64   `json:"nilai"`
	Retribusi   int64   `json:"retribusi"`
	DibuatOleh  string  `json:"dibuat_oleh"`
	DibuatPada  string  `json:"dibuat_pada"`
}

type RekapJSON struct {
	Dari            string  `json:"dari"`
	Sampai          string  `json:"sampai"`
	JumlahTransaksi int     `json:"jumlah_transaksi"`
	TotalBeratKg    float64 `json:"total_berat_kg"`
	TotalNilai      int64   `json:"total_nilai"`
	TotalRetribusi  int64   `json:"total_retribusi"`
	TarifRetribusi  float64 `json:"tarif_retribusi"`
}
```

Selesai bila: tes T-12 (tabel contoh di kontrak) lulus.

## B-06 · `validasi.go`

Supaya bisa mengembalikan **semua** field yang salah sekaligus dengan pesan yang tepat (bukan 400 karena gagal decode), decode body ke `json.RawMessage` per field, lalu periksa satu per satu:

```go
type transaksiBody struct {
	NamaNelayan json.RawMessage `json:"nama_nelayan"`
	JenisIkan   json.RawMessage `json:"jenis_ikan"`
	BeratKg     json.RawMessage `json:"berat_kg"`
	HargaPerKg  json.RawMessage `json:"harga_per_kg"`
	Tanggal     json.RawMessage `json:"tanggal"`
}

func isMissing(raw json.RawMessage) bool { return len(raw) == 0 || string(raw) == "null" }

// parseTransaksi mengembalikan input bersih, atau map field -> pesan (pesan persis dari kontrak).
func parseTransaksi(b transaksiBody, today time.Time, loc *time.Location) (TransaksiInput, map[string]string)
```

Aturan (pesan persis dari tabel validasi di kontrak):

| Field | Cara periksa |
| --- | --- |
| `nama_nelayan`, `jenis_ikan` | `json.Unmarshal` ke `string`; gagal atau kosong setelah `strings.TrimSpace` → "wajib diisi". Panjang dihitung dengan `utf8.RuneCountInString` |
| `berat_kg` | `json.Unmarshal` ke `float64` (string `"120.5"` ditolak); ≤ 0; > 100000; dua desimal: `math.Abs(f*100 - math.Round(f*100)) > 1e-6` berarti lebih dari 2 desimal |
| `harga_per_kg` | `json.Unmarshal` ke `int64` (gagal untuk `28000.5`, string, atau `2.8e4`) → pesan bilangan bulat; ≤ 0; > 100000000 |
| `tanggal` | `json.Unmarshal` ke `string`, lalu `time.ParseInLocation("2006-01-02", s, loc)`; tanggal tidak nyata (misal `2026-02-30`) otomatis gagal; setelah `today` (tanggal hari ini di `loc`) → pesan "tidak boleh melewati hari ini" |

Query tanggal (`dari`, `sampai`): keduanya kosong, atau keduanya terisi dan valid dengan `dari ≤ sampai`. Selain itu `400 BAD_REQUEST`.

"Hari ini" diambil dari `App.now()` (lihat B-08) supaya bisa diuji.

## B-07 · `auth.go`

Sesi di memori:

```go
type session struct {
	userID  int64
	role    string
	expires time.Time
}

type sessions struct {
	mu  sync.Mutex
	m   map[string]session
	ttl time.Duration
}
// create(userID, role) (token string)   — token: 32 byte crypto/rand, hex.EncodeToString
// get(token) (session, bool)            — hapus dan kembalikan false bila kedaluwarsa
// delete(token)
```

Login (`POST /api/login`):
1. `decodeBody` ke `{username, password string}`; kosong → 422 dengan pesan dari kontrak.
2. `FindUserByUsername`. Tidak ada, `aktif` false, atau `bcrypt.CompareHashAndPassword` gagal → **satu** respons yang sama: 401 `INVALID_CREDENTIALS`. Bila user tidak ditemukan, tetap jalankan `bcrypt.CompareHashAndPassword` terhadap hash palsu supaya waktu respons tidak membocorkan username mana yang ada. Hash palsu itu harus hash bcrypt **cost 12 yang valid**: buat sekali saat start dengan `bcrypt.GenerateFromPassword([]byte("dummy"), 12)` (hash rusak akan langsung gagal tanpa menghabiskan waktu).
3. Buat sesi, set cookie persis seperti di kontrak (`HttpOnly`, `SameSite=Lax`, `Path=/`, `Max-Age` = TTL detik, `Secure` bila `COOKIE_SECURE=true`).
4. `200 {"data": User}`.

Middleware:

```go
// withAuth membungkus handler yang butuh login. roles kosong = semua peran boleh.
func (a *App) withAuth(roles []string, h func(w http.ResponseWriter, r *http.Request, u User)) http.HandlerFunc
```

Urutan: cookie tidak ada atau sesi tidak dikenal → 401; `FindActiveUserByID` gagal → hapus sesi, 401; peran tidak termasuk `roles` → 403; selain itu panggil `h`.

## B-08 · Rute dan `main.go`

```go
type App struct {
	store    Store
	sessions *sessions
	cfg      Config
	loc      *time.Location
	now      func() time.Time // time.Now di production, waktu tetap di tes
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	petugas := []string{"petugas"}
	dinas := []string{"dinas"}
	semua := []string{}

	mux.HandleFunc("GET /health", a.handleHealth)
	mux.HandleFunc("POST /api/login", a.handleLogin)
	mux.HandleFunc("POST /api/logout", a.handleLogout)
	mux.HandleFunc("GET /api/me", a.withAuth(semua, a.handleMe))
	mux.HandleFunc("GET /api/transaksi", a.withAuth(semua, a.handleListTransaksi))
	mux.HandleFunc("POST /api/transaksi", a.withAuth(petugas, a.handleCreateTransaksi))
	mux.HandleFunc("PUT /api/transaksi/{id}", a.withAuth(petugas, a.handleUpdateTransaksi))
	mux.HandleFunc("DELETE /api/transaksi/{id}", a.withAuth(petugas, a.handleDeleteTransaksi))
	mux.HandleFunc("GET /api/rekap", a.withAuth(dinas, a.handleRekap))
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Data tidak ditemukan.", nil)
	})
	if a.cfg.StaticDir != "" {
		mux.Handle("/", http.FileServer(http.Dir(a.cfg.StaticDir)))
	}
	return a.logRequests(a.recoverPanic(mux))
}
```

- `id` diambil dengan `r.PathValue("id")` lalu `strconv.ParseInt`; bukan bilangan bulat positif → 400.
- `logRequests`: satu baris per request: method, path, status, durasi. Jangan mencatat body.
- `recoverPanic`: panic → log + 500 `INTERNAL_ERROR`.
- Method yang tidak terdaftar (misal `PATCH /api/transaksi/1`) jatuh ke penampung `"/api/"` dan dijawab 404 `NOT_FOUND`. Ini disengaja dan sudah diuji; jangan membuat handler 405 sendiri.
- `http.Server` dengan `ReadHeaderTimeout: 5 * time.Second`. Port dari `PORT`.

## B-09 · Handler transaksi, rekap, dan health

| Handler | Langkah |
| --- | --- |
| list | parse `dari`/`sampai` → `ListTransaksi` → ubah ke `[]TransaksiJSON` (hitung `nilai`, `retribusi`) → `200` dengan `meta`: `jumlah`, `dari`/`sampai` (`null` bila tidak dipakai), `tarif_retribusi` |
| create | `decodeBody` → `parseTransaksi` → 422 bila ada error → `CreateTransaksi` dengan `u.ID` → `GetTransaksi(idBaru)` → `201` |
| update | parse id → `decodeBody` → `parseTransaksi` → `UpdateTransaksi` (`ErrNotFound` → 404) → `GetTransaksi` → `200` |
| delete | parse id → `DeleteTransaksi` (`ErrNotFound` → 404) → `204` tanpa body |
| rekap | rentang default bulan ini bila keduanya kosong → `ListTransaksi` → jumlahkan → `200 {"data": RekapJSON}` |
| health | `Ping` dengan timeout 2 detik → `200` atau `503` sesuai kontrak |

Urutan cek di update sesuai kontrak: 401 → 403 → 400 (id bukan angka atau JSON rusak) → 422 (validasi) → 404 (id tidak ada).

## B-10 · Tes otomatis

Tes **tidak** butuh MySQL. Buat `memStore` di `memstore_test.go` yang memenuhi interface `Store`, lalu uji lewat `httptest`. Rakit `App` dengan `now` tetap, misalnya `2026-10-22 10:00 WITA`, dan tarif `0.03`. Simpan hash bcrypt asli dari `db/seed.sql` untuk akun demo di `memStore`.

| ID | Skenario | Harapan |
| --- | --- | --- |
| T-01 | Login `petugas01`/`petugas123` | 200, body `data.role = "petugas"`, header `Set-Cookie` berisi `HttpOnly` dan `SameSite=Lax` |
| T-02 | Login password salah, dan username tidak ada | Keduanya 401 `INVALID_CREDENTIALS` dengan pesan sama |
| T-03 | `GET /api/me` tanpa cookie | 401 `UNAUTHORIZED` |
| T-04 | Logout lalu `GET /api/me` dengan cookie lama | Logout 204 dengan `Max-Age=0`; `/api/me` 401 |
| T-05 | `GET /api/transaksi` pada data kosong dan berisi | Kosong → `"data":[]`; berisi → urut tanggal terbaru, lalu id terbesar; `meta.tarif_retribusi = 0.03` |
| T-06 | `POST` body valid (Pak Roni, 120.5, 28000) | 201, `nilai = 3374000`, `retribusi = 101220`, `dibuat_oleh = "petugas01"` |
| T-07 | `POST` berat `-1`, tanggal `2026-10-23`, nama kosong | 422 dengan **tiga** field di `fields`, pesan persis kontrak |
| T-08 | `POST` harga `28000.5`, berat `"120.5"` (string), berat `1.234` | 422 dengan pesan yang sesuai, **bukan** 400 |
| T-09 | `PUT` id ada, body valid | 200, data berubah, `dibuat_oleh` tetap |
| T-10 | `PUT`/`DELETE` id `9999` dan id `abc` | 404 dan 400 |
| T-11 | `DELETE` id ada, lalu `PUT` id itu | 204 lalu 404 |
| T-12 | Tabel contoh perhitungan di kontrak | Kelima baris dan total cocok persis |
| T-13 | `GET /api/rekap?dari=2026-10-01&sampai=2026-10-22` dengan 5 data contoh; tanpa parameter | `5 / 370.75 / 11577000 / 347310`; tanpa parameter → `dari = 2026-10-01`, `sampai = 2026-10-22` |
| T-14 | Query `dari` saja; `dari=2026-10-22&sampai=2026-10-01`; `dari=22-10-2026` | Semuanya 400 |
| T-15 | `/health` dengan store sehat dan store yang `Ping`-nya gagal | 200 dan 503 |
| T-16 | Dinas `POST /api/transaksi`; petugas `GET /api/rekap` | 403 keduanya |

```bash
cd backend
go vet ./...
go test ./...
```

## B-11 · Uji integrasi dengan MySQL sungguhan

1. Database siap sesuai `db/README.md` (dari abas), seed sudah dijalankan.
2. `cd backend && go run .`
3. Jalankan semua perintah curl di bagian 7 kontrak API. Respons harus sama bentuknya dengan kontrak.
4. Simpan transaksi dengan data yang **sama persis** dua kali lewat `PUT`; keduanya harus 200 (membuktikan `clientFoundRows=true` terpasang).

## Koordinasi

| Kapan | Dengan siapa | Apa |
| --- | --- | --- |
| Awal hari 1 | Database | Pastikan `db/queries.md` dan DSN sudah ada; sambil menunggu, kerjakan B-01 s.d. B-03, B-05, B-06 dengan tes |
| Setelah login, `/api/me`, `GET /api/transaksi` jalan (CP-2) | Frontend | Kabari Magen: dia boleh mematikan mode mock (`PUBLIC_MOCK=0`) untuk halaman itu |
| Setiap ada kebingungan kontrak | PM | Issue berlabel `kontrak`. Jangan mengubah bentuk respons sendiri |
| Sebelum fase 2 | PM | Pastikan `go build -o sitangkap .` di folder `backend/` menghasilkan satu binary, dan backend bisa melayani `STATIC_DIR` |

## Checklist selesai

- [ ] Hanya dua dependensi di `go.mod`
- [ ] `go vet ./...` bersih, `go test ./...` lulus T-01 s.d. T-16
- [ ] Semua respons dan error cocok dengan kontrak (dicek dengan curl)
- [ ] Tidak ada `fmt.Sprintf` untuk menyusun SQL; semua memakai `?`
- [ ] `.env` tidak ter-commit; `.env.example` ter-commit
- [ ] Log tidak berisi password atau isi body
- [ ] `STATIC_DIR` diisi `../frontend/dist` → `http://localhost:3000/` menampilkan halaman login hasil build

## Prompt siap pakai untuk AI agent

Kerjakan per tahap (satu sesi agent per tahap B-xx), jangan semuanya sekaligus.

```text
Kamu membantu peran BACKEND di proyek SiTangkap (Go, tanpa framework).
Baca dulu, berurutan: AGENTS.md, docs/00-overview.md, docs/01-api-contract.md, docs/roles/backend-rafli.md, dan bagian D-05 di docs/roles/database-abas.md.
Tugas sesi ini: kerjakan tahap <B-xx> di docs/roles/backend-rafli.md.
Aturan:
- Hanya ubah file di folder backend/.
- Ikuti docs/01-api-contract.md persis: nama field, kode status, kode dan pesan error.
- Dependensi hanya github.com/go-sql-driver/mysql dan golang.org/x/crypto/bcrypt.
- Go 1.26+, router net/http standar dengan pola "METHOD /path/{id}".
- Semua SQL dari db/queries.md dengan placeholder ?.
- Tulis atau perbarui tes yang relevan, lalu jalankan `go vet ./...` dan `go test ./...` dan tunjukkan hasilnya.
Kalau kontrak tidak menjawab sesuatu, berhenti dan tanyakan; jangan menebak atau menambah endpoint.
```
