package routes

import (
	"net/http"
	"os"
	"strings"
)

// fsTanpaDaftar melayani file statis hasil build frontend (STATIC_DIR) tanpa menampilkan daftar isi
// folder: folder tanpa index.html dianggap tidak ada (404).
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
