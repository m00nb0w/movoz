"use client";

import { useEffect, useState, type FormEvent } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
import type { RatingCycle } from "@/lib/types";

export default function CyclesPage() {
  const [cycles, setCycles] = useState<RatingCycle[]>([]);
  const [periodStart, setPeriodStart] = useState("");
  const [periodEnd, setPeriodEnd] = useState("");
  const [error, setError] = useState<string | null>(null);

  async function load() {
    setCycles(await api.get<RatingCycle[]>("/api/cycles"));
  }

  useEffect(() => {
    load();
  }, []);

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await api.post("/api/cycles", { period_start: periodStart, period_end: periodEnd });
      setPeriodStart("");
      setPeriodEnd("");
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create cycle");
    }
  }

  return (
    <main className="mx-auto max-w-3xl p-8">
      <h1 className="mb-6 text-2xl font-semibold text-zen-text">Rating Cycles</h1>

      <form onSubmit={handleCreate} className="mb-8 flex gap-3 rounded-lg border border-zen-border p-4">
        <input className="rounded border border-zen-border bg-transparent p-2" type="date" value={periodStart} onChange={(e) => setPeriodStart(e.target.value)} required />
        <input className="rounded border border-zen-border bg-transparent p-2" type="date" value={periodEnd} onChange={(e) => setPeriodEnd(e.target.value)} required />
        <button type="submit" className="rounded bg-accent-600 px-3 py-2 text-white">Open cycle</button>
      </form>
      {error && <p className="mb-4 text-sm text-red-500">{error}</p>}

      <ul className="divide-y divide-zen-border">
        {cycles.map((cycle) => (
          <li key={cycle.id} className="flex items-center justify-between py-3">
            <span className="text-zen-text">
              {cycle.period_start.slice(0, 10)} — {cycle.period_end.slice(0, 10)}
            </span>
            <Link href={`/cycles/${cycle.id}`} className="text-sm text-accent-600 hover:underline">
              View scores
            </Link>
          </li>
        ))}
      </ul>
    </main>
  );
}
