"use client";

import { usePathname, useRouter } from "next/navigation";
import { Tabs } from "@movoz/ui-web";

const ROUTES = [
  { value: "/", label: "Board" },
  { value: "/list", label: "List" },
  { value: "/stats", label: "Stats" },
];

export function NavTabs() {
  const pathname = usePathname();
  const router = useRouter();

  return <Tabs items={ROUTES} value={pathname} onChange={(value) => router.push(value)} />;
}
