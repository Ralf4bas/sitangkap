package main

import "net/http"

// handleRekap: tanpa dari dan sampai -> tanggal 1 bulan ini sampai hari ini (zona APP_TIMEZONE).
func (a *App) handleRekap(w http.ResponseWriter, r *http.Request, u User) {
	dari, sampai, ok := parseRentang(r.URL.Query())
	if !ok {
		badRequest(w)
		return
	}
	if dari == "" {
		today := a.today()
		dari = today.Format("2006-01") + "-01"
		sampai = today.Format(formatTanggal)
	}
	rows, err := a.store.ListTransaksi(r.Context(), dari, sampai)
	if err != nil {
		internalError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, hitungRekap(rows, dari, sampai, a.cfg.Tarif))
}
