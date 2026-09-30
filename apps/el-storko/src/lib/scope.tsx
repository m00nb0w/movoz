"use client";

import { createContext, useContext, useState, type ReactNode } from "react";
import type { Scope } from "@/lib/api";

const ScopeContext = createContext<{ scope: Scope; setScope: (scope: Scope) => void } | null>(
  null,
);

export function ScopeProvider({ children }: { children: ReactNode }) {
  const [scope, setScope] = useState<Scope>("mine");
  return <ScopeContext.Provider value={{ scope, setScope }}>{children}</ScopeContext.Provider>;
}

export function useScope() {
  const ctx = useContext(ScopeContext);
  if (!ctx) throw new Error("useScope must be used within a ScopeProvider");
  return ctx;
}
