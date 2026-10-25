package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"
)

type listMeta struct {
	Jumlah         int     `json:"jumlah"`
	Dari           *string `json:"dari"`
	Sampai         *string `json:"sampai"`
	TarifRetribusi float64 `json:"tarif_retribusi"`
}

type listResponse struct {
	Data []TransaksiJSON `json:"data"`
	Meta listMeta        `json:"meta"`
}

// today adalah "hari ini" di zona APP_TIMEZONE.
func (a *App) today() time.Time {
	return a.now().In(a.loc)
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

func (a *App) handleListTransaksi(w http.ResponseWriter, r *http.Request, u User) {
	dari, sampai, ok := parseRentang(r.URL.Query())
	if !ok {
		badRequest(w)
		return
	}
	rows, err := a.store.ListTransaksi(r.Context(), dari, sampai)
	if err != nil {
		internalError(w, r, err)
		return
	}
	data := make([]TransaksiJSON, 0, len(rows))
	for _, row := range rows {
		data = append(data, keTransaksiJSON(row, a.cfg.Tarif, a.loc))
	}
	writeJSON(w, http.StatusOK, listResponse{
		Data: data,
		Meta: listMeta{Jumlah: len(data), Dari: nilPadaKosong(dari), Sampai: nilPadaKosong(sampai), TarifRetribusi: a.cfg.Tarif},
	})
}

// kirimTransaksi membaca ulang transaksi dari database (lengkap dengan dibuat_oleh dan dibuat_pada).
func (a *App) kirimTransaksi(w http.ResponseWriter, r *http.Request, status int, id int64) {
	row, err := a.store.GetTransaksi(r.Context(), id)
	if err != nil {
		internalError(w, r, err)
		return
	}
	writeData(w, status, keTransaksiJSON(row, a.cfg.Tarif, a.loc))
}

func (a *App) handleCreateTransaksi(w http.ResponseWriter, r *http.Request, u User) {
	var b transaksiBody
	if !decodeBody(w, r, &b) {
		return
	}
	in, fields := parseTransaksi(b, a.today(), a.loc)
	if fields != nil {
		validationError(w, fields)
		return
	}
	id, err := a.store.CreateTransaksi(r.Context(), in, u.ID)
	if err != nil {
		internalError(w, r, err)
		return
	}
	a.kirimTransaksi(w, r, http.StatusCreated, id)
}

// Urutan cek sesuai kontrak: 401 -> 403 (withAuth) -> 400 (id, JSON) -> 422 -> 404.
func (a *App) handleUpdateTransaksi(w http.ResponseWriter, r *http.Request, u User) {
	id, ok := parseID(r)
	if !ok {
		badRequest(w)
		return
	}
	var b transaksiBody
	if !decodeBody(w, r, &b) {
		return
	}
	in, fields := parseTransaksi(b, a.today(), a.loc)
	if fields != nil {
		validationError(w, fields)
		return
	}
	err := a.store.UpdateTransaksi(r.Context(), id, in)
	if errors.Is(err, ErrNotFound) {
		notFound(w)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	a.kirimTransaksi(w, r, http.StatusOK, id)
}

func (a *App) handleDeleteTransaksi(w http.ResponseWriter, r *http.Request, u User) {
	id, ok := parseID(r)
	if !ok {
		badRequest(w)
		return
	}
	err := a.store.DeleteTransaksi(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		notFound(w)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
