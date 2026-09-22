"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { Button, Card, Container, Input, Text } from "@movoz/ui-web";
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
  // Number.isInteger matters as much as the range check: a number input still
  // accepts "1.5", and a fractional rank would otherwise reach the Go backend
  // and come back as a generic permutation error. Mirrors the same check on
  // the sibling chat page (chat/page.tsx).
  const hasOutOfRangeRank = usedRanks.some((r) => !Number.isInteger(r) || r < 1 || r > engineers.length);
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
    <Container maxWidth="sm" className="py-12">
      <Link
        href={`/cycles/${cycleId}/sub-attributes/${subAttributeId}/chat`}
        className="mb-4 inline-block text-sm text-accent hover:underline"
      >
        Open AI ranking assistant →
      </Link>
      <Text as="h1" font="marker" size="2xl" weight="bold" className="mb-2">
        Rank: {subAttributeName || `Sub-attribute #${subAttributeId}`}
      </Text>
      <Text size="sm" color="muted" className="mb-6">
        Assign each active engineer a unique rank from 1 (best) to {engineers.length} (last) — no ties. Use the AI
        chat assistant (above) to get a starting proposal, then adjust here before saving.
      </Text>

      <div className="mb-6 flex flex-col gap-2">
        {engineers.map((engineer) => (
          <Card key={engineer.id} variant="outlined" padding="sm" className="flex items-center justify-between">
            <Text>{engineer.name}</Text>
            <Input
              type="number"
              min={1}
              max={engineers.length}
              value={ranks[engineer.id] ?? ""}
              onChange={(e) => setRank(engineer.id, e.target.value)}
              className="w-16 text-center"
            />
          </Card>
        ))}
      </div>

      {hasDuplicateRank && (
        <Text size="sm" color="accent" className="mb-4">
          Two engineers share the same rank — ranks must be unique 1..{engineers.length}.
        </Text>
      )}
      {hasOutOfRangeRank && !hasDuplicateRank && (
        <Text size="sm" color="accent" className="mb-4">
          Ranks must be whole numbers between 1 and {engineers.length}.
        </Text>
      )}
      {error && (
        <Text size="sm" color="accent" className="mb-4">
          {error}
        </Text>
      )}
      {saved && (
        <Text size="sm" className="mb-4 text-green-600">
          Ranking saved.
        </Text>
      )}

      <Button onClick={handleSubmit} disabled={!allRanked || hasDuplicateRank || hasOutOfRangeRank}>
        Save ranking
      </Button>
    </Container>
  );
}
