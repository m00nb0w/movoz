"use client";

import { Tabs } from "@movoz/ui-web";
import { useScope } from "@/lib/scope";
import type { Scope } from "@/lib/api";

const SCOPES = [
  { value: "mine", label: "Mine" },
  { value: "agent", label: "Agent" },
];

export function ScopeToggle() {
  const { scope, setScope } = useScope();
  return <Tabs items={SCOPES} value={scope} onChange={(value) => setScope(value as Scope)} />;
}
