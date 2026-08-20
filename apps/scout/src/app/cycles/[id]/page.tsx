"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import type { EngineerCycleScore } from "@/lib/types";

export default function CycleViewPage() {
  const params = useParams<{ id: string }>();
  const cycleId = Number(params.id);
  const [scores, setScores] = useState<EngineerCycleScore[]>([]);

  useEffect(() => {
    api.get<EngineerCycleScore[]>(`/api/cycles/${cycleId}/scores`).then(setScores);
  }, [cycleId]);

  // Different engineers can have different sets of main attributes scored
  // for the same cycle (e.g. an engineer added after an earlier ranking
  // round has fewer entries), so the column set is the union across all
  // rows rather than just the first row's attributes — and each cell is
  // looked up by main_attribute_id rather than assumed to line up by
  // position.
  const mainAttributeColumns = Array.from(
    new Map(
      scores.flatMap((row) => row.main_attributes.map((m) => [m.main_attribute_id, m.name] as const))
    ).entries()
  ).sort(([a], [b]) => a - b);

  return (
    <main className="mx-auto max-w-3xl p-8">
      <h1 className="mb-6 text-2xl font-semibold text-zen-text">Cycle #{cycleId} — Team Scores</h1>

      <table className="w-full text-left text-sm">
        <thead>
          <tr className="border-b border-zen-border text-zen-muted">
            <th className="py-2">Engineer</th>
            <th className="py-2">Overall</th>
            {mainAttributeColumns.map(([id, name]) => (
              <th key={id} className="py-2">
                {name}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {scores.map((row) => {
            const scoresByAttributeId = new Map(row.main_attributes.map((m) => [m.main_attribute_id, m.score]));
            return (
              <tr key={row.engineer.id} className="border-b border-zen-border">
                <td className="py-2">
                  <Link href={`/engineers/${row.engineer.id}`} className="text-zen-text hover:underline">
                    {row.engineer.name}
                  </Link>
                </td>
                <td className="py-2 text-zen-text">{row.overall != null ? row.overall.toFixed(1) : "—"}</td>
                {mainAttributeColumns.map(([id]) => {
                  const score = scoresByAttributeId.get(id);
                  return (
                    <td key={id} className="py-2 text-zen-muted">
                      {score != null ? score.toFixed(1) : "—"}
                    </td>
                  );
                })}
              </tr>
            );
          })}
        </tbody>
      </table>

      {scores.length === 0 && <p className="text-sm text-zen-muted">No rankings submitted for this cycle yet.</p>}
    </main>
  );
}
