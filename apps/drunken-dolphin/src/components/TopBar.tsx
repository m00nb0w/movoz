"use client";

import { usePathname, useRouter } from "next/navigation";
import { Avatar, Container, Tabs, Text } from "@movoz/ui-web";
import { Icon } from "./icons";

const ROUTES = [
  { value: "/", label: "Today" },
  { value: "/newsletter", label: "Newsletter" },
  { value: "/archive", label: "Archive" },
  { value: "/expenses", label: "Expenses" },
];

export function TopBar({ showSearch = true }: { showSearch?: boolean }) {
  const pathname = usePathname();
  const router = useRouter();

  return (
    <div className="border-b border-line bg-paper">
      <Container maxWidth="2xl" className="flex items-center justify-between py-3">
        <div className="flex items-center gap-7">
          <Text as="span" font="marker" weight="semibold" size="lg" color="accent">
            Drunken Dolphin
          </Text>
          <Tabs items={ROUTES} value={pathname} onChange={(value) => router.push(value)} />
        </div>
        <div className="flex items-center gap-4">
          {showSearch && (
            <div className="flex min-w-[220px] items-center gap-2 rounded-[var(--radius-sm)] border border-line bg-paper-raised px-3 py-1.5">
              <Icon.search className="text-ink-soft" />
              <span className="text-xs text-ink-soft">Search digests, expenses…</span>
              <span className="font-mono-tabular ml-auto rounded border border-line px-1 text-[10px] text-ink-soft">
                ⌘K
              </span>
            </div>
          )}
          <span className="font-mono-tabular text-xs tracking-wide text-ink-soft">
            ASIA/SAIGON · 9:42
          </span>
          <Avatar fallback="Long" size="sm" />
        </div>
      </Container>
    </div>
  );
}
