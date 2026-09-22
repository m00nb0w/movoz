"use client";

import { useMemo, useState } from "react";
import { Button, Card, Container, IconButton, Input, Pill, Text } from "@movoz/ui-web";
import { TopBar } from "@/components/TopBar";
import { Icon } from "@/components/icons";
import { Sparkline } from "@/components/Sparkline";
import { Donut } from "@/components/Donut";
import { HBar } from "@/components/HBar";
import { BudgetGauge } from "@/components/BudgetGauge";
import { BucketDot } from "@/components/BucketChip";
import { fmtVND } from "@/lib/format";
import {
  CATEGORIES,
  MONTH_BUDGET,
  MONTH_BY_BUCKET,
  MONTH_DAYS_IN,
  MONTH_DAYS_TOTAL,
  MONTH_SPENT,
  RECENT_EXPENSES,
  TAGS,
  WEEKLY_TREND,
  type Bucket,
} from "@/data/mock";
import { AddExpenseModal } from "./AddExpenseModal";

type BucketFilter = Bucket | "all";
type RangeFilter = "7d" | "30d" | "90d" | "YTD";

const BUCKET_FILTERS: { value: BucketFilter; label: string }[] = [
  { value: "all", label: "All" },
  { value: "essentials", label: "Essentials" },
  { value: "lifestyle", label: "Lifestyle" },
  { value: "irregular", label: "Irregular" },
];

const RANGE_FILTERS: RangeFilter[] = ["7d", "30d", "90d", "YTD"];

export default function ExpensesPage() {
  const [filterBucket, setFilterBucket] = useState<BucketFilter>("all");
  const [filterRange, setFilterRange] = useState<RangeFilter>("30d");
  const [search, setSearch] = useState("");
  const [showAdd, setShowAdd] = useState(false);
  const [parseInput, setParseInput] = useState("lunch 120k");
  const [toast, setToast] = useState<{ amount: number } | null>(null);

  const filtered = useMemo(
    () =>
      RECENT_EXPENSES.filter((e) => {
        if (filterBucket !== "all" && e.bucket !== filterBucket) return false;
        const term = search.toLowerCase();
        if (!term) return true;
        return (
          e.text.toLowerCase().includes(term) ||
          (e.tag ?? "").toLowerCase().includes(term)
        );
      }),
    [filterBucket, search],
  );

  const filterTotal = filtered.reduce((s, e) => s + e.amount, 0);

  const expectedSpend = (MONTH_BUDGET * MONTH_DAYS_IN) / MONTH_DAYS_TOTAL;
  const aheadBy = MONTH_SPENT - expectedSpend;
  const monthTotal = MONTH_BY_BUCKET.reduce((s, b) => s + b.amount, 0);
  const lastWeek = WEEKLY_TREND[WEEKLY_TREND.length - 1];

  return (
    <div>
      <TopBar />

      {toast && (
        <div className="fixed top-20 right-8 z-[100] flex items-center gap-2.5 rounded bg-ink px-4.5 py-3 text-[13px] text-paper-raised shadow-lg">
          <span className="text-lg">🐬</span> Logged.{" "}
          <span className="font-mono-tabular text-[11px] opacity-70">
            {fmtVND(toast.amount)}₫
          </span>
        </div>
      )}

      <Container maxWidth="2xl" className="py-8 pb-16">
        {/* Header */}
        <div className="mb-7 flex items-end justify-between">
          <div>
            <div className="font-mono-tabular mb-1.5 text-[10px] uppercase tracking-widest text-ink-soft">
              Ledger · May 2026
            </div>
            <Text as="h1" font="marker" size="4xl" className="leading-none">
              <em className="italic">Where</em> the money goes.
            </Text>
          </div>
          <Button variant="primary" onClick={() => setShowAdd(true)}>
            <Icon.plus style={{ verticalAlign: "middle", marginRight: 6 }} /> Add expense
          </Button>
        </div>

        {/* Top metrics row */}
        <div className="mb-6 grid grid-cols-[1.4fr_1fr_1fr] gap-4.5">
          <Card variant="outlined" padding="lg" className="flex items-center gap-6">
            <BudgetGauge
              spent={MONTH_SPENT}
              budget={MONTH_BUDGET}
              daysIn={MONTH_DAYS_IN}
              daysTotal={MONTH_DAYS_TOTAL}
              size={200}
            />
            <div className="flex-1">
              <div className="font-mono-tabular mb-2 text-[10px] uppercase tracking-widest text-ink-soft">
                Pacing
              </div>
              <div className="mb-3 text-sm leading-relaxed text-ink-soft">
                You&apos;ve spent <strong className="text-ink">{fmtVND(MONTH_SPENT)}₫</strong> of{" "}
                <strong>{fmtVND(MONTH_BUDGET)}₫</strong> in {MONTH_DAYS_IN} days. Expected pace
                would be <span className="font-mono-tabular font-medium">{fmtVND(expectedSpend)}₫</span>.
              </div>
              <div className="text-xs text-accent">
                <Icon.arrowUp style={{ verticalAlign: "middle" }} /> Running ~{fmtVND(aheadBy)}₫ ahead.
                Soft slow-down recommended.
              </div>
            </div>
          </Card>

          <Card variant="outlined" padding="lg">
            <div className="font-mono-tabular mb-2.5 text-[10px] uppercase tracking-widest text-ink-soft">
              By bucket · MTD
            </div>
            <div className="flex items-center gap-4.5">
              <Donut
                segments={MONTH_BY_BUCKET.map((b) => ({
                  value: b.amount,
                  color: `var(--bucket-${b.bucket})`,
                }))}
                size={130}
                thickness={18}
                label={fmtVND(monthTotal)}
                sub="₫ TOTAL"
              />
              <div className="flex flex-1 flex-col gap-1.5">
                {MONTH_BY_BUCKET.map((b) => (
                  <div key={b.bucket} className="flex items-center gap-2 text-xs">
                    <BucketDot bucket={b.bucket} />
                    <span className="flex-1 text-ink-soft">{b.label}</span>
                    <span className="font-mono-tabular font-medium text-ink">
                      {Math.round(b.share * 100)}%
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </Card>

          <Card variant="outlined" padding="lg">
            <div className="font-mono-tabular mb-1.5 text-[10px] uppercase tracking-widest text-ink-soft">
              12-week trend
            </div>
            <div className="mb-1.5 flex items-baseline gap-2">
              <Text as="span" font="marker" size="2xl">
                {fmtVND(lastWeek)}
                <span className="text-lg text-ink-soft">₫</span>
              </Text>
              <span className="font-mono-tabular text-[10px] text-bucket-essentials">
                <Icon.arrowDown style={{ verticalAlign: "middle" }} /> 14% w/w
              </span>
            </div>
            <Sparkline
              data={WEEKLY_TREND}
              width={260}
              height={60}
              stroke="var(--ink)"
              fill="var(--line)"
              marker={WEEKLY_TREND.length - 1}
            />
            <div className="font-mono-tabular mt-1.5 flex justify-between text-[10px] text-ink-soft">
              <span>12w ago</span>
              <span>this week</span>
            </div>
          </Card>
        </div>

        {/* Category + tags */}
        <div className="mb-6 grid grid-cols-[1.6fr_1fr] gap-4.5">
          <Card variant="outlined" padding="lg">
            <div className="mb-4 flex items-baseline justify-between">
              <Text as="div" font="marker" size="lg">
                By category
              </Text>
              <span className="font-mono-tabular text-[10px] uppercase tracking-widest text-ink-soft">
                Sorted descending
              </span>
            </div>
            <HBar
              items={CATEGORIES.slice()
                .sort((a, b) => b.amount - a.amount)
                .map((c) => ({
                  label: c.name,
                  value: c.amount,
                  color: `var(--bucket-${c.bucket})`,
                }))}
              valueFmt={(v) => `${fmtVND(v)}₫`}
            />
          </Card>

          <Card variant="outlined" padding="lg">
            <div className="mb-4 flex items-baseline justify-between">
              <Text as="div" font="marker" size="lg">
                Active tags
              </Text>
              <span className="font-mono-tabular text-[10px] uppercase tracking-widest text-ink-soft">
                Trip & project
              </span>
            </div>
            <div className="flex flex-col gap-3.5">
              {TAGS.map((t) => (
                <div key={t.name} className="rounded border border-line bg-paper px-3.5 py-3">
                  <div className="mb-1 flex items-baseline justify-between">
                    <span className="font-mono-tabular text-[13px] text-ink">{t.name}</span>
                    <span className="font-mono-tabular text-[15px] font-medium text-ink">
                      {fmtVND(t.total)}
                      <span className="text-ink-soft">₫</span>
                    </span>
                  </div>
                  <div className="font-mono-tabular text-[10px] text-ink-soft">
                    {t.count} entries · {t.period}
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </div>

        {/* Filterable ledger */}
        <Card variant="outlined" padding="none">
          <div className="flex flex-wrap items-center gap-3 border-b border-line px-5.5 py-5">
            <Text as="div" font="marker" size="lg" className="mr-auto">
              Ledger
            </Text>
            <Input
              size="sm"
              icon={<Icon.search />}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search notes, tags…"
              className="min-w-[180px]"
            />
            <div className="flex gap-1">
              {BUCKET_FILTERS.map((b) => (
                <Pill key={b.value} active={filterBucket === b.value} onClick={() => setFilterBucket(b.value)}>
                  {b.label}
                </Pill>
              ))}
            </div>
            <div className="ml-2 flex gap-1">
              {RANGE_FILTERS.map((r) => (
                <Pill key={r} active={filterRange === r} onClick={() => setFilterRange(r)}>
                  {r}
                </Pill>
              ))}
            </div>
          </div>

          <div>
            <div className="font-mono-tabular grid grid-cols-[70px_60px_24px_1fr_110px_120px_100px_50px] items-center bg-paper px-5.5 py-2.5 text-[10px] uppercase tracking-widest text-ink-soft [border-bottom:1px_solid_var(--line)]">
              <span>Date</span>
              <span>Time</span>
              <span></span>
              <span>Note</span>
              <span>Category</span>
              <span>Tag</span>
              <span className="text-right">Amount</span>
              <span></span>
            </div>
            {filtered.map((e, i) => (
              <div
                key={`${e.date}-${e.time}-${i}`}
                className="grid grid-cols-[70px_60px_24px_1fr_110px_120px_100px_50px] items-center px-5.5 py-2.5 text-[13px] text-ink-soft"
                style={{ borderBottom: i === filtered.length - 1 ? "none" : "1px solid var(--line)" }}
              >
                <span className="font-mono-tabular text-[11px] text-ink-soft">{e.date}</span>
                <span className="font-mono-tabular text-[11px] text-ink-soft">{e.time}</span>
                <BucketDot bucket={e.bucket} size={7} />
                <span className="text-ink">{e.text}</span>
                <span className="font-mono-tabular text-[11px] uppercase tracking-wide text-ink-soft">
                  {e.category}
                </span>
                <span
                  className="font-mono-tabular text-[11px]"
                  style={{ color: e.tag ? "var(--bucket-irregular)" : "var(--ink-soft)" }}
                >
                  {e.tag || "—"}
                </span>
                <span className="font-mono-tabular text-right text-sm font-medium text-ink">
                  {fmtVND(e.amount)}
                  <span className="text-[11px] text-ink-soft"> ₫</span>
                </span>
                <span className="flex justify-end opacity-50">
                  <IconButton
                    variant="ghost"
                    size="sm"
                    label="Edit"
                    icon={
                      <svg width="13" height="13" viewBox="0 0 16 16" fill="none">
                        <path
                          d="M3 13l1-3 7-7 2 2-7 7-3 1zM10 4l2 2"
                          stroke="currentColor"
                          strokeWidth="1.3"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                        />
                      </svg>
                    }
                  />
                </span>
              </div>
            ))}
            <div className="flex items-center justify-between bg-paper px-5.5 py-3.5 [border-top:2px_solid_var(--ink)]">
              <span className="font-mono-tabular text-[11px] uppercase tracking-widest text-ink-soft">
                {filtered.length} entries · filtered total
              </span>
              <span className="font-mono-tabular text-lg font-medium text-ink">
                {fmtVND(filterTotal)}
                <span className="text-[13px] text-ink-soft"> ₫</span>
              </span>
            </div>
          </div>
        </Card>
      </Container>

      {showAdd && (
        <AddExpenseModal
          parseInput={parseInput}
          onParseInputChange={setParseInput}
          onCancel={() => setShowAdd(false)}
          onSave={(parsed) => {
            if (!parsed) return;
            setShowAdd(false);
            setToast({ amount: parsed.amount });
            setTimeout(() => setToast(null), 3000);
          }}
        />
      )}
    </div>
  );
}
