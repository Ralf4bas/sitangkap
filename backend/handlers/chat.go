package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"sitangkap/backend/helpers"
	"sitangkap/backend/models"
	"sitangkap/backend/services"
)

// Pesan validasi persis dari kontrak API (POST /api/chat).
const (
	MsgPertanyaanWajib   = "Pertanyaan wajib diisi."
	MsgPertanyaanPanjang = "Pertanyaan paling banyak 200 karakter."
)

type chatBody struct {
	Pesan   json.RawMessage `json:"pesan"`
	Konteks json.RawMessage `json:"konteks"`
}

// ChatHandler: POST /api/chat. Hanya membaca data; isi pesan tidak pernah dicatat ke log (BR-12).
func (h *Handler) ChatHandler(w http.ResponseWriter, r *http.Request, u models.User) {
	var b chatBody
	if !helpers.DecodeBody(w, r, &b) {
		return
	}
	var pesan string
	if json.Unmarshal(b.Pesan, &pesan) != nil || strings.TrimSpace(pesan) == "" {
		helpers.ValidationError(w, map[string]string{"pesan": MsgPertanyaanWajib})
		return
	}
	if utf8.RuneCountInString(strings.TrimSpace(pesan)) > 200 {
		helpers.ValidationError(w, map[string]string{"pesan": MsgPertanyaanPanjang})
		return
	}

	k := services.BacaKonteks(b.Konteks)
	p := services.PahamiPesan(pesan, k, h.today())
	var rows []models.TransaksiRow
	var err error
	switch {
	case p.TanggalTidakValid:
	case p.Dari != "":
		rows, err = h.Store.ListTransaksi(r.Context(), p.Dari, p.Sampai)
	case p.Niat == services.NiatTransaksiTerakhir:
		rows, err = h.Store.ListTransaksi(r.Context(), "", "")
	}
	if err != nil {
		helpers.InternalError(w, r, err)
		return
	}
	log.Printf("chat niat=%q", p.Niat)
	helpers.WriteData(w, http.StatusOK, services.SusunJawaban(p, k, rows, h.Cfg.Tarif, u.Username))
}
