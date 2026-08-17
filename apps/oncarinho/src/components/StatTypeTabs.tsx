"use client";

import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";

const STATS = ["goals", "assists", "cards"] as const;
type Stat = (typeof STATS)[number];

interface StatTypeTabsProps {
  selected: Stat;
  basePath: string;
  extraParams?: Record<string, string>;
}

export function StatTypeTabs({ selected, basePath, extraParams = {} }: StatTypeTabsProps) {
  const router = useRouter();
  const t = useTranslations("leaderboard");

  function select(stat: Stat) {
    const params = new URLSearchParams({ ...extraParams, stat });
    router.push(`${basePath}?${params.toString()}`);
  }

  return (
    <div role="radiogroup" className="flex flex-wrap gap-1 border border-zen-border p-1">
      {STATS.map((stat) => (
        <button
          key={stat}
          role="radio"
          aria-checked={stat === selected}
          onClick={() => select(stat)}
          className={
            stat === selected
              ? "bg-zen-text px-3 py-1 text-sm text-zen-bg"
              : "px-3 py-1 text-sm text-zen-muted hover:text-zen-text"
          }
        >
          {t(stat)}
        </button>
      ))}
    </div>
  );
}
