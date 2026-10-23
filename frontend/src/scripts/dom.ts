// Pembantu kecil untuk skrip halaman. Data dari API selalu masuk lewat textContent / createElement
// (tidak pernah disisipkan sebagai markup HTML), supaya aman dari XSS.

export function $<T extends HTMLElement = HTMLElement>(id: string): T {
  const el = document.getElementById(id);
  if (!el) throw new Error(`Elemen #${id} tidak ditemukan`);
  return el as T;
}

export function el<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  teks?: string,
  kelas?: string,
): HTMLElementTagNameMap[K] {
  const e = document.createElement(tag);
  if (teks !== undefined) e.textContent = teks;
  if (kelas) e.className = kelas;
  return e;
}

export function tampilkan(e: HTMLElement, pesan: string): void {
  e.textContent = pesan;
  e.hidden = false;
}

export function sembunyikan(e: HTMLElement): void {
  e.textContent = '';
  e.hidden = true;
}
