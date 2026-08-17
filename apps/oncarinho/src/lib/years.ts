import type { Matchday } from "./api/types";

export function availableYears(matchdays: Matchday[]): number[] {
  const years = new Set(matchdays.map((m) => new Date(m.played_on).getUTCFullYear()));
  return [...years].sort((a, b) => b - a);
}
