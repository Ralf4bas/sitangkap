// FE-03 · Pratinjau nilai dan retribusi di form. Rumus sama dengan backend (kontrak bagian 3).

export function hitungNilai(beratKg: number, hargaPerKg: number): number {
  const centi = Math.round(beratKg * 100);
  return Math.floor((centi * hargaPerKg + 50) / 100);
}

export function hitungRetribusi(nilai: number, tarif: number): number {
  return Math.round(nilai * tarif);
}
