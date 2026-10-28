// FE-12 · Panel "Tanya Si Tangkap". Konteks percakapan hanya disimpan di memori halaman (BR-12):
// tidak di localStorage atau sessionStorage, dan hilang saat halaman dimuat ulang.

import { ApiError, chat } from './api';
import type { KonteksChat } from './api';
import { $, el, sembunyikan, tampilkan } from './dom';

let konteks: KonteksChat | null = null;

const buka = $<HTMLButtonElement>('chat-buka');
const panel = $('chat-panel');
const daftar = $('chat-daftar');
const form = $<HTMLFormElement>('chat-form');
const input = $<HTMLInputElement>('chat-input');
const galat = $('chat-galat');
const kirim = $<HTMLButtonElement>('chat-kirim');

buka.addEventListener('click', () => {
  const terbuka = panel.hidden;
  panel.hidden = !terbuka;
  buka.setAttribute('aria-expanded', String(terbuka));
  if (terbuka) input.focus();
});

function tambah(teks: string, kelas: string): HTMLElement {
  const p = el('p', teks, 'chat-pesan ' + kelas);
  daftar.append(p);
  daftar.scrollTop = daftar.scrollHeight;
  return p;
}

form.addEventListener('submit', async (e) => {
  e.preventDefault();
  sembunyikan(galat);
  const pesan = input.value;
  if (pesan.trim() === '') {
    tampilkan(galat, 'Pertanyaan wajib diisi.');
    return;
  }

  kirim.disabled = true;
  const saya = tambah(pesan.trim(), 'chat-saya');
  const mengetik = tambah('Si Tangkap sedang mengetik…', 'chat-bot chat-mengetik');
  try {
    const j = await chat(pesan, konteks);
    tambah(j.jawaban, 'chat-bot');
    konteks = j.konteks;
    input.value = '';
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      location.href = '/';
      return;
    }
    if (err instanceof ApiError && err.status === 422) {
      saya.remove();
      tampilkan(galat, err.fields?.pesan ?? err.message);
    } else {
      tambah(err instanceof ApiError ? err.message : 'Terjadi kesalahan pada server.', 'chat-bot chat-gagal');
    }
  } finally {
    mengetik.remove();
    kirim.disabled = false;
    input.focus();
  }
});
