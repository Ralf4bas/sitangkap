// Package helpers berisi helper respons JSON yang dipakai handlers dan middlewares,
// supaya bentuk respons sukses dan error selalu sama dengan kontrak API bagian 1.
package helpers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

// Pesan default dari tabel error di kontrak API bagian 1.
const (
	MsgBadRequest         = "Permintaan tidak valid."
	MsgUnauthorized       = "Silakan login terlebih dahulu."
	MsgInvalidCredentials = "Username atau password salah."
	MsgForbidden          = "Anda tidak memiliki akses untuk tindakan ini."
	MsgNotFound           = "Data tidak ditemukan."
	MsgValidation         = "Data yang dikirim belum benar."
	MsgInternal           = "Terjadi kesalahan pada server."
)

// MaxBodyBytes adalah batas ukuran body request (kontrak: body > 10 KB -> 400).
const MaxBodyBytes = 10 << 10

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

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("gagal menulis respons JSON: %v", err)
	}
}

func WriteError(w http.ResponseWriter, status int, code, message string, fields map[string]string) {
	WriteJSON(w, status, errorResponse{Error: errorDetail{Code: code, Message: message, Fields: fields}})
}

// WriteData menulis respons sukses {"data": ...}.
func WriteData(w http.ResponseWriter, status int, data any) {
	WriteJSON(w, status, dataResponse{Data: data})
}

func BadRequest(w http.ResponseWriter) {
	WriteError(w, http.StatusBadRequest, "BAD_REQUEST", MsgBadRequest, nil)
}

func Unauthorized(w http.ResponseWriter) {
	WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", MsgUnauthorized, nil)
}

func InvalidCredentials(w http.ResponseWriter) {
	WriteError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", MsgInvalidCredentials, nil)
}

func Forbidden(w http.ResponseWriter) {
	WriteError(w, http.StatusForbidden, "FORBIDDEN", MsgForbidden, nil)
}

func NotFound(w http.ResponseWriter) {
	WriteError(w, http.StatusNotFound, "NOT_FOUND", MsgNotFound, nil)
}

func ValidationError(w http.ResponseWriter, fields map[string]string) {
	WriteError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", MsgValidation, fields)
}

// InternalError mencatat error asli ke log; klien hanya menerima pesan umum (NFR-008).
func InternalError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("error %s %s: %v", r.Method, r.URL.Path, err)
	WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", MsgInternal, nil)
}

// DecodeBody membaca satu objek JSON (maksimal 10 KB) ke dst.
// JSON rusak, body kosong, terlalu besar, atau berisi data tambahan -> 400 BAD_REQUEST dan false.
func DecodeBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		BadRequest(w)
		return false
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		BadRequest(w)
		return false
	}
	return true
}
