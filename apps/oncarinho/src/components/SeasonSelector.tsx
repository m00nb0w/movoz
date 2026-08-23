"use client";

import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";

interface SeasonSelectorProps {
  years: number[];
  selected: number | "all";
  basePath: string;
  includeAllTime?: boolean;
  extraParams?: Record<string, string>;
}

export function SeasonSelector({
  years,
  selected,
  basePath,
  includeAllTime = false,
  extraParams = {},
}: SeasonSelectorProps) {
  const router = useRouter();
  const t = useTranslations("common");

  function select(value: number | "all") {
    const params = new URLSearchParams({ ...extraParams, year: String(value) });
    router.push(`${basePath}?${params.toString()}`);
  }

  const options: (number | "all")[] = includeAllTime ? ["all", ...years] : years;

  return (
    <div role="radiogroup" className="flex flex-wrap gap-1 border border-line p-1">
      {options.map((option) => (
        <button
          key={option}
          role="radio"
          aria-checked={option === selected}
          onClick={() => select(option)}
          className={
            option === selected
              ? "bg-ink px-3 py-1 text-sm text-paper"
              : "px-3 py-1 text-sm text-ink-soft hover:text-ink"
          }
        >
          {option === "all" ? t("allTime") : option}
        </button>
      ))}
    </div>
  );
}
