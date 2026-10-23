// FE-04 · Tipe data dan ApiError (bentuk persis dari docs/01-api-contract.md).
// Dipisah dari api.ts agar mock.ts bisa memakainya tanpa import melingkar.
// api.ts mengekspor ulang semuanya, jadi halaman cukup mengimpor dari './api'.

export type Role = 'petugas' | 'dinas';

export interface User {
  id: number;
  username: string;
  nama: string;
  role: Role;
}

export interface Transaksi {
  id: number;
  nama_nelayan: string;
  jenis_ikan: string;
  berat_kg: number;
  harga_per_kg: number;
  tanggal: string;
  nilai: number;
  retribusi: number;
  dibuat_oleh: string;
  dibuat_pada: string;
}

export interface TransaksiInput {
  // null = isian kosong atau tidak bisa dibaca sebagai angka; backend menjawab 422 dengan pesan field
  nama_nelayan: string;
  jenis_ikan: string;
  berat_kg: number | null;
  harga_per_kg: number | null;
  tanggal: string;
}

export interface Rekap {
  dari: string;
  sampai: string;
  jumlah_transaksi: number;
  total_berat_kg: number;
  total_nilai: number;
  total_retribusi: number;
  tarif_retribusi: number;
}

export interface Rentang {
  dari: string;
  sampai: string;
}

export interface DaftarTransaksi {
  data: Transaksi[];
  meta: { jumlah: number; dari: string | null; sampai: string | null; tarif_retribusi: number };
}

export class ApiError extends Error {
  status: number;
  code: string;
  fields?: Record<string, string>;
  constructor(status: number, code: string, message: string, fields?: Record<string, string>) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.fields = fields;
  }
}
