"use client";

import { useEffect, useState, type FormEvent } from "react";
import Link from "next/link";
import { Button, Card, Container, Input, Text } from "@movoz/ui-web";
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
    <Container maxWidth="md" className="py-12">
      <Text as="h1" font="marker" size="2xl" weight="bold" className="mb-6">
        Rating Cycles
      </Text>

      <Card variant="outlined" className="mb-8">
        <form onSubmit={handleCreate} className="flex flex-wrap items-end gap-3">
          <Input label="Period start" type="date" value={periodStart} onChange={(e) => setPeriodStart(e.target.value)} required />
          <Input label="Period end" type="date" value={periodEnd} onChange={(e) => setPeriodEnd(e.target.value)} required />
          <Button type="submit">Open cycle</Button>
        </form>
        {error && (
          <Text size="sm" color="accent" className="mt-3">
            {error}
          </Text>
        )}
      </Card>

      <Card variant="outlined" padding="none">
        {cycles.length === 0 ? (
          <Text size="sm" color="muted" className="p-4">
            No cycles yet.
          </Text>
        ) : (
          cycles.map((cycle, i) => (
            <div
              key={cycle.id}
              className={`flex items-center justify-between p-4 ${i > 0 ? "border-t border-line" : ""}`}
            >
              <Text size="base">
                {cycle.period_start.slice(0, 10)} — {cycle.period_end.slice(0, 10)}
              </Text>
              <Link href={`/cycles/${cycle.id}`} className="text-sm text-accent hover:underline">
                View scores
              </Link>
            </div>
          ))
        )}
      </Card>
    </Container>
  );
}
