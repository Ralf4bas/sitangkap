package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"sitangkap/backend/helpers"
)

// pencatatStatus merekam kode status untuk log request.
type pencatatStatus struct {
	http.ResponseWriter
	status int
}

func (p *pencatatStatus) WriteHeader(code int) {
	if p.status == 0 {
		p.status = code
	}
	p.ResponseWriter.WriteHeader(code)
}

func (p *pencatatStatus) Write(b []byte) (int, error) {
	if p.status == 0 {
		p.status = http.StatusOK
	}
	return p.ResponseWriter.Write(b)
}

// LogRequests mencatat satu baris per request: method, path, status, durasi. Body tidak pernah dicatat.
func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mulai := time.Now()
		p := &pencatatStatus{ResponseWriter: w}
		next.ServeHTTP(p, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, p.status, time.Since(mulai).Round(time.Millisecond))
	})
}

// RecoverPanic: panic -> log + 500 INTERNAL_ERROR.
func RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				helpers.InternalError(w, r, fmt.Errorf("panic: %v", v))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
