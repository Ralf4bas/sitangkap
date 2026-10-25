// FE-04 · Mode mock (PUBLIC_MOCK=1): perilaku backend palsu, bentuk respons dan pesan sesuai kontrak.
// Data disimpan di memori (hilang saat halaman dimuat ulang). "Sesi" mock ada di sessionStorage.

import { ApiError } from './types';
import type { DaftarTransaksi, Rekap, Rentang, Role, Transaksi, TransaksiInput, User } from './types';
import { awalBulanISO, hariIniISO } from './format';
import { hitungNilai, hitungRetribusi } from './hitung';

const TARIF = 0.03;
const KUNCI_SESI = 'sitangkap_mock_user';

const AKUN: Array<User & { password: string }> = [
  { id: 1, username: 'petugas01', password: 'petugas123', nama: 'Petugas Contoh', role: 'petugas' },
  { id: 2, username: 'dinas01', password: 'dinas123', nama: 'Dinas Contoh', role: 'dinas' },
];

function buat(id: number, nama: string, ikan: string, beratKg: number, harga: number, tanggal: string): Transaksi {
  const nilai = hitungNilai(beratKg, harga);
  return {
    id,
    nama_nelayan: nama,
    jenis_ikan: ikan,
    berat_kg: beratKg,
    harga_per_kg: harga,
    tanggal,
    nilai,
    retribusi: hitungRetribusi(nilai, TARIF),
    dibuat_oleh: 'petugas01',
    dibuat_pada: `${tanggal}T09:00:00+08:00`,
  };
}

// 8 transaksi, sama persis dengan db/seed.sql (id sesuai urutan INSERT di seed).
// Lima baris Oktober 2026 = tabel contoh kontrak bagian 3 (total 370,75 kg).
const data: Transaksi[] = [
  buat(1, 'Pak Roni', 'Cakalang', 120.5, 28000, '2026-10-22'),
  buat(2, 'Pak Yance', 'Tuna', 45, 65000, '2026-10-22'),
  buat(3, 'Ibu Merry', 'Julung-julung', 30.25, 22000, '2026-10-22'),
  buat(4, 'Pak Hendrik', 'Tongkol', 80, 25000, '2026-10-21'),
  buat(5, 'Pak Roni', 'Cakalang', 95, 27500, '2026-10-21'),
  buat(6, 'Pak Yance', 'Tuna', 60, 63000, '2026-09-15'),
  buat(7, 'Ibu Merry', 'Layang', 40, 18000, '2026-09-10'),
  buat(8, 'Pak Hendrik', 'Cakalang', 110, 26000, '2026-08-05'),
];
let idBerikut = 9;

function galat(status: number, code: string, message: string, fields?: Record<string, string>): ApiError {
  return new ApiError(status, code, message, fields);
}
const UNAUTHORIZED = () => galat(401, 'UNAUTHORIZED', 'Silakan login terlebih dahulu.');
const FORBIDDEN = () => galat(403, 'FORBIDDEN', 'Anda tidak memiliki akses untuk tindakan ini.');
const BAD_REQUEST = () => galat(400, 'BAD_REQUEST', 'Permintaan tidak valid.');
const NOT_FOUND = () => galat(404, 'NOT_FOUND', 'Data tidak ditemukan.');

function tanpaPassword(a: User & { password: string }): User {
  return { id: a.id, username: a.username, nama: a.nama, role: a.role };
}

function sesiAktif(role?: Role): User {
  const id = Number(sessionStorage.getItem(KUNCI_SESI));
  const akun = AKUN.find((a) => a.id === id);
  if (!akun) throw UNAUTHORIZED();
  if (role && akun.role !== role) throw FORBIDDEN();
  return tanpaPassword(akun);
}

function tanggalNyata(s: string): boolean {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(s)) return false;
  const [y, m, d] = s.split('-').map(Number);
  const t = new Date(Date.UTC(y, m - 1, d));
  return t.getUTCFullYear() === y && t.getUTCMonth() === m - 1 && t.getUTCDate() === d;
}

function validasi(input: TransaksiInput): Record<string, string> {
  const f: Record<string, string> = {};

  const nama = (input.nama_nelayan ?? '').trim();
  if (nama.length === 0) f.nama_nelayan = 'Nama nelayan wajib diisi.';
  else if ([...nama].length > 100) f.nama_nelayan = 'Nama nelayan paling banyak 100 karakter.';

  const ikan = (input.jenis_ikan ?? '').trim();
  if (ikan.length === 0) f.jenis_ikan = 'Jenis ikan wajib diisi.';
  else if ([...ikan].length > 60) f.jenis_ikan = 'Jenis ikan paling banyak 60 karakter.';

  const b = input.berat_kg;
  if (typeof b !== 'number' || !Number.isFinite(b)) f.berat_kg = 'Berat wajib diisi dengan angka.';
  else if (b <= 0) f.berat_kg = 'Berat harus lebih dari 0.';
  else if (b > 100000) f.berat_kg = 'Berat paling besar 100000 kg.';
  else if (Math.abs(b * 100 - Math.round(b * 100)) > 1e-7) f.berat_kg = 'Berat paling banyak 2 angka di belakang koma.';

  const h = input.harga_per_kg;
  if (typeof h !== 'number' || !Number.isInteger(h)) f.harga_per_kg = 'Harga per kg wajib diisi dengan bilangan bulat.';
  else if (h <= 0) f.harga_per_kg = 'Harga per kg harus lebih dari 0.';
  else if (h > 100000000) f.harga_per_kg = 'Harga per kg paling besar Rp 100.000.000.';

  const t = input.tanggal ?? '';
  if (!tanggalNyata(t)) f.tanggal = 'Tanggal harus berformat YYYY-MM-DD dan valid.';
  else if (t > hariIniISO()) f.tanggal = 'Tanggal tidak boleh melewati hari ini.';

  return f;
}

function periksaRentang(range?: Rentang): void {
  if (!range) return;
  if (!tanggalNyata(range.dari) || !tanggalNyata(range.sampai) || range.dari > range.sampai) throw BAD_REQUEST();
}

function periksaId(id: number): void {
  if (!Number.isInteger(id) || id <= 0) throw BAD_REQUEST();
}

function urut(daftar: Transaksi[]): Transaksi[] {
  return [...daftar].sort((a, b) => (a.tanggal === b.tanggal ? b.id - a.id : a.tanggal < b.tanggal ? 1 : -1));
}

function dalamRentang(range?: Rentang): Transaksi[] {
  return data.filter((t) => !range || (t.tanggal >= range.dari && t.tanggal <= range.sampai));
}

function simpanDariInput(input: TransaksiInput, id: number, dibuatOleh: string, dibuatPada: string): Transaksi {
  const nilai = hitungNilai(input.berat_kg as number, input.harga_per_kg as number);
  return {
    id,
    nama_nelayan: input.nama_nelayan.trim(),
    jenis_ikan: input.jenis_ikan.trim(),
    berat_kg: input.berat_kg as number,
    harga_per_kg: input.harga_per_kg as number,
    tanggal: input.tanggal,
    nilai,
    retribusi: hitungRetribusi(nilai, TARIF),
    dibuat_oleh: dibuatOleh,
    dibuat_pada: dibuatPada,
  };
}

function waktuSekarang(): string {
  const d = new Date();
  const dua = (n: number) => String(n).padStart(2, '0');
  return `${hariIniISO()}T${dua(d.getHours())}:${dua(d.getMinutes())}:${dua(d.getSeconds())}+08:00`;
}

// ---- Fungsi yang dipanggil api.ts (tanda tangan sama dengan api.ts) ----

export async function login(username: string, password: string): Promise<User> {
  const f: Record<string, string> = {};
  if (!username.trim()) f.username = 'Username wajib diisi.';
  if (!password) f.password = 'Password wajib diisi.';
  if (Object.keys(f).length > 0) throw galat(422, 'VALIDATION_ERROR', 'Data yang dikirim belum benar.', f);

  const akun = AKUN.find((a) => a.username === username.trim() && a.password === password);
  if (!akun) throw galat(401, 'INVALID_CREDENTIALS', 'Username atau password salah.');
  sessionStorage.setItem(KUNCI_SESI, String(akun.id));
  return tanpaPassword(akun);
}

export async function logout(): Promise<void> {
  sessionStorage.removeItem(KUNCI_SESI);
}

export async function me(): Promise<User> {
  return sesiAktif();
}

export async function listTransaksi(range?: Rentang): Promise<DaftarTransaksi> {
  sesiAktif();
  periksaRentang(range);
  const hasil = urut(dalamRentang(range));
  return {
    data: hasil,
    meta: { jumlah: hasil.length, dari: range ? range.dari : null, sampai: range ? range.sampai : null, tarif_retribusi: TARIF },
  };
}

export async function createTransaksi(input: TransaksiInput): Promise<Transaksi> {
  const user = sesiAktif('petugas');
  const f = validasi(input);
  if (Object.keys(f).length > 0) throw galat(422, 'VALIDATION_ERROR', 'Data yang dikirim belum benar.', f);
  const baru = simpanDariInput(input, idBerikut++, user.username, waktuSekarang());
  data.push(baru);
  return baru;
}

export async function updateTransaksi(id: number, input: TransaksiInput): Promise<Transaksi> {
  sesiAktif('petugas');
  periksaId(id);
  const f = validasi(input);
  if (Object.keys(f).length > 0) throw galat(422, 'VALIDATION_ERROR', 'Data yang dikirim belum benar.', f);
  const i = data.findIndex((t) => t.id === id);
  if (i < 0) throw NOT_FOUND();
  const lama = data[i];
  const baru = simpanDariInput(input, lama.id, lama.dibuat_oleh, lama.dibuat_pada);
  data[i] = baru;
  return baru;
}

export async function deleteTransaksi(id: number): Promise<void> {
  sesiAktif('petugas');
  periksaId(id);
  const i = data.findIndex((t) => t.id === id);
  if (i < 0) throw NOT_FOUND();
  data.splice(i, 1);
}

export async function rekap(range?: Rentang): Promise<Rekap> {
  sesiAktif('dinas');
  periksaRentang(range);
  const r: Rentang = range ?? { dari: awalBulanISO(), sampai: hariIniISO() };
  const hasil = dalamRentang(r);
  return {
    dari: r.dari,
    sampai: r.sampai,
    jumlah_transaksi: hasil.length,
    total_berat_kg: Math.round(hasil.reduce((s, t) => s + t.berat_kg, 0) * 100) / 100,
    total_nilai: hasil.reduce((s, t) => s + t.nilai, 0),
    total_retribusi: hasil.reduce((s, t) => s + t.retribusi, 0),
    tarif_retribusi: TARIF,
  };
}
