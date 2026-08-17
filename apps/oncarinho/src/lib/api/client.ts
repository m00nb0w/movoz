import type {
  Player,
  PlayerProfile,
  LeaderboardEntry,
  Summary,
  Matchday,
  MatchStat,
  StatInput,
  Position,
} from "./types";

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message);
  }
}

function apiBase(): string {
  if (typeof window === "undefined") {
    return process.env.ONCARINHO_API_URL || "http://localhost:8081";
  }
  return "";
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${apiBase()}/api${path}`, {
    ...init,
    credentials: "include",
    headers: { "Content-Type": "application/json", ...init?.headers },
    cache: "no-store",
  });

  if (res.status === 401 && typeof window !== "undefined" && !path.startsWith("/auth/login")) {
    window.location.href = "/admin";
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiError(res.status, body.error || res.statusText);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}

export const api = {
  getPlayers: (includeInactive = false) =>
    request<Player[]>(`/players${includeInactive ? "?active=all" : ""}`),
  getPlayerProfile: (id: number) => request<PlayerProfile>(`/players/${id}`),
  getLeaderboard: (year: number | "all", stat: "goals" | "assists" | "cards") =>
    request<LeaderboardEntry[]>(`/leaderboard?year=${year}&stat=${stat}`),
  getSummary: (year?: number) => request<Summary>(`/summary${year ? `?year=${year}` : ""}`),
  getMatchdays: (year?: number) => request<Matchday[]>(`/matchdays${year ? `?year=${year}` : ""}`),
  getMatchdayStats: (matchdayId: number) => request<MatchStat[]>(`/matchdays/${matchdayId}/stats`),

  login: (password: string) =>
    request<{ status: string }>(`/auth/login`, {
      method: "POST",
      body: JSON.stringify({ password }),
    }),
  createPlayer: (name: string, position: Position | null) =>
    request<Player>(`/players`, { method: "POST", body: JSON.stringify({ name, position }) }),
  updatePlayer: (id: number, name: string, position: Position | null) =>
    request<Player>(`/players/${id}`, {
      method: "PUT",
      body: JSON.stringify({ name, position }),
    }),
  deactivatePlayer: (id: number) => request<void>(`/players/${id}`, { method: "DELETE" }),
  reactivatePlayer: (id: number) =>
    request<void>(`/players/${id}/reactivate`, { method: "POST" }),
  createMatchday: (playedOn: string) =>
    request<Matchday>(`/matchdays`, {
      method: "POST",
      body: JSON.stringify({ played_on: playedOn }),
    }),
  upsertMatchdayStats: (matchdayId: number, entries: StatInput[]) =>
    request<MatchStat[]>(`/matchdays/${matchdayId}/stats`, {
      method: "PUT",
      body: JSON.stringify({ entries }),
    }),
  deleteMatchdayStat: (matchdayId: number, playerId: number) =>
    request<void>(`/matchdays/${matchdayId}/stats/${playerId}`, { method: "DELETE" }),
};
