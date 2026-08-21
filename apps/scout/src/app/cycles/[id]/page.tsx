"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import type { EngineerCycleScore, MainAttribute, SubAttribute } from "@/lib/types";

export default function CycleViewPage() {
  const params = useParams<{ id: string }>();
  const cycleId = Number(params.id);
  const [scores, setScores] = useState<EngineerCycleScore[]>([]);
  const [mainAttributes, setMainAttributes] = useState<MainAttribute[]>([]);
  const [subAttributesByMain, setSubAttributesByMain] = useState<Record<number, SubAttribute[]>>({});

  useEffect(() => {
    api.get<EngineerCycleScore[]>(`/api/cycles/${cycleId}/scores`).then(setScores);
  }, [cycleId]);

  // F6/F9 entry point: ranking a sub-attribute (and the AI assistant nested
  // under it) is the product's primary workflow, so this cycle's
  // sub-attributes have to be reachable from here — otherwise the only way in
  // is hand-typing a URL containing both a cycle id and a sub-attribute id.
  // Same grouping fetch as src/app/attributes/page.tsx: one call per main
  // attribute, `active=all` so a cycle that was ranked on a now-deactivated
  // sub-attribute still shows it.
  useEffect(() => {
    async function loadAttributes() {
      const mains = await api.get<MainAttribute[]>("/api/main-attributes");
      setMainAttributes(mains);
      const entries = await Promise.all(
        mains.map(
          async (m) =>
            [m.id, await api.get<SubAttribute[]>(`/api/sub-attributes?main_attribute_id=${m.id}&active=all`)] as const
        )
      );
      setSubAttributesByMain(Object.fromEntries(entries));
    }
    loadAttributes();
  }, []);

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

      <section className="mt-10">
        <h2 className="mb-1 text-lg font-medium text-zen-text">Rank sub-attributes</h2>
        <p className="mb-4 text-sm text-zen-muted">
          Pick a sub-attribute to rank every active engineer 1..N for this cycle — manually, or with the AI assistant
          linked from that page.
        </p>

        {mainAttributes.length === 0 && <p className="text-sm text-zen-muted">No attributes defined yet.</p>}

        <div className="space-y-4">
          {mainAttributes.map((main) => {
            const subs = subAttributesByMain[main.id] ?? [];
            return (
              <div key={main.id} className="rounded-lg border border-zen-border p-4">
                <h3 className="mb-2 font-medium text-zen-text">{main.name}</h3>
                {subs.length === 0 ? (
                  <p className="text-sm text-zen-muted">No sub-attributes yet.</p>
                ) : (
                  <ul className="divide-y divide-zen-border">
                    {subs.map((sub) => (
                      <li key={sub.id} className="py-2">
                        <Link
                          href={`/cycles/${cycleId}/sub-attributes/${sub.id}`}
                          className={`text-sm hover:underline ${sub.is_active ? "text-accent-600" : "text-zen-muted"}`}
                        >
                          {sub.name}
                          {!sub.is_active && " (inactive)"}
                        </Link>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            );
          })}
        </div>
      </section>
    </main>
  );
}
