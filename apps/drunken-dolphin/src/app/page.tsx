"use client";

import { useState } from "react";
import { Button, Card, Container, Divider, IconButton, Pill, Text } from "@movoz/ui-web";
import { TopBar } from "@/components/TopBar";
import { Icon } from "@/components/icons";
import { Sparkline } from "@/components/Sparkline";
import { BudgetGauge } from "@/components/BudgetGauge";
import { BucketDot } from "@/components/BucketChip";
import { fmtVND } from "@/lib/format";
import {
  TODAY,
  DIGEST_TODAY,
  TODAY_EXPENSES,
  MONTH_BY_BUCKET,
  MONTH_BUDGET,
  MONTH_SPENT,
  MONTH_DAYS_IN,
  MONTH_DAYS_TOTAL,
  DAILY_SPEND,
  type Reaction,
} from "@/data/mock";

const YESTERDAY_TOTAL = 4_875_000;

export default function DashboardPage() {
  const [reactions, setReactions] = useState<Record<string, Reaction>>({});
  const [savedItems, setSavedItems] = useState<Set<string>>(new Set());
  const [dolphinSeen, setDolphinSeen] = useState(false);

  const setReact = (key: string, val: Exclude<Reaction, null>) =>
    setReactions((r) => ({ ...r, [key]: r[key] === val ? null : val }));

  const toggleSave = (k: string) =>
    setSavedItems((s) => {
      const n = new Set(s);
      if (n.has(k)) n.delete(k);
      else n.add(k);
      return n;
    });

  const featuredItems = [
    DIGEST_TODAY.topics[0].items[0],
    DIGEST_TODAY.topics[1].items[0],
    DIGEST_TODAY.topics[3].items[0],
    DIGEST_TODAY.topics[0].items[1],
    DIGEST_TODAY.topics[2].items[1],
  ];

  const todayTotal = TODAY_EXPENSES.reduce((s, e) => s + e.amount, 0);

  return (
    <div>
      <TopBar />

      <Container maxWidth="2xl" className="py-8 pb-16">
        {/* Greeting */}
        <div className="mb-7 flex items-end justify-between">
          <div>
            <div className="font-mono-tabular mb-2 text-[11px] uppercase tracking-widest text-ink-soft">
              {TODAY.weekday} · {TODAY.date} · Vol. 187
            </div>
            <Text as="h1" font="marker" size="4xl" className="leading-none">
              Good morning, <em className="italic">Long.</em>
              <span
                onMouseEnter={() => setDolphinSeen(true)}
                style={{
                  marginLeft: 14,
                  fontSize: 34,
                  opacity: dolphinSeen ? 1 : 0.18,
                  transition: "opacity 0.5s",
                  cursor: "default",
                }}
              >
                🐬
              </span>
            </Text>
            <div className="mt-3.5 max-w-[620px] text-[15px] text-ink-soft">
              {DIGEST_TODAY.unread} unread items in today&apos;s digest. You&apos;re at{" "}
              <strong className="text-ink">
                {Math.round((MONTH_SPENT / MONTH_BUDGET) * 100)}%
              </strong>{" "}
              of this month&apos;s budget on day {MONTH_DAYS_IN}.{" "}
              <span className="text-accent">Slightly over pace.</span>
            </div>
          </div>
          <div className="text-right">
            <div className="movoz-margin-note mb-1">quiet morning ☕</div>
            <Divider className="ml-auto w-[120px] border-t-2 border-b-2 border-ink" />
          </div>
        </div>

        {/* Main grid */}
        <div className="grid grid-cols-[1.45fr_1fr] gap-5">
          {/* LEFT — Newsletter */}
          <Card variant="outlined" padding="lg">
            <div className="mb-5 flex items-start justify-between">
              <div>
                <div className="font-mono-tabular mb-1.5 text-[10px] uppercase tracking-widest text-ink-soft">
                  The Morning Digest
                </div>
                <Text as="div" font="marker" size="xl">
                  {DIGEST_TODAY.label}
                </Text>
                <div className="mt-1.5 text-[13px] text-ink-soft">{DIGEST_TODAY.summary}</div>
              </div>
              <Button variant="ghost" size="sm">
                All {DIGEST_TODAY.itemCount} items <Icon.arrowRight style={{ marginLeft: 4, verticalAlign: "middle" }} />
              </Button>
            </div>

            {/* Topic strip */}
            <div className="mb-5 flex flex-wrap gap-2">
              {DIGEST_TODAY.topics.map((t) => (
                <div
                  key={t.slug}
                  className="inline-flex items-center gap-1.5 rounded-full border border-line bg-paper px-2.5 py-1 text-[11px] text-ink"
                >
                  {t.name}{" "}
                  <span className="font-mono-tabular text-[10px] text-ink-soft">{t.count}</span>
                </div>
              ))}
            </div>

            {/* Featured items */}
            <div className="flex flex-col">
              {featuredItems.map((item, i) => {
                const key = item.title;
                const myReact = reactions[key] ?? item.reaction;
                const saved = savedItems.has(key);
                return (
                  <div
                    key={key}
                    className="py-3.5"
                    style={{ borderTop: i === 0 ? "none" : "1px solid var(--line)" }}
                  >
                    <div className="mb-1 flex items-baseline gap-3">
                      <span className="font-mono-tabular min-w-[60px] text-[9px] uppercase tracking-widest text-ink-soft">
                        {item.source.split(".")[0].slice(0, 8)}
                      </span>
                      <a className="font-marker flex-1 cursor-pointer text-lg leading-snug text-ink no-underline">
                        {item.title}
                      </a>
                      <span className="font-mono-tabular text-[10px] text-ink-soft">
                        {item.minutes} min
                      </span>
                    </div>
                    <div className="mb-2 pl-[72px] text-[13px] leading-relaxed text-ink-soft">
                      {item.summary}
                    </div>
                    <div className="flex items-center gap-1 pl-[72px]">
                      <IconButton
                        variant="ghost"
                        size="sm"
                        label="Loved"
                        icon={<Icon.heart />}
                        onClick={() => setReact(key, "loved")}
                        style={{ color: myReact === "loved" ? "var(--bucket-lifestyle)" : "var(--ink-soft)" }}
                      />
                      <IconButton
                        variant="ghost"
                        size="sm"
                        label="Boring"
                        icon={<Icon.yawn />}
                        onClick={() => setReact(key, "boring")}
                        style={{ color: myReact === "boring" ? "var(--ink)" : "var(--ink-soft)" }}
                      />
                      <IconButton
                        variant="ghost"
                        size="sm"
                        label="Kill source"
                        icon={<Icon.kill />}
                        onClick={() => setReact(key, "kill")}
                        style={{ color: myReact === "kill" ? "var(--terracotta)" : "var(--ink-soft)" }}
                      />
                      <span className="mx-1.5 h-3.5 w-px bg-line" />
                      <IconButton
                        variant="ghost"
                        size="sm"
                        label="Save"
                        icon={<Icon.bookmark />}
                        onClick={() => toggleSave(key)}
                        style={{ color: saved ? "var(--ink)" : "var(--ink-soft)" }}
                      />
                      <IconButton variant="ghost" size="sm" label="Open" icon={<Icon.external />} />
                      {item.tag && (
                        <span className="font-mono-tabular ml-auto rounded border border-line px-1.5 py-0.5 text-[9px] tracking-widest text-ink-soft">
                          {item.tag.toUpperCase()}
                        </span>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </Card>

          {/* RIGHT — Money */}
          <div className="flex flex-col gap-5">
            {/* Budget gauge */}
            <Card variant="outlined" padding="lg">
              <div className="mb-2 flex items-baseline justify-between">
                <div>
                  <div className="font-mono-tabular mb-1 text-[10px] uppercase tracking-widest text-ink-soft">
                    May 2026 · Budget
                  </div>
                  <Text as="div" font="marker" size="lg">
                    Month-to-date
                  </Text>
                </div>
                <Button variant="ghost" size="sm">
                  Details
                </Button>
              </div>

              <div className="mb-1.5 flex justify-center">
                <BudgetGauge
                  spent={MONTH_SPENT}
                  budget={MONTH_BUDGET}
                  daysIn={MONTH_DAYS_IN}
                  daysTotal={MONTH_DAYS_TOTAL}
                />
              </div>

              {/* Bucket strip */}
              <div className="mb-2.5 flex h-2 overflow-hidden rounded">
                {MONTH_BY_BUCKET.map((b) => (
                  <div
                    key={b.bucket}
                    style={{ flex: b.amount, backgroundColor: `var(--bucket-${b.bucket})` }}
                  />
                ))}
              </div>
              <div className="flex justify-between text-[11px]">
                {MONTH_BY_BUCKET.map((b) => (
                  <div key={b.bucket} className="flex-1">
                    <div className="flex items-center gap-1.5">
                      <BucketDot bucket={b.bucket} />
                      <span className="font-mono-tabular text-[10px] uppercase tracking-widest text-ink-soft">
                        {b.label}
                      </span>
                    </div>
                    <div className="font-mono-tabular mt-1 text-sm font-medium text-ink">
                      {fmtVND(b.amount)}
                      <span className="text-[11px] text-ink-soft"> ₫</span>
                    </div>
                  </div>
                ))}
              </div>
            </Card>

            {/* Today's expenses */}
            <Card variant="outlined" padding="lg">
              <div className="mb-3.5 flex items-baseline justify-between">
                <div>
                  <div className="font-mono-tabular mb-1 text-[10px] uppercase tracking-widest text-ink-soft">
                    Today · {TODAY_EXPENSES.length} entries
                  </div>
                  <div className="flex items-baseline gap-2">
                    <Text as="span" font="marker" size="2xl">
                      {fmtVND(todayTotal)}
                      <span className="text-lg text-ink-soft">₫</span>
                    </Text>
                    <span
                      className="font-mono-tabular text-[10px] tracking-wide"
                      style={{
                        color:
                          YESTERDAY_TOTAL > todayTotal ? "var(--bucket-essentials)" : "var(--terracotta)",
                      }}
                    >
                      <Icon.arrowDown style={{ verticalAlign: "middle" }} />{" "}
                      {Math.round((1 - todayTotal / YESTERDAY_TOTAL) * 100)}% vs yest.
                    </span>
                  </div>
                </div>
                <Button variant="primary" size="sm">
                  <Icon.plus style={{ verticalAlign: "middle" }} /> Add
                </Button>
              </div>

              <div>
                {TODAY_EXPENSES.map((e, i) => (
                  <div
                    key={e.id}
                    className="flex items-center justify-between py-2.5"
                    style={{
                      borderBottom: i === TODAY_EXPENSES.length - 1 ? "none" : "1px solid var(--line)",
                    }}
                  >
                    <div className="flex items-center gap-3">
                      <span className="font-mono-tabular min-w-[40px] text-[10px] text-ink-soft">
                        {e.time}
                      </span>
                      <BucketDot bucket={e.bucket} />
                      <span className="text-[13px] text-ink">{e.text}</span>
                    </div>
                    <span className="font-mono-tabular text-[13px] font-medium text-ink">
                      {fmtVND(e.amount)}
                      <span className="text-ink-soft"> ₫</span>
                    </span>
                  </div>
                ))}
              </div>

              <Divider className="my-3.5" />
              <div className="flex items-center justify-between">
                <div className="font-mono-tabular text-[10px] uppercase tracking-widest text-ink-soft">
                  Last 9 days
                </div>
                <Sparkline
                  data={DAILY_SPEND}
                  width={160}
                  height={26}
                  stroke="var(--ink)"
                  marker={DAILY_SPEND.length - 1}
                />
              </div>
            </Card>
          </div>
        </div>

        {/* Footer strip — agent status */}
        <div className="mt-7 flex items-center justify-between border-t border-line px-1 py-3.5">
          <div className="font-mono-tabular text-[10px] uppercase tracking-widest text-ink-soft">
            <span className="mr-2 inline-block h-1.5 w-1.5 rounded-full bg-bucket-essentials align-middle" />
            Drunken Dolphin · last digest 7:01 · last expense 19:20 · all systems nominal
          </div>
          <div className="font-mono-tabular text-[10px] tracking-widest text-ink-soft">
            movoz.private · v1.0
          </div>
        </div>
      </Container>
    </div>
  );
}
