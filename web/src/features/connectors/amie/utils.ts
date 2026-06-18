export function ageHoursOf(iso: string | undefined): number {
  if (!iso) return 0;
  const t = Date.parse(iso);
  if (Number.isNaN(t)) return 0;
  return (Date.now() - t) / 3600_000;
}

export function formatDate(iso?: string): string {
  if (!iso) return "—";
  const t = Date.parse(iso);
  if (Number.isNaN(t)) return "—";
  return new Date(t).toLocaleString();
}

export function pluralize(noun: string, count: number, pluralForm?: string): string {
  if (count === 1) return noun;
  return pluralForm ?? `${noun}s`;
}
