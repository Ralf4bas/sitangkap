// FE-04 · Semua panggilan ke backend. Path relatif /api/..., tanpa redirect (itu tugas guard.ts).

import { ApiError } from './types';
import type { DaftarTransaksi, JawabanChat, KonteksChat, Rekap, Rentang, Transaksi, TransaksiInput, User } from './types';
import * as mock from './mock';

export * from './types';

// Mode mock hanya aktif bila PUBLIC_MOCK=1 (Jenkins membangun dengan PUBLIC_MOCK=0).
const MOCK = import.meta.env.PUBLIC_MOCK === '1';

interface Badan<T> {
  data: T;
  meta?: DaftarTransaksi['meta'];
}

async function panggil<T>(method: string, path: string, body?: unknown): Promise<Badan<T> | null> {
  let res: Response;
  try {
    res = await fetch(path, {
      method,
      headers: body !== undefined ? { 'Content-Type': 'application/json' } : undefined,
      body: body !== undefined ? JSON.stringify(body) : undefined,
      credentials: 'same-origin',
    });
  } catch {
    throw new ApiError(0, 'NETWORK', 'Tidak bisa terhubung ke server.');
  }

  if (res.status === 204) return null;

  let json: any;
  try {
    json = await res.json();
  } catch {
    throw new ApiError(res.status, 'INTERNAL_ERROR', 'Terjadi kesalahan pada server.');
  }

  if (!res.ok) {
    const e = json && json.error ? json.error : {};
    throw new ApiError(
      res.status,
      e.code ?? 'INTERNAL_ERROR',
      e.message ?? 'Terjadi kesalahan pada server.',
      e.fields,
    );
  }
  return json as Badan<T>;
}

function query(range?: Rentang): string {
  if (!range) return '';
  const p = new URLSearchParams({ dari: range.dari, sampai: range.sampai });
  return '?' + p.toString();
}

export async function login(username: string, password: string): Promise<User> {
  if (MOCK) return mock.login(username, password);
  const r = await panggil<User>('POST', '/api/login', { username, password });
  return r!.data;
}

export async function logout(): Promise<void> {
  if (MOCK) return mock.logout();
  await panggil('POST', '/api/logout');
}

export async function me(): Promise<User> {
  if (MOCK) return mock.me();
  const r = await panggil<User>('GET', '/api/me');
  return r!.data;
}

export async function listTransaksi(range?: Rentang): Promise<DaftarTransaksi> {
  if (MOCK) return mock.listTransaksi(range);
  const r = await panggil<Transaksi[]>('GET', '/api/transaksi' + query(range));
  return { data: r!.data, meta: r!.meta! };
}

export async function createTransaksi(input: TransaksiInput): Promise<Transaksi> {
  if (MOCK) return mock.createTransaksi(input);
  const r = await panggil<Transaksi>('POST', '/api/transaksi', input);
  return r!.data;
}

export async function updateTransaksi(id: number, input: TransaksiInput): Promise<Transaksi> {
  if (MOCK) return mock.updateTransaksi(id, input);
  const r = await panggil<Transaksi>('PUT', `/api/transaksi/${id}`, input);
  return r!.data;
}

export async function deleteTransaksi(id: number): Promise<void> {
  if (MOCK) return mock.deleteTransaksi(id);
  await panggil('DELETE', `/api/transaksi/${id}`);
}

export async function rekap(range?: Rentang): Promise<Rekap> {
  if (MOCK) return mock.rekap(range);
  const r = await panggil<Rekap>('GET', '/api/rekap' + query(range));
  return r!.data;
}

// konteks = konteks dari jawaban sebelumnya; hanya dikirim bila ada.
export async function chat(pesan: string, konteks?: KonteksChat | null): Promise<JawabanChat> {
  if (MOCK) return mock.chat(pesan, konteks);
  const r = await panggil<JawabanChat>('POST', '/api/chat', konteks ? { pesan, konteks } : { pesan });
  return r!.data;
}
