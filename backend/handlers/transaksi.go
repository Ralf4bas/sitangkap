package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"sitangkap/backend/helpers"
	"sitangkap/backend/models"
	"sitangkap/backend/repository"
	"sitangkap/backend/services"
)

type listMeta struct {
	Jumlah         int     `json:"jumlah"`
	Dari           *string `json:"dari"`
	Sampai         *string `json:"sampai"`
	TarifRetribusi float64 `json:"tarif_retribusi"`
}

type listResponse struct {
	Data []models.TransaksiJSON `json:"data"`
	Meta listMeta               `json:"meta"`
}

// parseID mengambil {id} dari path. Bukan bilangan bulat positif -> ok = false (400).
func parseID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id, err == nil && id > 0
}

func nilPadaKosong(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ListTransaksiHandler: GET /api/transaksi.
func (h *Handler) ListTransaksiHandler(w http.ResponseWriter, r *http.Request, u models.User) {
	dari, sampai, ok := services.ParseRentang(r.URL.Query())
	if !ok {
		helpers.BadRequest(w)
		return
	}
	rows, err := h.Store.ListTransaksi(r.Context(), dari, sampai)
	if err != nil {
		helpers.InternalError(w, r, err)
		return
	}
	data := make([]models.TransaksiJSON, 0, len(rows))
	for _, row := range rows {
		data = append(data, services.KeTransaksiJSON(row, h.Cfg.Tarif, h.Cfg.Location))
	}
	helpers.WriteJSON(w, http.StatusOK, listResponse{
		Data: data,
		Meta: listMeta{Jumlah: len(data), Dari: nilPadaKosong(dari), Sampai: nilPadaKosong(sampai), TarifRetribusi: h.Cfg.Tarif},
	})
}

// kirimTransaksi membaca ulang transaksi dari database (lengkap dengan dibuat_oleh dan dibuat_pada).
func (h *Handler) kirimTransaksi(w http.ResponseWriter, r *http.Request, status int, id int64) {
	row, err := h.Store.GetTransaksi(r.Context(), id)
	if err != nil {
		helpers.InternalError(w, r, err)
		return
	}
	helpers.WriteData(w, status, services.KeTransaksiJSON(row, h.Cfg.Tarif, h.Cfg.Location))
}

// CreateTransaksiHandler: POST /api/transaksi.
func (h *Handler) CreateTransaksiHandler(w http.ResponseWriter, r *http.Request, u models.User) {
	var b services.TransaksiBody
	if !helpers.DecodeBody(w, r, &b) {
		return
	}
	in, fields := services.ParseTransaksi(b, h.today(), h.Cfg.Location)
	if fields != nil {
		helpers.ValidationError(w, fields)
		return
	}
	id, err := h.Store.CreateTransaksi(r.Context(), in, u.ID)
	if err != nil {
		helpers.InternalError(w, r, err)
		return
	}
	h.kirimTransaksi(w, r, http.StatusCreated, id)
}

// UpdateTransaksiHandler: PUT /api/transaksi/{id}.
// Urutan cek sesuai kontrak: 401 -> 403 (middleware) -> 400 (id, JSON) -> 422 -> 404.
func (h *Handler) UpdateTransaksiHandler(w http.ResponseWriter, r *http.Request, u models.User) {
	id, ok := parseID(r)
	if !ok {
		helpers.BadRequest(w)
		return
	}
	var b services.TransaksiBody
	if !helpers.DecodeBody(w, r, &b) {
		return
	}
	in, fields := services.ParseTransaksi(b, h.today(), h.Cfg.Location)
	if fields != nil {
		helpers.ValidationError(w, fields)
		return
	}
	err := h.Store.UpdateTransaksi(r.Context(), id, in)
	if errors.Is(err, repository.ErrNotFound) {
		helpers.NotFound(w)
		return
	}
	if err != nil {
		helpers.InternalError(w, r, err)
		return
	}
	h.kirimTransaksi(w, r, http.StatusOK, id)
}

// DeleteTransaksiHandler: DELETE /api/transaksi/{id}.
func (h *Handler) DeleteTransaksiHandler(w http.ResponseWriter, r *http.Request, u models.User) {
	id, ok := parseID(r)
	if !ok {
		helpers.BadRequest(w)
		return
	}
	err := h.Store.DeleteTransaksi(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		helpers.NotFound(w)
		return
	}
	if err != nil {
		helpers.InternalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
