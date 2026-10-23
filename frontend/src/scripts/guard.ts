// FE-05 · Penjaga halaman petugas dan dinas.

import { ApiError, me } from './api';
import type { Role, User } from './api';

function jalurPeran(role: Role): string {
  return role === 'petugas' ? '/petugas' : '/dinas';
}

/** Pemanggilan paling awal di petugas.ts dan dinas.ts.
 *  401 -> ke halaman login; peran salah -> ke halaman perannya sendiri. Dalam kedua kasus
 *  promise tidak pernah selesai, jadi kode setelahnya tidak berjalan selama browser berpindah halaman.
 *  Error lain (misalnya jaringan) dilempar agar halaman bisa menampilkan pesannya. */
export async function requireRole(role: Role): Promise<User> {
  let user: User;
  try {
    user = await me();
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      location.href = '/';
      return new Promise<User>(() => {});
    }
    throw err;
  }
  if (user.role !== role) {
    location.href = jalurPeran(user.role);
    return new Promise<User>(() => {});
  }
  return user;
}
