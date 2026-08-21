"use client";

import { useEffect, useState, type FormEvent } from "react";
import { useParams } from "next/navigation";
import { api } from "@/lib/api";
import type {
  Engineer,
  EngineerCard as EngineerCardData,
  HighlightEntry,
  MetricSnapshot,
  RatingCycle,
  TrendPoint,
} from "@/lib/types";

export default function EngineerCardPage() {
  const params = useParams<{ id: string }>();
  const engineerId = Number(params.id);

  const [engineer, setEngineer] = useState<Engineer | null>(null);
  const [cycles, setCycles] = useState<RatingCycle[]>([]);
  const [selectedCycleId, setSelectedCycleId] = useState<number | null>(null);
  const [card, setCard] = useState<EngineerCardData | null>(null);
  const [trend, setTrend] = useState<TrendPoint[]>([]);
  const [metrics, setMetrics] = useState<MetricSnapshot[]>([]);
  const [highlights, setHighlights] = useState<HighlightEntry[]>([]);
  const [newKind, setNewKind] = useState<"highlight" | "lowlight">("highlight");
  const [newBody, setNewBody] = useState("");
  const [duplicateWarning, setDuplicateWarning] = useState<string | null>(null);
  const [checkingDuplicate, setCheckingDuplicate] = useState(false);

  useEffect(() => {
    async function loadStatic() {
      const [eng, cycleList, trendData, metricSnapshots] = await Promise.all([
        api.get<Engineer>(`/api/engineers/${engineerId}`),
        api.get<RatingCycle[]>("/api/cycles"),
        api.get<TrendPoint[]>(`/api/engineers/${engineerId}/trend`),
        api.get<MetricSnapshot[]>(`/api/engineers/${engineerId}/metrics`),
      ]);
      setEngineer(eng);
      setCycles(cycleList);
      setTrend(trendData);
      setMetrics(metricSnapshots);
      if (cycleList.length > 0) setSelectedCycleId(cycleList[0].id);
    }
    loadStatic();
  }, [engineerId]);

  useEffect(() => {
    if (selectedCycleId == null) return;
    api.get<EngineerCardData>(`/api/engineers/${engineerId}/card?cycleId=${selectedCycleId}`).then(setCard);
  }, [engineerId, selectedCycleId]);

  useEffect(() => {
    api.get<HighlightEntry[]>(`/api/engineers/${engineerId}/highlights`).then(setHighlights);
  }, [engineerId]);

  async function saveEntry() {
    await api.post(`/api/engineers/${engineerId}/highlights`, { kind: newKind, body: newBody });
    setNewBody("");
    setDuplicateWarning(null);
    const updated = await api.get<HighlightEntry[]>(`/api/engineers/${engineerId}/highlights`);
    setHighlights(updated);
  }

  async function handleAddEntry(e: FormEvent) {
    e.preventDefault();
    if (!newBody.trim()) return;

    setCheckingDuplicate(true);
    setDuplicateWarning(null);

    // F14/NF3: the duplicate flag must never block the save. The backend
    // already degrades gracefully (200, is_duplicate=false) when the AI
    // call itself fails, but a genuine client-side/network failure calling
    // this endpoint would otherwise throw here — catch that too so the save
    // still proceeds without a flag rather than getting stuck.
    try {
      const check = await api.post<{ is_duplicate: boolean; matched_entry_id: number | null; similarity_note: string }>(
        `/api/engineers/${engineerId}/highlights/check-duplicate`,
        { body: newBody }
      );
      setCheckingDuplicate(false);

      if (check.is_duplicate) {
        setDuplicateWarning(`Possible duplicate: ${check.similarity_note}`);
        return;
      }
    } catch {
      setCheckingDuplicate(false);
    }

    await saveEntry();
  }

  if (!engineer) return null;

  const scoredPoints = trend.filter((t) => t.overall != null);
  const points = scoredPoints
    .map((t, i) => {
      const x = scoredPoints.length > 1 ? (i / (scoredPoints.length - 1)) * 300 : 150;
      const y = 100 - (((t.overall as number) - 50) / 50) * 100;
      return `${x},${y}`;
    })
    .join(" ");

  return (
    <main className="mx-auto max-w-3xl p-8">
      <h1 className="mb-1 text-2xl font-semibold text-zen-text">{engineer.name}</h1>
      <p className="mb-6 text-sm text-zen-muted">{engineer.role}</p>

      <div className="mb-6 flex items-center gap-3">
        <label className="text-sm text-zen-muted">Cycle:</label>
        <select
          value={selectedCycleId ?? ""}
          onChange={(e) => setSelectedCycleId(Number(e.target.value))}
          className="rounded border border-zen-border bg-transparent p-2"
        >
          {cycles.map((c) => (
            <option key={c.id} value={c.id}>
              {c.period_start} — {c.period_end}
            </option>
          ))}
        </select>
      </div>

      {card && (
        <section className="mb-8 rounded-lg border border-zen-border p-4">
          <p className="mb-3 text-lg text-zen-text">
            Overall: <strong>{card.overall != null ? card.overall.toFixed(1) : "—"}</strong>
          </p>
          <ul className="space-y-1">
            {card.main_attributes.map((m) => (
              <li key={m.main_attribute_id} className="flex justify-between text-sm">
                <span className="text-zen-text">{m.name}</span>
                <span className="text-zen-muted">{m.score.toFixed(1)}</span>
              </li>
            ))}
          </ul>
        </section>
      )}

      <section className="rounded-lg border border-zen-border p-4">
        <h2 className="mb-3 font-medium text-zen-text">Overall trend</h2>
        {points ? (
          <svg viewBox="0 0 300 100" className="h-32 w-full">
            <polyline points={points} fill="none" stroke="currentColor" strokeWidth={2} className="text-accent-600" />
          </svg>
        ) : (
          <p className="text-sm text-zen-muted">No scored cycles yet.</p>
        )}
      </section>

      <section className="mt-8 rounded-lg border border-zen-border p-4">
        <h2 className="mb-3 font-medium text-zen-text">Synced metrics</h2>
        {metrics.length === 0 ? (
          <p className="text-sm text-zen-muted">No synced metrics yet.</p>
        ) : (
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-zen-border text-zen-muted">
                <th className="py-2">Period</th>
                <th className="py-2">PRs raised</th>
                <th className="py-2">PRs reviewed</th>
                <th className="py-2">Tickets closed</th>
                <th className="py-2">Complexity</th>
              </tr>
            </thead>
            <tbody>
              {metrics.map((m) => (
                <tr key={m.id} className="border-b border-zen-border">
                  <td className="py-2 text-zen-text">
                    {m.period_start} – {m.period_end}
                  </td>
                  <td className="py-2 text-zen-muted">{m.prs_raised}</td>
                  <td className="py-2 text-zen-muted">{m.prs_reviewed}</td>
                  <td className="py-2 text-zen-muted">{m.tickets_closed}</td>
                  <td className="py-2 text-zen-muted">{m.complexity_score.toFixed(1)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      <section className="mt-8 rounded-lg border border-zen-border p-4">
        <h2 className="mb-3 font-medium text-zen-text">Highlights &amp; lowlights</h2>

        <form onSubmit={handleAddEntry} className="mb-4 space-y-2">
          <div className="flex gap-2">
            <select
              value={newKind}
              onChange={(e) => setNewKind(e.target.value as "highlight" | "lowlight")}
              className="rounded border border-zen-border bg-transparent p-2 text-sm"
            >
              <option value="highlight">Highlight</option>
              <option value="lowlight">Lowlight</option>
            </select>
            <input
              className="flex-1 rounded border border-zen-border bg-transparent p-2 text-sm"
              placeholder="What happened?"
              value={newBody}
              onChange={(e) => setNewBody(e.target.value)}
            />
            <button type="submit" disabled={checkingDuplicate} className="rounded bg-accent-600 px-3 py-2 text-sm text-white disabled:opacity-50">
              {checkingDuplicate ? "Checking..." : "Add"}
            </button>
          </div>
          {duplicateWarning && (
            <div className="rounded border border-yellow-500 p-2 text-sm text-yellow-700">
              <p>{duplicateWarning}</p>
              <button type="button" onClick={saveEntry} className="mt-1 underline">
                Save anyway
              </button>
            </div>
          )}
        </form>

        <ul className="space-y-2">
          {highlights.map((h) => (
            <li key={h.id} className="text-sm">
              <span className={h.kind === "highlight" ? "text-green-600" : "text-red-500"}>
                {h.kind === "highlight" ? "★" : "▼"}
              </span>{" "}
              <span className="text-zen-muted">{h.created_at.slice(0, 10)}</span>{" "}
              <span className="text-zen-text">{h.body}</span>
            </li>
          ))}
        </ul>
      </section>
    </main>
  );
}
