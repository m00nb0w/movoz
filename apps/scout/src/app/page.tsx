"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
import type { RosterEntry } from "@/lib/types";

export default function DashboardPage() {
  const [roster, setRoster] = useState<RosterEntry[]>([]);

  useEffect(() => {
    api.get<RosterEntry[]>("/api/dashboard").then(setRoster);
  }, []);

  return (
    <main className="mx-auto max-w-3xl p-8">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-zen-text">Dashboard</h1>
        <nav className="flex gap-4 text-sm text-accent-600">
          <Link href="/engineers" className="hover:underline">
            Roster
          </Link>
          <Link href="/attributes" className="hover:underline">
            Attributes
          </Link>
          <Link href="/cycles" className="hover:underline">
            Cycles
          </Link>
        </nav>
      </div>

      <ul className="divide-y divide-zen-border">
        {roster.map((entry) => (
          <li key={entry.engineer.id} className="flex items-center justify-between py-3">
            <Link href={`/engineers/${entry.engineer.id}`} className="font-medium text-zen-text hover:underline">
              {entry.engineer.name}
            </Link>
            <div className="text-right text-sm">
              <div className="text-zen-text">{entry.latest_overall != null ? entry.latest_overall.toFixed(1) : "—"}</div>
              <div className="text-zen-muted">{entry.last_cycle_date ?? "no cycles yet"}</div>
            </div>
          </li>
        ))}
      </ul>

      {roster.length === 0 && <p className="text-sm text-zen-muted">No active engineers yet.</p>}
    </main>
  );
}
