"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { api } from "@/lib/api";
import type { Engineer, MainAttribute, SubAttribute, SubAttributeRanking } from "@/lib/types";

export default function RankSubAttributePage() {
  const params = useParams<{ id: string; subId: string }>();
  const cycleId = Number(params.id);
  const subAttributeId = Number(params.subId);

  const [engineers, setEngineers] = useState<Engineer[]>([]);
  const [subAttributeName, setSubAttributeName] = useState("");
  const [ranks, setRanks] = useState<Record<number, number>>({});
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    async function load() {
      const activeEngineers = await api.get<Engineer[]>("/api/engineers");
      setEngineers(activeEngineers);

      const mains = await api.get<MainAttribute[]>("/api/main-attributes");
      for (const main of mains) {
        const subs = await api.get<SubAttribute[]>(`/api/sub-attributes?main_attribute_id=${main.id}&active=all`);
        const match = subs.find((s) => s.id === subAttributeId);
        if (match) {
          setSubAttributeName(match.name);
          break;
        }
      }

      const existing = await api.get<SubAttributeRanking[]>(
        `/api/cycles/${cycleId}/sub-attributes/${subAttributeId}/ranking`
      );
      const initialRanks: Record<number, number> = {};
      existing.forEach((r) => {
        initialRanks[r.engineer_id] = r.rank;
      });
      setRanks(initialRanks);
    }
    load();
  }, [cycleId, subAttributeId]);

  const usedRanks = useMemo(() => Object.values(ranks), [ranks]);
  const hasDuplicateRank = new Set(usedRanks).size !== usedRanks.length;
  const hasOutOfRangeRank = usedRanks.some((r) => r < 1 || r > engineers.length);
  const allRanked = engineers.length > 0 && engineers.every((e) => ranks[e.id] != null);

  function setRank(engineerId: number, rawValue: string) {
    setSaved(false);
    if (rawValue === "") {
      // Clearing the field means "not yet ranked" — drop the entry rather
      // than coercing to 0, which would otherwise silently pass allRanked
      // and produce an out-of-range rank in the submitted payload.
      setRanks((prev) => {
        const next = { ...prev };
        delete next[engineerId];
        return next;
      });
      return;
    }
    setRanks((prev) => ({ ...prev, [engineerId]: Number(rawValue) }));
  }

  async function handleSubmit() {
    setError(null);
    setSaved(false);
    try {
      const rankings = engineers.map((e) => ({ engineer_id: e.id, rank: ranks[e.id] }));
      await api.put(`/api/cycles/${cycleId}/sub-attributes/${subAttributeId}/ranking`, { rankings });
      setSaved(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save ranking");
    }
  }

  return (
    <main className="mx-auto max-w-2xl p-8">
      <h1 className="mb-2 text-2xl font-semibold text-zen-text">
        Rank: {subAttributeName || `Sub-attribute #${subAttributeId}`}
      </h1>
      <p className="mb-6 text-sm text-zen-muted">
        Assign each active engineer a unique rank from 1 (best) to {engineers.length} (last) — no ties. Use the AI
        chat assistant (below, once built) to get a starting proposal, then adjust here before saving.
      </p>

      <ul className="mb-6 space-y-2">
        {engineers.map((engineer) => (
          <li key={engineer.id} className="flex items-center justify-between rounded border border-zen-border p-3">
            <span className="text-zen-text">{engineer.name}</span>
            <input
              type="number"
              min={1}
              max={engineers.length}
              value={ranks[engineer.id] ?? ""}
              onChange={(e) => setRank(engineer.id, e.target.value)}
              className="w-16 rounded border border-zen-border bg-transparent p-1 text-center"
            />
          </li>
        ))}
      </ul>

      {hasDuplicateRank && (
        <p className="mb-4 text-sm text-red-500">
          Two engineers share the same rank — ranks must be unique 1..{engineers.length}.
        </p>
      )}
      {hasOutOfRangeRank && !hasDuplicateRank && (
        <p className="mb-4 text-sm text-red-500">
          Ranks must be between 1 and {engineers.length}.
        </p>
      )}
      {error && <p className="mb-4 text-sm text-red-500">{error}</p>}
      {saved && <p className="mb-4 text-sm text-green-600">Ranking saved.</p>}

      <button
        onClick={handleSubmit}
        disabled={!allRanked || hasDuplicateRank || hasOutOfRangeRank}
        className="rounded bg-accent-600 px-4 py-2 text-white disabled:opacity-50"
      >
        Save ranking
      </button>
    </main>
  );
}
