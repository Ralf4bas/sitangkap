package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

var semuaKunciConfig = []string{
	"PORT", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME",
	"RETRIBUSI_TARIF", "APP_TIMEZONE", "SESSION_TTL_HOURS", "COOKIE_SECURE", "STATIC_DIR",
}

// kosongkanEnv mengosongkan semua variabel konfigurasi selama satu tes.
func kosongkanEnv(t *testing.T) {
	t.Helper()
	for _, k := range semuaKunciConfig {
		t.Setenv(k, "")
	}
}

func TestLoadDotEnvTidakMenimpaEnvironment(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	isi := "# komentar\n\nSITANGKAP_UJI_A=dari_file\nSITANGKAP_UJI_B=\"pakai kutip\"\nbaris tanpa sama dengan\n"
	if err := os.WriteFile(path, []byte(isi), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SITANGKAP_UJI_A", "dari_env")
	os.Unsetenv("SITANGKAP_UJI_B")
	t.Cleanup(func() { os.Unsetenv("SITANGKAP_UJI_B") })

	if err := LoadDotEnv(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("SITANGKAP_UJI_A"); got != "dari_env" {
		t.Errorf("SITANGKAP_UJI_A = %q, mau dari_env", got)
	}
	if got := os.Getenv("SITANGKAP_UJI_B"); got != "pakai kutip" {
		t.Errorf("SITANGKAP_UJI_B = %q, mau pakai kutip", got)
	}
}

func TestLoadDotEnvFileTidakAda(t *testing.T) {
	if err := LoadDotEnv(filepath.Join(t.TempDir(), "tidak-ada.env")); err != nil {
		t.Fatalf("file tidak ada harus diabaikan, dapat %v", err)
	}
}

func TestLoadConfigDefault(t *testing.T) {
	kosongkanEnv(t)
	t.Setenv("DB_PASSWORD", "rahasia")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "3000" || cfg.DBUser != "sitangkap_app" || cfg.DBName != "sitangkap" {
		t.Errorf("default salah: %+v", cfg)
	}
	if cfg.Tarif != 0.03 || cfg.SessionTTL != 8*time.Hour || cfg.CookieSecure || cfg.StaticDir != "" {
		t.Errorf("default salah: %+v", cfg)
	}
	if cfg.Location.String() != "Asia/Makassar" {
		t.Errorf("zona = %s, mau Asia/Makassar", cfg.Location)
	}
}

func TestLoadConfigMenolakNilaiSalah(t *testing.T) {
	kosongkanEnv(t)
	t.Setenv("RETRIBUSI_TARIF", "1.5")
	t.Setenv("APP_TIMEZONE", "Bukan/Zona")
	t.Setenv("SESSION_TTL_HOURS", "0")
	t.Setenv("COOKIE_SECURE", "mungkin")
	t.Setenv("PORT", "abc")
	_, err := Load()
	if err == nil {
		t.Fatal("harus error")
	}
	for _, kunci := range []string{"PORT", "DB_PASSWORD", "RETRIBUSI_TARIF", "APP_TIMEZONE", "SESSION_TTL_HOURS", "COOKIE_SECURE"} {
		if !strings.Contains(err.Error(), kunci) {
			t.Errorf("pesan error tidak menyebut %s: %v", kunci, err)
		}
	}
}

func TestDSN(t *testing.T) {
	cfg := Config{DBUser: "sitangkap_app", DBPassword: "p@ss:w/rd", DBHost: "127.0.0.1", DBPort: "3306", DBName: "sitangkap"}
	dsn := cfg.DSN()
	// loc=UTC adalah nilai default driver, jadi FormatDSN tidak menuliskannya; dicek lewat ParseDSN.
	mc, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if mc.Loc != time.UTC {
		t.Errorf("Loc = %v, mau UTC", mc.Loc)
	}
	for _, bagian := range []string{
		"sitangkap_app:p@ss:w/rd@tcp(127.0.0.1:3306)/sitangkap?",
		"parseTime=true", "clientFoundRows=true", "charset=utf8mb4", "time_zone=%27%2B00%3A00%27",
	} {
		if !strings.Contains(dsn, bagian) {
			t.Errorf("DSN %q tidak berisi %q", dsn, bagian)
		}
	}
}
