"use client";

import { useEffect, useState, type FormEvent } from "react";
import { api } from "@/lib/api";
import type { MainAttribute, SubAttribute } from "@/lib/types";

export default function AttributesPage() {
  const [mainAttributes, setMainAttributes] = useState<MainAttribute[]>([]);
  const [subAttributesByMain, setSubAttributesByMain] = useState<Record<number, SubAttribute[]>>({});
  const [newMainKey, setNewMainKey] = useState("");
  const [newMainName, setNewMainName] = useState("");
  const [newSubName, setNewSubName] = useState<Record<number, string>>({});

  async function load() {
    const mains = await api.get<MainAttribute[]>("/api/main-attributes");
    setMainAttributes(mains);
    const entries = await Promise.all(
      mains.map(async (m) => [m.id, await api.get<SubAttribute[]>(`/api/sub-attributes?main_attribute_id=${m.id}&active=all`)] as const)
    );
    setSubAttributesByMain(Object.fromEntries(entries));
  }

  useEffect(() => {
    load();
  }, []);

  async function handleCreateMain(e: FormEvent) {
    e.preventDefault();
    await api.post("/api/main-attributes", { key: newMainKey, name: newMainName });
    setNewMainKey("");
    setNewMainName("");
    await load();
  }

  async function handleCreateSub(mainAttributeId: number, e: FormEvent) {
    e.preventDefault();
    const name = newSubName[mainAttributeId];
    if (!name) return;
    await api.post("/api/sub-attributes", { main_attribute_id: mainAttributeId, name });
    setNewSubName((prev) => ({ ...prev, [mainAttributeId]: "" }));
    await load();
  }

  async function toggleSubActive(sub: SubAttribute) {
    if (sub.is_active) {
      await api.delete(`/api/sub-attributes/${sub.id}`);
      await load();
    }
  }

  return (
    <main className="mx-auto max-w-3xl p-8">
      <h1 className="mb-6 text-2xl font-semibold text-zen-text">Attributes</h1>

      <form onSubmit={handleCreateMain} className="mb-8 flex gap-3 rounded-lg border border-zen-border p-4">
        <input className="flex-1 rounded border border-zen-border bg-transparent p-2" placeholder="key (e.g. delivery_speed)" value={newMainKey} onChange={(e) => setNewMainKey(e.target.value)} required />
        <input className="flex-1 rounded border border-zen-border bg-transparent p-2" placeholder="Name (e.g. Delivery Speed)" value={newMainName} onChange={(e) => setNewMainName(e.target.value)} required />
        <button type="submit" className="rounded bg-accent-600 px-3 py-2 text-white">Add main attribute</button>
      </form>

      <div className="space-y-6">
        {mainAttributes.map((main) => (
          <section key={main.id} className="rounded-lg border border-zen-border p-4">
            <h2 className="mb-3 text-lg font-medium text-zen-text">{main.name}</h2>
            <ul className="mb-3 divide-y divide-zen-border">
              {(subAttributesByMain[main.id] ?? []).map((sub) => (
                <li key={sub.id} className="flex items-center justify-between py-2">
                  <span className={sub.is_active ? "text-zen-text" : "text-zen-muted line-through"}>{sub.name}</span>
                  {sub.is_active && (
                    <button onClick={() => toggleSubActive(sub)} className="text-sm text-zen-muted hover:text-zen-text">
                      Deactivate
                    </button>
                  )}
                </li>
              ))}
            </ul>
            <form onSubmit={(e) => handleCreateSub(main.id, e)} className="flex gap-2">
              <input
                className="flex-1 rounded border border-zen-border bg-transparent p-2 text-sm"
                placeholder="New sub-attribute name"
                value={newSubName[main.id] ?? ""}
                onChange={(e) => setNewSubName((prev) => ({ ...prev, [main.id]: e.target.value }))}
              />
              <button type="submit" className="rounded bg-accent-600 px-3 py-1 text-sm text-white">Add</button>
            </form>
          </section>
        ))}
      </div>
    </main>
  );
}
