export function fmtVND(n: number): string {
  if (n >= 1_000_000) {
    const v = (n / 1_000_000).toFixed(n % 1_000_000 === 0 ? 0 : 2).replace(/\.?0+$/, "");
    return `${v}M`;
  }
  if (n >= 1_000) return `${Math.round(n / 1_000)}k`;
  return String(n);
}

export function fmtVNDfull(n: number): string {
  return `${n.toLocaleString("en-US")} ₫`;
}
