"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Card, Container, Stack, Text } from "@movoz/ui-web";
import { api } from "@/lib/api";
import type { RosterEntry } from "@/lib/types";

export default function DashboardPage() {
  const [roster, setRoster] = useState<RosterEntry[]>([]);

  useEffect(() => {
    api.get<RosterEntry[]>("/api/dashboard").then(setRoster);
  }, []);

  return (
    <Container maxWidth="md" className="py-12">
      <Stack direction="horizontal" justify="between" align="center" className="mb-6">
        <Text as="h1" font="marker" size="2xl" weight="bold">
          Dashboard
        </Text>
        <Stack direction="horizontal" gap={4}>
          <Link href="/engineers" className="text-sm text-accent hover:underline">
            Roster
          </Link>
          <Link href="/attributes" className="text-sm text-accent hover:underline">
            Attributes
          </Link>
          <Link href="/cycles" className="text-sm text-accent hover:underline">
            Cycles
          </Link>
        </Stack>
      </Stack>

      {roster.length === 0 ? (
        <Card variant="outlined">
          <Text size="sm" color="muted">
            No active engineers yet.
          </Text>
        </Card>
      ) : (
        <Card variant="outlined" padding="none">
          {roster.map((entry, i) => (
            <div
              key={entry.engineer.id}
              className={`flex items-center justify-between p-4 ${i > 0 ? "border-t border-line" : ""}`}
            >
              <Link href={`/engineers/${entry.engineer.id}`} className="font-medium text-ink hover:underline">
                {entry.engineer.name}
              </Link>
              <div className="text-right">
                <Text size="sm">{entry.latest_overall != null ? entry.latest_overall.toFixed(1) : "—"}</Text>
                <Text size="sm" color="muted">
                  {entry.last_cycle_date?.slice(0, 10) ?? "no cycles yet"}
                </Text>
              </div>
            </div>
          ))}
        </Card>
      )}
    </Container>
  );
}
