package models

// KonteksChat adalah objek "Konteks chat" di kontrak API. Server tidak menyimpannya:
// frontend menyimpan di memori halaman dan mengirimnya kembali bersama pertanyaan berikutnya (BR-12).
type KonteksChat struct {
	Niat   *string  `json:"niat"`
	Metrik []string `json:"metrik"`
	Dari   *string  `json:"dari"`
	Sampai *string  `json:"sampai"`
	Gaya   string   `json:"gaya"`
}

// JawabanChat adalah objek "Jawaban chat" di kontrak API (POST /api/chat).
type JawabanChat struct {
	Jawaban  string      `json:"jawaban"`
	Dikenali bool        `json:"dikenali"`
	Niat     *string     `json:"niat"`
	Metrik   []string    `json:"metrik"`
	Dari     *string     `json:"dari"`
	Sampai   *string     `json:"sampai"`
	Konteks  KonteksChat `json:"konteks"`
}
