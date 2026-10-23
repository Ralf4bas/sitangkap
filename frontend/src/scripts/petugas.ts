// FE-07 · Logika halaman petugas: daftar, tambah, ubah, hapus.

import {
  ApiError,
  createTransaksi,
  deleteTransaksi,
  listTransaksi,
  logout,
  updateTransaksi,
} from './api';
import type { Transaksi, TransaksiInput } from './api';
import { $, el, sembunyikan, tampilkan } from './dom';
import { berat, hariIniISO, parseBerat, parseHarga, persen, rupiah, tanggalID } from './format';
import { requireRole } from './guard';
import { hitungNilai, hitungRetribusi } from './hitung';

const halaman = $('halaman');
const pesanHalaman = $('pesan-halaman');
const isiTabel = $('isi-tabel');
const footerTabel = $('footer-tabel');

const dlgForm = $<HTMLDialogElement>('dlg-form');
const formTransaksi = $<HTMLFormElement>('form-transaksi');
const judulForm = $('judul-form');
const galatForm = $('galat-form');
const fNama = $<HTMLInputElement>('f-nama');
const fIkan = $<HTMLInputElement>('f-ikan');
const fBerat = $<HTMLInputElement>('f-berat');
const fHarga = $<HTMLInputElement>('f-harga');
const fTanggal = $<HTMLInputElement>('f-tanggal');
const btnSimpan = $<HTMLButtonElement>('btn-simpan');
const pvNilai = $('pv-nilai');
const pvRetribusi = $('pv-retribusi');
const pvLabelRetribusi = $('pv-label-retribusi');

const dlgHapus = $<HTMLDialogElement>('dlg-hapus');
const rincianHapus = $('rincian-hapus');
const galatHapus = $('galat-hapus');
const btnHapus = $<HTMLButtonElement>('btn-hapus');

// nama field API -> id elemen pesan error di form
const PESAN_FIELD: Record<string, string> = {
  nama_nelayan: 'e-nama',
  jenis_ikan: 'e-ikan',
  berat_kg: 'e-berat',
  harga_per_kg: 'e-harga',
  tanggal: 'e-tanggal',
};

let tarif: number | null = null; // dari meta.tarif_retribusi
let daftar: Transaksi[] = [];
let idDiubah: number | null = null; // null = mode tambah
let idDihapus: number | null = null;

function tanganiGalat(err: unknown, tampil: (pesan: string) => void): void {
  if (err instanceof ApiError) {
    if (err.status === 401) {
      location.href = '/';
      return;
    }
    tampil(err.message);
  } else {
    tampil('Terjadi kesalahan pada server.');
  }
}

// ---------- Tabel ----------

function sel(teks: string, kelas?: string): HTMLTableCellElement {
  return el('td', teks, kelas);
}

function barisKosong(): HTMLTableRowElement {
  const tr = el('tr');
  const td = el('td');
  td.colSpan = 8;
  const kotak = el('div', undefined, 'kosong');
  kotak.append(
    el('strong', 'Belum ada transaksi'),
    el('span', 'Catat hasil lelang pertama dengan tombol "+ Tambah transaksi".'),
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

  const aksi = el('td', undefined, 'aksi');
  const ubah = el('button', 'Ubah', 'btn btn-sekunder');
  ubah.type = 'button';
  ubah.addEventListener('click', () => bukaFormUbah(t));
  const hapus = el('button', 'Hapus', 'btn btn-sekunder');
  hapus.type = 'button';
  hapus.addEventListener('click', () => bukaKonfirmasiHapus(t));
  aksi.append(ubah, hapus);
  tr.append(aksi);
  return tr;
}

function gambarTabel(): void {
  isiTabel.replaceChildren();
  if (daftar.length === 0) isiTabel.append(barisKosong());
  else for (const t of daftar) isiTabel.append(barisTransaksi(t));
  footerTabel.textContent = `Menampilkan ${daftar.length} transaksi`;
}

async function muatDaftar(): Promise<void> {
  sembunyikan(pesanHalaman);
  try {
    const hasil = await listTransaksi();
    daftar = hasil.data;
    tarif = hasil.meta.tarif_retribusi;
    gambarTabel();
    perbaruiPratinjau();
  } catch (err) {
    tanganiGalat(err, (p) => tampilkan(pesanHalaman, p));
  }
}

// ---------- Form tambah/ubah ----------

function bersihkanGalatForm(): void {
  sembunyikan(galatForm);
  for (const id of Object.values(PESAN_FIELD)) sembunyikan($(id));
}

function perbaruiPratinjau(): void {
  if (tarif !== null) pvLabelRetribusi.textContent = `Retribusi (tarif ${persen(tarif)}%)`;
  else pvLabelRetribusi.textContent = 'Retribusi';

  const b = parseBerat(fBerat.value);
  const h = parseHarga(fHarga.value);
  if (b === null || h === null || b <= 0 || h <= 0) {
    pvNilai.textContent = '–';
    pvRetribusi.textContent = '–';
    return;
  }
  const nilai = hitungNilai(b, h);
  pvNilai.textContent = rupiah(nilai);
  pvRetribusi.textContent = tarif !== null ? rupiah(hitungRetribusi(nilai, tarif)) : '–';
}

function isiForm(t: Transaksi | null): void {
  fNama.value = t ? t.nama_nelayan : '';
  fIkan.value = t ? t.jenis_ikan : '';
  fBerat.value = t ? berat(t.berat_kg) : '';
  fHarga.value = t ? String(t.harga_per_kg) : '';
  fTanggal.value = t ? t.tanggal : hariIniISO();
  fTanggal.max = hariIniISO();
  bersihkanGalatForm();
  perbaruiPratinjau();
}

function bukaFormTambah(): void {
  idDiubah = null;
  judulForm.textContent = 'Tambah transaksi';
  btnSimpan.textContent = 'Simpan';
  isiForm(null);
  dlgForm.showModal();
  fNama.focus();
}

function bukaFormUbah(t: Transaksi): void {
  idDiubah = t.id;
  judulForm.textContent = 'Ubah transaksi';
  btnSimpan.textContent = 'Simpan perubahan';
  isiForm(t);
  dlgForm.showModal();
  fNama.focus();
}

function bacaForm(): TransaksiInput {
  return {
    nama_nelayan: fNama.value,
    jenis_ikan: fIkan.value,
    berat_kg: parseBerat(fBerat.value),
    harga_per_kg: parseHarga(fHarga.value),
    tanggal: fTanggal.value,
  };
}

formTransaksi.addEventListener('submit', async (e) => {
  e.preventDefault();
  bersihkanGalatForm();
  btnSimpan.disabled = true;
  try {
    const input = bacaForm();
    if (idDiubah === null) await createTransaksi(input);
    else await updateTransaksi(idDiubah, input);
    dlgForm.close();
    await muatDaftar();
  } catch (err) {
    if (err instanceof ApiError && err.status === 422 && err.fields) {
      let adaDiLuarForm = false;
      for (const [nama, pesan] of Object.entries(err.fields)) {
        const id = PESAN_FIELD[nama];
        if (id) tampilkan($(id), pesan);
        else adaDiLuarForm = true;
      }
      if (adaDiLuarForm) tampilkan(galatForm, err.message);
    } else {
      tanganiGalat(err, (p) => tampilkan(galatForm, p));
    }
  } finally {
    btnSimpan.disabled = false;
  }
});

for (const input of [fBerat, fHarga]) input.addEventListener('input', perbaruiPratinjau);
$('btn-tambah').addEventListener('click', bukaFormTambah);
$('btn-batal-form').addEventListener('click', () => dlgForm.close());

// ---------- Konfirmasi hapus ----------

function bukaKonfirmasiHapus(t: Transaksi): void {
  idDihapus = t.id;
  rincianHapus.textContent =
    `${t.nama_nelayan} · ${t.jenis_ikan} · ${berat(t.berat_kg)} kg · ${tanggalID(t.tanggal)}. ` +
    'Data yang dihapus tidak bisa dikembalikan.';
  sembunyikan(galatHapus);
  dlgHapus.showModal();
}

btnHapus.addEventListener('click', async () => {
  if (idDihapus === null) return;
  btnHapus.disabled = true;
  try {
    await deleteTransaksi(idDihapus);
    dlgHapus.close();
    await muatDaftar();
  } catch (err) {
    tanganiGalat(err, (p) => tampilkan(galatHapus, p));
  } finally {
    btnHapus.disabled = false;
  }
});
$('btn-batal-hapus').addEventListener('click', () => dlgHapus.close());

// ---------- Mulai ----------

$('btn-keluar').addEventListener('click', async () => {
  try {
    await logout();
  } finally {
    location.href = '/';
  }
});

async function mulai(): Promise<void> {
  try {
    const user = await requireRole('petugas');
    $('topbar-user').textContent = `${user.username} · Petugas TPI`;
    halaman.hidden = false;
    await muatDaftar();
  } catch (err) {
    halaman.hidden = false;
    tanganiGalat(err, (p) => tampilkan(pesanHalaman, p));
  }
}

mulai();
