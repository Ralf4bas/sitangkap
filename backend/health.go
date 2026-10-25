package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

type healthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

// handleHealth dipakai Jenkins (smoke test). Ping database gagal dalam 2 detik -> 503.
func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := a.store.Ping(ctx); err != nil {
		log.Printf("health: database tidak terjangkau: %v", err)
		writeJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "error", DB: "down"})
		return
	}
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok", DB: "up"})
}
