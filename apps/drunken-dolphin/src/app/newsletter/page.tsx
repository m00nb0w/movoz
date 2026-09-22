"use client";

import { useState } from "react";
import { Container, Divider, IconButton, Pill, Text } from "@movoz/ui-web";
import { TopBar } from "@/components/TopBar";
import { Icon } from "@/components/icons";
import { DIGEST_TODAY, type Reaction } from "@/data/mock";

type Tab = "today" | "saved";

export default function NewsletterPage() {
  const [reactions, setReactions] = useState<Record<string, Reaction>>({});
  const [savedItems, setSavedItems] = useState<Set<string>>(new Set());
  const [activeTopic, setActiveTopic] = useState<string>("all");
  const [tab, setTab] = useState<Tab>("today");

  const setReact = (key: string, val: Exclude<Reaction, null>) =>
    setReactions((r) => ({ ...r, [key]: r[key] === val ? null : val }));

  const toggleSave = (k: string) =>
    setSavedItems((s) => {
      const n = new Set(s);
      if (n.has(k)) n.delete(k);
      else n.add(k);
      return n;
    });

  const visibleTopics =
    activeTopic === "all"
      ? DIGEST_TODAY.topics
      : DIGEST_TODAY.topics.filter((t) => t.slug === activeTopic);

  const filterChips = [
    { slug: "all", name: "All", count: DIGEST_TODAY.itemCount },
    ...DIGEST_TODAY.topics.map((t) => ({ slug: t.slug, name: t.name, count: t.count })),
  ];

  return (
    <div>
      <TopBar />

      <Container maxWidth="md" className="py-9 pb-16">
        {/* Masthead */}
        <div className="mb-7 text-center">
          <Divider className="mb-3.5 border-t-2 border-b-2 border-ink" style={{ height: 4 }} />
          <div className="font-mono-tabular mb-1.5 text-[10px] uppercase tracking-[0.24em] text-ink-soft">
            The Morning Digest · Vol. 187 · Sunday Edition
          </div>
          <Text as="h1" font="marker" size="4xl">
            <em className="italic">Sunday,</em> May 3 · 2026
          </Text>
          <div className="mx-auto mt-2 max-w-[580px] text-sm leading-relaxed text-ink-soft">
            {DIGEST_TODAY.summary} · curated by{" "}
            <span className="movoz-margin-note text-lg">Drunken Dolphin 🐬</span> at 7:01 AM
          </div>
          <Divider className="mt-4 border-t-2 border-b-2 border-ink" style={{ height: 4 }} />
        </div>

        {/* Filter row */}
        <div className="mb-7 flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap gap-1">
            {filterChips.map((t) => (
              <Pill key={t.slug} active={activeTopic === t.slug} onClick={() => setActiveTopic(t.slug)}>
                {t.name}{" "}
                <span className="font-mono-tabular text-[10px] opacity-70">{t.count}</span>
              </Pill>
            ))}
          </div>
          <div className="flex gap-1">
            <Pill active={tab === "today"} onClick={() => setTab("today")}>
              Today
            </Pill>
            <Pill active={tab === "saved"} onClick={() => setTab("saved")}>
              <Icon.bookmark style={{ verticalAlign: "middle", marginRight: 4 }} /> Saved ·{" "}
              {savedItems.size}
            </Pill>
          </div>
        </div>

        {/* Topic sections */}
        <div className="grid grid-cols-1 gap-9">
          {visibleTopics.map((topic) => (
            <section key={topic.slug}>
              <div className="mb-3.5 flex items-baseline gap-3.5">
                <Text as="h2" font="marker" size="2xl" className="italic">
                  {topic.name}
                </Text>
                <Divider className="mb-1.5 flex-1" />
                <span className="font-mono-tabular text-[10px] tracking-widest text-ink-soft">
                  {topic.count} ITEMS
                </span>
              </div>

              <div style={{ columnCount: 2, columnGap: 32, columnRule: "1px solid var(--line)" }}>
                {topic.items.map((item) => {
                  const key = item.title;
                  const myReact = reactions[key] ?? item.reaction;
                  const saved = savedItems.has(key);
                  return (
                    <article
                      key={key}
                      className="mb-5 border-b border-dashed border-line pb-4"
                      style={{ breakInside: "avoid" }}
                    >
                      <div className="mb-1.5 flex items-baseline gap-2">
                        <span className="font-mono-tabular text-[9px] uppercase tracking-widest text-ink-soft">
                          {item.source}
                        </span>
                        <span className="flex-1" />
                        <span className="font-mono-tabular text-[9px] text-ink-soft">
                          {item.minutes} min
                        </span>
                        {item.tag && (
                          <span className="font-mono-tabular rounded-[2px] border border-line px-1 py-px text-[9px] tracking-wide text-ink-soft">
                            {item.tag.toUpperCase()}
                          </span>
                        )}
                      </div>
                      <a className="font-marker mb-1.5 block cursor-pointer text-lg leading-snug text-ink no-underline">
                        {item.title}
                      </a>
                      <p className="mb-2.5 text-[13.5px] leading-relaxed text-ink-soft">
                        {item.summary}
                      </p>
                      <div className="flex items-center gap-0.5">
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
                        <span className="mx-1 h-3 w-px bg-line" />
                        <IconButton
                          variant="ghost"
                          size="sm"
                          label="Save"
                          icon={<Icon.bookmark />}
                          onClick={() => toggleSave(key)}
                          style={{ color: saved ? "var(--ink)" : "var(--ink-soft)" }}
                        />
                        <IconButton variant="ghost" size="sm" label="Open" icon={<Icon.external />} />
                      </div>
                    </article>
                  );
                })}
              </div>
            </section>
          ))}
        </div>
      </Container>
    </div>
  );
}
