package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
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

// logRequests mencatat satu baris per request: method, path, status, durasi. Body tidak pernah dicatat.
func (a *App) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mulai := time.Now()
		p := &pencatatStatus{ResponseWriter: w}
		next.ServeHTTP(p, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, p.status, time.Since(mulai).Round(time.Millisecond))
	})
}

// recoverPanic: panic -> log + 500 INTERNAL_ERROR.
func (a *App) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				internalError(w, r, fmt.Errorf("panic: %v", v))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// fsTanpaDaftar melayani file statis hasil build frontend tanpa menampilkan daftar isi folder:
// folder tanpa index.html dianggap tidak ada (404).
type fsTanpaDaftar struct {
	fs http.FileSystem
}

func (f fsTanpaDaftar) Open(name string) (http.File, error) {
	file, err := f.fs.Open(name)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	if info.IsDir() {
		index, err := f.fs.Open(strings.TrimSuffix(name, "/") + "/index.html")
		if err != nil {
			file.Close()
			return nil, os.ErrNotExist
		}
		index.Close()
	}
	return file, nil
}
