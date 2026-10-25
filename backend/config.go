package main

import (
	"bufio"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Config berisi seluruh konfigurasi dari environment (kontrak API bagian 5).
type Config struct {
	Port         string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	Tarif        float64
	Location     *time.Location
	SessionTTL   time.Duration
	CookieSecure bool
	StaticDir    string
}

// loadDotEnv adalah pembaca .env minimal. Format: KEY=VALUE, baris kosong dan '#' diabaikan.
// Variabel yang sudah ada di environment tidak ditimpa, supaya systemd bisa memakai cara yang sama.
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

func getenv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

// loadConfig membaca dan memvalidasi konfigurasi. Semua kesalahan dilaporkan sekaligus.
func loadConfig() (Config, error) {
	cfg := Config{
		Port:       getenv("PORT", "3000"),
		DBHost:     getenv("DB_HOST", "127.0.0.1"),
		DBPort:     getenv("DB_PORT", "3306"),
		DBUser:     getenv("DB_USER", "sitangkap_app"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     getenv("DB_NAME", "sitangkap"),
		StaticDir:  getenv("STATIC_DIR", ""),
	}
	var salah []string

	if p, err := strconv.Atoi(cfg.Port); err != nil || p < 1 || p > 65535 {
		salah = append(salah, "PORT harus bilangan bulat 1-65535")
	}
	if cfg.DBPassword == "" {
		salah = append(salah, "DB_PASSWORD wajib diisi")
	}

	tarif, err := strconv.ParseFloat(getenv("RETRIBUSI_TARIF", "0.03"), 64)
	if err != nil || !(tarif > 0 && tarif < 1) {
		salah = append(salah, "RETRIBUSI_TARIF harus desimal lebih dari 0 dan kurang dari 1 (contoh 0.03)")
	}
	cfg.Tarif = tarif

	loc, err := time.LoadLocation(getenv("APP_TIMEZONE", "Asia/Makassar"))
	if err != nil {
		salah = append(salah, "APP_TIMEZONE tidak dikenal (contoh Asia/Makassar)")
	}
	cfg.Location = loc

	jam, err := strconv.Atoi(getenv("SESSION_TTL_HOURS", "8"))
	if err != nil || jam <= 0 {
		salah = append(salah, "SESSION_TTL_HOURS harus bilangan bulat lebih dari 0")
	}
	cfg.SessionTTL = time.Duration(jam) * time.Hour

	secure, err := strconv.ParseBool(getenv("COOKIE_SECURE", "false"))
	if err != nil {
		salah = append(salah, "COOKIE_SECURE harus true atau false")
	}
	cfg.CookieSecure = secure

	if len(salah) > 0 {
		return cfg, errors.New("konfigurasi tidak valid: " + strings.Join(salah, "; "))
	}
	return cfg, nil
}

// DSN menyusun alamat database lewat mysql.Config supaya password berisi karakter khusus tetap aman.
// Hasilnya setara dengan DSN di db/queries.md. Jangan pernah mencatat nilai ini ke log.
func (c Config) DSN() string {
	mc := mysql.NewConfig()
	mc.User = c.DBUser
	mc.Passwd = c.DBPassword
	mc.Net = "tcp"
	mc.Addr = net.JoinHostPort(c.DBHost, c.DBPort)
	mc.DBName = c.DBName
	mc.ParseTime = true
	mc.Loc = time.UTC
	mc.ClientFoundRows = true
	mc.Params = map[string]string{"charset": "utf8mb4", "time_zone": "'+00:00'"}
	return mc.FormatDSN()
}
