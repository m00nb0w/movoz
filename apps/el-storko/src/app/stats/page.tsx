"use client";

import { useEffect, useState } from "react";
import { Text } from "@movoz/ui-web";
import { getBurnRate, getStatsSummary, type BurnRatePoint, type StatsSummaryResponse } from "@/lib/api";
import { StatCard } from "@/components/StatCard";
import { TrendLine } from "@/components/TrendLine";
import { StatusBreakdownChart } from "@/components/StatusBreakdownChart";

const WINDOW_DAYS = 30;

export default function StatsPage() {
  const [points, setPoints] = useState<BurnRatePoint[]>([]);
  const [summary, setSummary] = useState<StatsSummaryResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getBurnRate(WINDOW_DAYS)
      .then((res) => setPoints(res.points))
      .catch((err) => setError(err instanceof Error ? err.message : "failed to load stats"));

    getStatsSummary()
      .then((res) => setSummary(res))
      .catch((err) => setError(err instanceof Error ? err.message : "failed to load stats"));
  }, []);

  const totalCompleted = points.reduce((sum, p) => sum + p.completed, 0);
  const latestOpen = points.length > 0 ? points[points.length - 1].open : 0;
  const completionRatePercent = summary ? Math.round(summary.completion_rate * 100) : 0;

  return (
    <main className="mx-auto max-w-6xl px-4 py-8">
      <h1 className="mb-6 text-2xl font-bold text-ink">Stats</h1>

      {error && (
        <Text size="sm" color="accent" className="mb-4">
          {error}
        </Text>
      )}

      <div className="mb-4 grid grid-cols-2 gap-4 sm:grid-cols-4">
        <StatCard title="Open items">
          <Text as="p" size="lg" weight="semibold">
            {summary?.open_items ?? 0}
          </Text>
        </StatCard>

        <StatCard title="Completed this week">
          <Text as="p" size="lg" weight="semibold">
            {summary?.completed_this_week ?? 0}
          </Text>
        </StatCard>

        <StatCard title="Completion rate">
          <Text as="p" size="lg" weight="semibold">
            {completionRatePercent}%
          </Text>
        </StatCard>

        <StatCard title="Total tracked">
          <Text as="p" size="lg" weight="semibold">
            {summary?.total_tracked ?? 0}
          </Text>
        </StatCard>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <StatCard
          title="Completion throughput"
          summary={`${totalCompleted} completed in last ${points.length || WINDOW_DAYS} days`}
        >
          <TrendLine values={points.map((p) => p.completed)} color="var(--terracotta)" />
        </StatCard>

        <StatCard title="Open backlog" summary={`${latestOpen} open today`}>
          <TrendLine values={points.map((p) => p.open)} color="var(--ink)" />
        </StatCard>

        <StatCard title="Status breakdown" summary="Items by status">
          <StatusBreakdownChart breakdown={summary?.status_breakdown ?? {}} />
        </StatCard>
      </div>
    </main>
  );
}
