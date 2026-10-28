// FE-08 · Logika halaman dinas: filter tanggal, rekap, dan daftar transaksi (hanya baca).

import { ApiError, listTransaksi, logout, rekap } from './api';
import type { Rentang, Transaksi } from './api';
import { $, el, sembunyikan, tampilkan } from './dom';
import { awalBulanISO, awalTahunISO, berat, hariIniISO, persen, rupiah, tanggalID } from './format';
import { requireRole } from './guard';

const halaman = $('halaman');
const pesanHalaman = $('pesan-halaman');
const inputDari = $<HTMLInputElement>('dari');
const inputSampai = $<HTMLInputElement>('sampai');
const isiTabel = $('isi-tabel');
const footerTabel = $('footer-tabel');
const keterangan = $('keterangan');

function tanganiGalat(err: unknown): void {
  if (err instanceof ApiError) {
    if (err.status === 401) {
      location.href = '/';
      return;
    }
    tampilkan(pesanHalaman, err.message);
  } else {
    tampilkan(pesanHalaman, 'Terjadi kesalahan pada server.');
  }
}

function sel(teks: string, kelas?: string): HTMLTableCellElement {
  return el('td', teks, kelas);
}

function barisKosong(): HTMLTableRowElement {
  const tr = el('tr');
  const td = el('td');
  td.colSpan = 7;
  const kotak = el('div', undefined, 'kosong');
  kotak.append(
    el('strong', 'Tidak ada transaksi pada rentang tanggal ini'),
    el('span', 'Coba ubah tanggal Dari dan Sampai, atau pilih Bulan ini.'),
  );
  td.append(kotak);
  tr.append(td);
  return tr;
}

function barisTransaksi(t: Transaksi): HTMLTableRowElement {
  const tr = el('tr');
  tr.append(
    sel(tanggalID(t.tanggal)),
    sel(t.nama_nelayan),
    sel(t.jenis_ikan),
    sel(berat(t.berat_kg), 'angka'),
    sel(rupiah(t.harga_per_kg), 'angka'),
    sel(rupiah(t.nilai), 'angka'),
    sel(rupiah(t.retribusi), 'angka'),
  );
  return tr;
}

async function muat(range: Rentang): Promise<void> {
  sembunyikan(pesanHalaman);
  try {
    const [r, daftar] = await Promise.all([rekap(range), listTransaksi(range)]);

    $('k-berat').textContent = `${berat(r.total_berat_kg)} kg`;
    $('k-nilai').textContent = rupiah(r.total_nilai);
    $('k-retribusi').textContent = rupiah(r.total_retribusi);
    $('k-jumlah').textContent = String(r.jumlah_transaksi);
    keterangan.textContent =
      `Periode ${tanggalID(r.dari)} sampai ${tanggalID(r.sampai)}. ` +
      `Retribusi dihitung dengan tarif ${persen(r.tarif_retribusi)}%.`;

    isiTabel.replaceChildren();
    if (daftar.data.length === 0) isiTabel.append(barisKosong());
    else for (const t of daftar.data) isiTabel.append(barisTransaksi(t));
    footerTabel.textContent = `Menampilkan ${daftar.data.length} transaksi`;
  } catch (err) {
    tanganiGalat(err);
  }
}

function terapkan(): void {
  const dari = inputDari.value;
  const sampai = inputSampai.value;
  if (!dari || !sampai) {
    tampilkan(pesanHalaman, 'Isi kedua tanggal.');
    return;
  }
  if (dari > sampai) {
    tampilkan(pesanHalaman, 'Tanggal awal tidak boleh setelah tanggal akhir.');
    return;
  }
  void muat({ dari, sampai });
}

function pakaiRentangCepat(jenis: string): void {
  const hariIni = hariIniISO();
  inputSampai.value = hariIni;
  if (jenis === 'hari') inputDari.value = hariIni;
  else if (jenis === 'bulan') inputDari.value = awalBulanISO();
  else inputDari.value = awalTahunISO();
  terapkan();
}

$('btn-terapkan').addEventListener('click', terapkan);
for (const tombol of document.querySelectorAll<HTMLButtonElement>('[data-cepat]')) {
  tombol.addEventListener('click', () => pakaiRentangCepat(tombol.dataset.cepat ?? 'bulan'));
}

$('btn-keluar').addEventListener('click', async () => {
  try {
    await logout();
  } finally {
    location.href = '/';
  }
});

async function mulai(): Promise<void> {
  try {
    const user = await requireRole('dinas');
    $('topbar-user').textContent = `${user.username} · Dinas Kabupaten/Kota`;
    halaman.hidden = false;
    inputDari.value = awalBulanISO();
    inputSampai.value = hariIniISO();
    await muat({ dari: inputDari.value, sampai: inputSampai.value });
  } catch (err) {
    halaman.hidden = false;
    tanganiGalat(err);
  }
}

mulai();
