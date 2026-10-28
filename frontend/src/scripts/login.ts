// FE-06 · Logika halaman login.

import { ApiError, login, me } from './api';
import type { Role } from './api';
import { $, sembunyikan, tampilkan } from './dom';

const form = $<HTMLFormElement>('form-login');
const username = $<HTMLInputElement>('username');
const password = $<HTMLInputElement>('password');
const tombol = $<HTMLButtonElement>('btn-masuk');
const pesan = $('pesan-error');

function keHalaman(role: Role): void {
  location.href = role === 'petugas' ? '/petugas' : '/dinas';
}

// Sudah login? Langsung ke halaman sesuai peran. Gagal (401 atau jaringan) = tetap di halaman login.
me()
  .then((u) => keHalaman(u.role))
  .catch(() => {});

form.addEventListener('submit', async (e) => {
  e.preventDefault();
  sembunyikan(pesan);
  tombol.disabled = true;
  tombol.textContent = 'Memproses...';
  try {
    const user = await login(username.value, password.value);
    keHalaman(user.role);
  } catch (err) {
    if (err instanceof ApiError) {
      const rincian = err.fields ? Object.values(err.fields).join(' ') : '';
      tampilkan(pesan, rincian || err.message);
    } else {
      tampilkan(pesan, 'Terjadi kesalahan pada server.');
    }
    tombol.disabled = false;
    tombol.textContent = 'Masuk';
  }
});
