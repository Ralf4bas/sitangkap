// FE-03 · Format tampilan dan pembacaan input angka.

const fmtRupiah = new Intl.NumberFormat('id-ID');
const fmtBerat = new Intl.NumberFormat('id-ID', { maximumFractionDigits: 2 });

/** 3374000 -> "Rp 3.374.000" */
export function rupiah(n: number): string {
  return 'Rp ' + fmtRupiah.format(n);
}

/** 120.5 -> "120,5"; 45 -> "45"; 30.25 -> "30,25" */
export function berat(n: number): string {
  return fmtBerat.format(n);
}

/** "2026-10-22" -> "22/10/2026". Sengaja memotong string, bukan lewat new Date(). */
export function tanggalID(s: string): string {
  return `${s.slice(8, 10)}/${s.slice(5, 7)}/${s.slice(0, 4)}`;
}

function dua(n: number): string {
  return String(n).padStart(2, '0');
}

/** Tanggal lokal browser YYYY-MM-DD. Bukan toISOString() (itu UTC dan bisa mundur sehari). */
export function hariIniISO(): string {
  const d = new Date();
  return `${d.getFullYear()}-${dua(d.getMonth() + 1)}-${dua(d.getDate())}`;
}

/** Tanggal 1 bulan ini. */
export function awalBulanISO(): string {
  const d = new Date();
  return `${d.getFullYear()}-${dua(d.getMonth() + 1)}-01`;
}

/** Tanggal 1 Januari tahun ini. */
export function awalTahunISO(): string {
  return `${new Date().getFullYear()}-01-01`;
}

/** "120,5" atau "120.5" -> 120.5; kosong atau tidak valid -> null. Nilai negatif tetap dibaca,
 *  supaya backend yang menolaknya dengan pesan resmi. */
export function parseBerat(s: string): number | null {
  const t = s.trim();
  if (!/^-?\d+([.,]\d+)?$/.test(t)) return null;
  return Number(t.replace(',', '.'));
}

/** "28.000" atau "28000" -> 28000; kosong atau tidak valid -> null. */
export function parseHarga(s: string): number | null {
  const t = s.trim();
  if (/^-?\d+$/.test(t)) return Number(t);
  if (/^-?\d{1,3}(\.\d{3})+$/.test(t)) return Number(t.replace(/\./g, ''));
  return null;
}

/** 0.03 -> "3"; 0.0285 -> "2,85" */
export function persen(tarif: number): string {
  return (Math.round(tarif * 10000) / 100).toLocaleString('id-ID');
}
