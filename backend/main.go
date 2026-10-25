package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"
	_ "time/tzdata" // supaya time.LoadLocation("Asia/Makassar") jalan juga di Windows
)

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
	log.Printf("database %s terhubung", cfg.DBName)
}
