"use client";

import { useState, useMemo } from "react";
import { Card, Container, Input, Pill, Text } from "@movoz/ui-web";
import { TopBar } from "@/components/TopBar";
import { Icon } from "@/components/icons";
import { DIGEST_ARCHIVE } from "@/data/mock";

// Deterministic last-30-day intensity values (avoid SSR/CSR mismatch from Math.random).
const HEATMAP_DAYS = [
  0.42, 0.18, 0.65, 0.81, 0.33, 0.55, 0.27, 0.92, 0.41, 0.36,
  0.71, 0.13, 0.58, 0.49, 0.84, 0.22, 0.66, 0.30, 0.51, 0.74,
  0.39, 0.62, 0.18, 0.46, 0.88, 0.33, 0.57, 0.41, 0.79, 0.66,
];

const SEARCH_SUGGESTIONS = ["oauth", "fido2", "sonnet", "sbv", "metro line 2"];

export default function ArchivePage() {
  const [search, setSearch] = useState("");

  const filtered = useMemo(
    () =>
      DIGEST_ARCHIVE.filter(
        (d) =>
          !search ||
          d.summary.toLowerCase().includes(search.toLowerCase()) ||
          d.date.toLowerCase().includes(search.toLowerCase()),
      ),
    [search],
  );

  return (
    <div>
      <TopBar />

      <Container maxWidth="md" className="py-9 pb-16">
        {/* Header */}
        <div className="mb-7">
          <div className="font-mono-tabular mb-1.5 text-[10px] uppercase tracking-[0.18em] text-ink-soft">
            The Morning Digest · Archive
          </div>
          <Text as="h1" font="marker" size="4xl">
            All <em className="italic">187</em> mornings.
          </Text>
        </div>

        {/* Search + heatmap */}
        <div className="mb-8 grid grid-cols-[1fr_360px] gap-6">
          <Card variant="outlined" padding="sm">
            <div className="font-mono-tabular mb-2.5 text-[10px] uppercase tracking-widest text-ink-soft">
              Full-text search
            </div>
            <Input
              size="sm"
              icon={<Icon.search />}
              placeholder='e.g. "OAuth resource indicators"'
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              helperText={`${filtered.length} matches`}
            />
            <div className="mt-3.5 flex flex-wrap gap-1.5">
              {SEARCH_SUGGESTIONS.map((s) => (
                <Pill key={s} onClick={() => setSearch(s)} className="text-sm">
                  {s}
                </Pill>
              ))}
            </div>
          </Card>

          <Card variant="outlined" padding="sm">
            <div className="font-mono-tabular mb-2.5 text-[10px] uppercase tracking-widest text-ink-soft">
              Last 30 days · item density
            </div>
            <div className="grid grid-cols-[repeat(15,1fr)] gap-1">
              {HEATMAP_DAYS.map((intensity, i) => {
                const items = Math.floor(intensity * 18);
                return (
                  <div
                    key={i}
                    title={`${items} items`}
                    className="aspect-square rounded-[2px] border border-line"
                    style={{ background: `rgba(37,35,30,${intensity * 0.6 + 0.05})` }}
                  />
                );
              })}
            </div>
            <div className="font-mono-tabular mt-2.5 flex justify-between text-[10px] tracking-wide text-ink-soft">
              <span>30d ago</span>
              <span>today</span>
            </div>
          </Card>
        </div>

        {/* Day cards */}
        <div className="font-mono-tabular mb-3.5 text-[10px] uppercase tracking-widest text-ink-soft">
          Recent
        </div>
        <div className="grid grid-cols-2 gap-px overflow-hidden rounded border border-line bg-line">
          {filtered.map((d) => (
            <div key={d.date} className="cursor-pointer bg-paper px-5.5 py-5">
              <div className="mb-2 flex items-baseline justify-between">
                <Text as="h3" font="marker" size="xl" className="italic">
                  {d.date}
                </Text>
                <span className="font-mono-tabular text-[10px] tracking-widest text-ink-soft">
                  {d.items} items
                </span>
              </div>
              <div className="mb-3 text-[13px] leading-relaxed text-ink-soft">{d.summary}</div>
              <div className="flex gap-1.5">
                {d.topics.map((t) => (
                  <span
                    key={t}
                    className="font-mono-tabular rounded-[2px] border border-line px-1.5 py-0.5 text-[9px] uppercase tracking-widest text-ink-soft"
                  >
                    {t}
                  </span>
                ))}
              </div>
            </div>
          ))}
        </div>
      </Container>
    </div>
  );
}
