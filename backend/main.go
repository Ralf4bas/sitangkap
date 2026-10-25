package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"
	_ "time/tzdata" // supaya time.LoadLocation("Asia/Makassar") jalan juga di Windows
)

type App struct {
	store    Store
	sessions *sessions
	cfg      Config
	loc      *time.Location
	now      func() time.Time // time.Now di production, waktu tetap di tes
}

func newApp(store Store, cfg Config, now func() time.Time) *App {
	return &App{
		store:    store,
		sessions: newSessions(cfg.SessionTTL, now),
		cfg:      cfg,
		loc:      cfg.Location,
		now:      now,
	}
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
	// Path /api/ lain, termasuk method yang tidak terdaftar (misal PATCH), dijawab 404 NOT_FOUND.
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		notFound(w)
	})
	if a.cfg.StaticDir != "" {
		mux.Handle("/", http.FileServer(fsTanpaDaftar{http.Dir(a.cfg.StaticDir)}))
	}
	return a.logRequests(a.recoverPanic(mux))
}

func openDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func main() {
	if err := loadDotEnv(".env"); err != nil {
		log.Printf("gagal membaca .env: %v", err)
		os.Exit(1)
	}
	cfg, err := loadConfig()
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	db, err := openDB(cfg)
	if err != nil {
		log.Printf("gagal terhubung ke database %s:%s/%s: %v", cfg.DBHost, cfg.DBPort, cfg.DBName, err)
		os.Exit(1)
	}
	defer db.Close()

	app := newApp(&mysqlStore{db: db}, cfg, time.Now)
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           app.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("SiTangkap backend berjalan di http://localhost:%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
