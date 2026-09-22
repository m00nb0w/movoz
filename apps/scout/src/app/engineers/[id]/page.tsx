"use client";

import { useEffect, useState, type FormEvent } from "react";
import { useParams } from "next/navigation";
import { Badge, Button, Card, Container, Dropdown, Input, Text } from "@movoz/ui-web";
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

  const selectedCycle = cycles.find((c) => c.id === selectedCycleId);
  const kindOptions = [
    { label: "Highlight", value: "highlight" },
    { label: "Lowlight", value: "lowlight" },
  ];

  return (
    <Container maxWidth="md" className="py-12">
      <Text as="h1" font="marker" size="2xl" weight="bold" className="mb-1">
        {engineer.name}
      </Text>
      <Text size="sm" color="muted" className="mb-6">
        {engineer.role}
      </Text>

      <div className="mb-6 flex items-center gap-3">
        <Text size="sm" color="muted">
          Cycle:
        </Text>
        <Dropdown
          trigger={
            <Button variant="secondary" size="sm">
              {selectedCycle ? `${selectedCycle.period_start.slice(0, 10)} — ${selectedCycle.period_end.slice(0, 10)}` : "Select cycle"}
            </Button>
          }
          items={cycles.map((c) => ({
            label: `${c.period_start.slice(0, 10)} — ${c.period_end.slice(0, 10)}`,
            value: String(c.id),
          }))}
          onSelect={(value) => setSelectedCycleId(Number(value))}
        />
      </div>

      {card && (
        <Card variant="outlined" className="mb-8">
          <Text size="lg" className="mb-3">
            Overall: <strong>{card.overall != null ? card.overall.toFixed(1) : "—"}</strong>
          </Text>
          <div className="flex flex-col gap-1">
            {card.main_attributes.map((m) => (
              <div key={m.main_attribute_id} className="flex justify-between">
                <Text size="sm">{m.name}</Text>
                <Text size="sm" color="muted">
                  {m.score.toFixed(1)}
                </Text>
              </div>
            ))}
          </div>
        </Card>
      )}

      <Card variant="outlined">
        <Text as="h2" font="marker" weight="semibold" className="mb-3">
          Overall trend
        </Text>
        {points ? (
          <svg viewBox="0 0 300 100" className="h-32 w-full">
            <polyline points={points} fill="none" stroke="currentColor" strokeWidth={2} className="text-accent" />
          </svg>
        ) : (
          <Text size="sm" color="muted">
            No scored cycles yet.
          </Text>
        )}
      </Card>

      <Card variant="outlined" className="mt-8">
        <Text as="h2" font="marker" weight="semibold" className="mb-3">
          Synced metrics
        </Text>
        {metrics.length === 0 ? (
          <Text size="sm" color="muted">
            No synced metrics yet.
          </Text>
        ) : (
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-line text-ink-soft">
                <th className="py-2">Period</th>
                <th className="py-2">PRs raised</th>
                <th className="py-2">PRs reviewed</th>
                <th className="py-2">Tickets closed</th>
                <th className="py-2">Complexity</th>
              </tr>
            </thead>
            <tbody>
              {metrics.map((m) => (
                <tr key={m.id} className="border-b border-line">
                  <td className="py-2 text-ink">
                    {m.period_start.slice(0, 10)} – {m.period_end.slice(0, 10)}
                  </td>
                  <td className="py-2 text-ink-soft">{m.prs_raised}</td>
                  <td className="py-2 text-ink-soft">{m.prs_reviewed}</td>
                  <td className="py-2 text-ink-soft">{m.tickets_closed}</td>
                  <td className="py-2 text-ink-soft">{m.complexity_score.toFixed(1)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Card>

      <Card variant="outlined" className="mt-8">
        <Text as="h2" font="marker" weight="semibold" className="mb-3">
          Highlights &amp; lowlights
        </Text>

        <form onSubmit={handleAddEntry} className="mb-4 flex flex-col gap-2">
          <div className="flex gap-2">
            <Dropdown
              trigger={<Button variant="secondary" size="sm">{newKind === "highlight" ? "Highlight" : "Lowlight"}</Button>}
              items={kindOptions}
              onSelect={(value) => setNewKind(value as "highlight" | "lowlight")}
            />
            <Input
              className="flex-1"
              size="sm"
              placeholder="What happened?"
              value={newBody}
              onChange={(e) => setNewBody(e.target.value)}
            />
            <Button type="submit" size="sm" disabled={checkingDuplicate}>
              {checkingDuplicate ? "Checking..." : "Add"}
            </Button>
          </div>
          {duplicateWarning && (
            <Card variant="outlined" padding="sm" className="border-amber-500">
              <Text size="sm" className="text-amber-700 dark:text-amber-400">
                {duplicateWarning}
              </Text>
              <button type="button" onClick={saveEntry} className="mt-1 text-sm text-ink underline">
                Save anyway
              </button>
            </Card>
          )}
        </form>

        <div className="flex flex-col gap-2">
          {highlights.map((h) => (
            <div key={h.id} className="flex items-center gap-2">
              <Badge variant="subtle" color={h.kind === "highlight" ? "success" : "danger"} size="sm">
                {h.kind}
              </Badge>
              <Text size="sm" color="muted">
                {h.created_at.slice(0, 10)}
              </Text>
              <Text size="sm">{h.body}</Text>
            </div>
          ))}
        </div>
      </Card>
    </Container>
  );
}
