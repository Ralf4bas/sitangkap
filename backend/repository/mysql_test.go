package repository

import (
	"testing"
)

func TestBeratKeCentiDanSebaliknya(t *testing.T) {
	kasus := []struct {
		db    string
		centi int64
	}{
		{"120.50", 12050}, {"45.00", 4500}, {"30.25", 3025}, {"0.01", 1}, {"100000.00", 10000000}, {"0.29", 29},
	}
	for _, k := range kasus {
		c, err := beratKeCenti(k.db)
		if err != nil || c != k.centi {
			t.Errorf("beratKeCenti(%q) = %d, %v; mau %d", k.db, c, err, k.centi)
		}
		if got := centiKeDecimal(k.centi); got != k.db {
			t.Errorf("centiKeDecimal(%d) = %q; mau %q", k.centi, got, k.db)
		}
	}
	if _, err := beratKeCenti("bukan angka"); err == nil {
		t.Error("harus error untuk teks bukan angka")
	}
}
