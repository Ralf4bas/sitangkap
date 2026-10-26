// Package routes mendaftarkan semua endpoint dari kontrak API bagian 4.
package routes

import (
	"net/http"

	"sitangkap/backend/handlers"
	"sitangkap/backend/helpers"
	"sitangkap/backend/middlewares"
)

// SetupRoutes merakit router. staticDir kosong = tanpa file statis (development).
func SetupRoutes(h *handlers.Handler, auth middlewares.Auth, staticDir string) http.Handler {
	mux := http.NewServeMux()
	petugas := []string{"petugas"}
	dinas := []string{"dinas"}
	semua := []string{}

	mux.HandleFunc("GET /health", h.HealthHandler)
	mux.HandleFunc("POST /api/login", h.LoginHandler)
	mux.HandleFunc("POST /api/logout", h.LogoutHandler)
	mux.HandleFunc("GET /api/me", auth.Require(semua, h.MeHandler))
	mux.HandleFunc("GET /api/transaksi", auth.Require(semua, h.ListTransaksiHandler))
	mux.HandleFunc("POST /api/transaksi", auth.Require(petugas, h.CreateTransaksiHandler))
	mux.HandleFunc("PUT /api/transaksi/{id}", auth.Require(petugas, h.UpdateTransaksiHandler))
	mux.HandleFunc("DELETE /api/transaksi/{id}", auth.Require(petugas, h.DeleteTransaksiHandler))
	mux.HandleFunc("GET /api/rekap", auth.Require(dinas, h.RekapHandler))
	mux.HandleFunc("POST /api/chat", auth.Require(semua, h.ChatHandler))
	// Path /api/ lain, termasuk method yang tidak terdaftar (misal PATCH), dijawab 404 NOT_FOUND.
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		helpers.NotFound(w)
	})
	if staticDir != "" {
		mux.Handle("/", http.FileServer(fsTanpaDaftar{http.Dir(staticDir)}))
	}
	return middlewares.LogRequests(middlewares.RecoverPanic(mux))
}
