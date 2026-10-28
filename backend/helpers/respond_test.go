package helpers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteErrorTanpaFields(t *testing.T) {
	rec := httptest.NewRecorder()
	NotFound(rec)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
	want := `{"error":{"code":"NOT_FOUND","message":"Data tidak ditemukan."}}`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Errorf("body = %s, mau %s", got, want)
	}
}

func TestWriteErrorDenganFields(t *testing.T) {
	rec := httptest.NewRecorder()
	ValidationError(rec, map[string]string{"berat_kg": "Berat harus lebih dari 0."})
	want := `{"error":{"code":"VALIDATION_ERROR","message":"Data yang dikirim belum benar.","fields":{"berat_kg":"Berat harus lebih dari 0."}}}`
	if rec.Code != http.StatusUnprocessableEntity || strings.TrimSpace(rec.Body.String()) != want {
		t.Errorf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestWriteDataSliceKosong(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteData(rec, http.StatusOK, make([]int, 0))
	if got := strings.TrimSpace(rec.Body.String()); got != `{"data":[]}` {
		t.Errorf("body = %s", got)
	}
}

func TestDecodeBody(t *testing.T) {
	kasus := []struct {
		nama string
		body string
		ok   bool
	}{
		{"objek valid", `{"username":"a"}`, true},
		{"JSON rusak", `{"username":`, false},
		{"kosong", ``, false},
		{"data tambahan", `{"username":"a"} {}`, false},
		{"lebih dari 10 KB", `{"username":"` + strings.Repeat("a", MaxBodyBytes) + `"}`, false},
	}
	for _, k := range kasus {
		t.Run(k.nama, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(k.body))
			var dst struct {
				Username string `json:"username"`
			}
			if got := DecodeBody(rec, req, &dst); got != k.ok {
				t.Fatalf("decodeBody = %v, mau %v", got, k.ok)
			}
			if !k.ok && (rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "BAD_REQUEST")) {
				t.Errorf("status %d body %s", rec.Code, rec.Body.String())
			}
		})
	}
}
