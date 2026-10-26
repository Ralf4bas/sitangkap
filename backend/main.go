package main

import (
	"log"
	"net/http"
	"os"
	"time"
	_ "time/tzdata" // supaya time.LoadLocation("Asia/Makassar") jalan juga di Windows

	"sitangkap/backend/config"
	"sitangkap/backend/handlers"
	"sitangkap/backend/middlewares"
	"sitangkap/backend/repository"
	"sitangkap/backend/routes"
	"sitangkap/backend/services"
)

func main() {
	if err := config.LoadDotEnv(".env"); err != nil {
		log.Printf("gagal membaca .env: %v", err)
		os.Exit(1)
	}
	cfg, err := config.Load()
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	db, err := config.OpenDB(cfg)
	if err != nil {
		log.Printf("gagal terhubung ke database %s:%s/%s: %v", cfg.DBHost, cfg.DBPort, cfg.DBName, err)
		os.Exit(1)
	}
	defer db.Close()

	store := repository.NewMySQLStore(db)
	sessions := services.NewSessions(cfg.SessionTTL, time.Now)
	h := handlers.New(store, sessions, cfg, time.Now)
	auth := middlewares.Auth{Store: store, Sessions: sessions}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           routes.SetupRoutes(h, auth, cfg.StaticDir),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("SiTangkap backend berjalan di http://localhost:%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
