package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

// Pesan default dari tabel error di kontrak API bagian 1.
const (
	msgBadRequest         = "Permintaan tidak valid."
	msgUnauthorized       = "Silakan login terlebih dahulu."
	msgInvalidCredentials = "Username atau password salah."
	msgForbidden          = "Anda tidak memiliki akses untuk tindakan ini."
	msgNotFound           = "Data tidak ditemukan."
	msgValidation         = "Data yang dikirim belum benar."
	msgInternal           = "Terjadi kesalahan pada server."
)

// Batas ukuran body request (kontrak: body > 10 KB -> 400).
const maxBodyBytes = 10 << 10

type errorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type errorResponse struct {
	Error errorDetail `json:"error"`
}

type dataResponse struct {
	Data any `json:"data"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("gagal menulis respons JSON: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string, fields map[string]string) {
	writeJSON(w, status, errorResponse{Error: errorDetail{Code: code, Message: message, Fields: fields}})
}

func writeData(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, dataResponse{Data: data})
}

func badRequest(w http.ResponseWriter) {
	writeError(w, http.StatusBadRequest, "BAD_REQUEST", msgBadRequest, nil)
}

func unauthorized(w http.ResponseWriter) {
	writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", msgUnauthorized, nil)
}

func forbidden(w http.ResponseWriter) {
	writeError(w, http.StatusForbidden, "FORBIDDEN", msgForbidden, nil)
}

func notFound(w http.ResponseWriter) {
	writeError(w, http.StatusNotFound, "NOT_FOUND", msgNotFound, nil)
}

func validationError(w http.ResponseWriter, fields map[string]string) {
	writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", msgValidation, fields)
}

// internalError mencatat error asli ke log; klien hanya menerima pesan umum (NFR-008).
func internalError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("error %s %s: %v", r.Method, r.URL.Path, err)
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", msgInternal, nil)
}

// decodeBody membaca satu objek JSON (maksimal 10 KB) ke dst.
// JSON rusak, body kosong, terlalu besar, atau berisi data tambahan -> 400 BAD_REQUEST dan false.
func decodeBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		badRequest(w)
		return false
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		badRequest(w)
		return false
	}
	return true
}
