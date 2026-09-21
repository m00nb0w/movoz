"use client";

import { useEffect, useState } from "react";
import { listWorkItems, type WorkItem, type WorkItemSource } from "@/lib/api";

export function WorkItemList() {
  const [items, setItems] = useState<WorkItem[]>([]);
  const [source, setSource] = useState<WorkItemSource | "">("");
  const [epicId, setEpicId] = useState<string>("");
  const [error, setError] = useState<string | null>(null);

  const epics = items.filter((item) => item.type === "epic");

  useEffect(() => {
    async function load() {
      try {
        const data = await listWorkItems({
          source: source || undefined,
          parent_id: epicId ? Number(epicId) : undefined,
        });
        setItems(data);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : "failed to load work items");
      }
    }
    load();
  }, [source, epicId]);

  return (
    <div>
      <div className="mb-4 flex gap-3">
        <select
          className="rounded border border-line bg-paper px-2 py-1 text-sm text-ink"
          value={source}
          onChange={(e) => setSource(e.target.value as WorkItemSource | "")}
        >
          <option value="">All sources</option>
          <option value="personal">Personal</option>
          <option value="jira">Jira</option>
          <option value="agent">Agent</option>
        </select>
        <select
          className="rounded border border-line bg-paper px-2 py-1 text-sm text-ink"
          value={epicId}
          onChange={(e) => setEpicId(e.target.value)}
        >
          <option value="">All epics</option>
          {epics.map((epic) => (
            <option key={epic.id} value={epic.id}>
              {epic.title}
            </option>
          ))}
        </select>
      </div>

      {error && <p className="mb-4 text-sm text-accent-dark">{error}</p>}

      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-line text-left text-ink-soft">
            <th className="py-2">Title</th>
            <th className="py-2">Type</th>
            <th className="py-2">Status</th>
            <th className="py-2">Source</th>
          </tr>
        </thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.id} className="border-b border-line">
              <td className="py-2 text-ink">{item.title}</td>
              <td className="py-2 text-ink-soft">{item.type}</td>
              <td className="py-2 text-ink-soft">{item.status}</td>
              <td className="py-2 text-ink-soft">{item.source}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
