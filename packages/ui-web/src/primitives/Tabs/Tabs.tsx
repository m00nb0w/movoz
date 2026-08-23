"use client";

import { cn } from "../../utils/cn";

export interface TabItem {
  value: string;
  label: string;
  accent?: boolean;
}

export interface TabsProps {
  items: (string | TabItem)[];
  value: string;
  onChange: (value: string) => void;
  tone?: "light" | "dock";
  className?: string;
}

export function Tabs({ items, value, onChange, tone = "light", className }: TabsProps) {
  const dark = tone === "dock";

  return (
    <div
      role="tablist"
      className={cn(
        "inline-flex items-center gap-1.5 p-1.5 rounded-[var(--radius-lg)] border",
        dark ? "bg-dock border-dock-line" : "bg-paper-raised border-line",
        className
      )}
    >
      {items.map((item) => {
        const itemValue = typeof item === "string" ? item : item.value;
        const label = typeof item === "string" ? item : item.label;
        const accent = typeof item === "object" && item.accent;
        const active = itemValue === value;

        return (
          <button
            key={itemValue}
            type="button"
            role="tab"
            aria-selected={active}
            onClick={() => onChange(itemValue)}
            className={cn(
              "px-4 py-2 rounded-[var(--radius-md)] font-marker font-semibold text-base leading-none transition-colors duration-150",
              active
                ? accent
                  ? "bg-accent text-white"
                  : dark
                    ? "bg-paper text-ink"
                    : "bg-pencil-light text-ink"
                : accent
                  ? "text-accent border border-accent"
                  : dark
                    ? "text-pencil border border-transparent"
                    : "text-ink-soft border border-transparent"
            )}
          >
            {label}
          </button>
        );
      })}
    </div>
  );
}
