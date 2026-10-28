package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"sitangkap/backend/helpers"
)

type healthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

// HealthHandler: GET /health, dipakai Jenkins (smoke test). Ping database gagal dalam 2 detik -> 503.
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.Store.Ping(ctx); err != nil {
		log.Printf("health: database tidak terjangkau: %v", err)
		helpers.WriteJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "error", DB: "down"})
		return
	}
	helpers.WriteJSON(w, http.StatusOK, healthResponse{Status: "ok", DB: "up"})
}
