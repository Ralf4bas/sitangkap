package main

import (
	"context"
	"errors"
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

func TestMemStoreUrutanDanNotFound(t *testing.T) {
	s := newMemStore()
	s.isiContohKontrak()
	ctx := context.Background()

	rows, err := s.ListTransaksi(ctx, "", "")
	if err != nil {
		t.Fatal(err)
	}
	var ids []int64
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	want := []int64{3, 2, 1, 5, 4}
	if len(ids) != len(want) {
		t.Fatalf("ids = %v, mau %v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("ids = %v, mau %v", ids, want)
		}
	}

	rentang, _ := s.ListTransaksi(ctx, "2026-10-21", "2026-10-21")
	if len(rentang) != 2 {
		t.Errorf("rentang 21 Okt = %d baris, mau 2", len(rentang))
	}
	if _, err := s.GetTransaksi(ctx, 99); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetTransaksi(99) err = %v", err)
	}
	if err := s.DeleteTransaksi(ctx, 99); !errors.Is(err, ErrNotFound) {
		t.Errorf("DeleteTransaksi(99) err = %v", err)
	}
}
