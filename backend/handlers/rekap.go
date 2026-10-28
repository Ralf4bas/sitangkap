package handlers

import (
	"net/http"

	"sitangkap/backend/helpers"
	"sitangkap/backend/models"
	"sitangkap/backend/services"
)

// RekapHandler: GET /api/rekap. Tanpa dari dan sampai -> tanggal 1 bulan ini sampai hari ini (zona APP_TIMEZONE).
func (h *Handler) RekapHandler(w http.ResponseWriter, r *http.Request, u models.User) {
	dari, sampai, ok := services.ParseRentang(r.URL.Query())
	if !ok {
		helpers.BadRequest(w)
		return
	}
	if dari == "" {
		today := h.today()
		dari = today.Format("2006-01") + "-01"
		sampai = today.Format(services.FormatTanggal)
	}
	rows, err := h.Store.ListTransaksi(r.Context(), dari, sampai)
	if err != nil {
		helpers.InternalError(w, r, err)
		return
	}
	helpers.WriteData(w, http.StatusOK, services.HitungRekap(rows, dari, sampai, h.Cfg.Tarif))
}
