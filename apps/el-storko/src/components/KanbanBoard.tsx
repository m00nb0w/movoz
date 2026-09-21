"use client";

import { useEffect, useState } from "react";
import {
  listWorkItems,
  updateWorkItem,
  deleteWorkItem,
  createWorkItem,
  type WorkItem,
  type WorkItemStatus,
} from "@/lib/api";
import { WorkItemCard } from "./WorkItemCard";

const COLUMNS: { status: WorkItemStatus; label: string }[] = [
  { status: "todo", label: "To Do" },
  { status: "in_progress", label: "In Progress" },
  { status: "blocked", label: "Blocked" },
  { status: "done", label: "Done" },
];

export function KanbanBoard() {
  const [items, setItems] = useState<WorkItem[]>([]);
  const [newTitle, setNewTitle] = useState("");
  const [error, setError] = useState<string | null>(null);

  async function refresh() {
    try {
      const data = await listWorkItems();
      setItems(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "failed to load work items");
    }
  }

  useEffect(() => {
    refresh();
  }, []);

  async function handleAdd() {
    if (!newTitle.trim()) return;
    await createWorkItem({ type: "task", title: newTitle.trim() });
    setNewTitle("");
    await refresh();
  }

  async function handleStatusChange(id: number, status: WorkItemStatus) {
    await updateWorkItem(id, { status });
    await refresh();
  }

  async function handleDelete(id: number) {
    await deleteWorkItem(id);
    await refresh();
  }

  return (
    <div>
      <div className="mb-6 flex gap-2">
        <input
          className="flex-1 rounded border border-line bg-paper px-3 py-2 text-sm text-ink"
          placeholder="Quick add a task..."
          value={newTitle}
          onChange={(e) => setNewTitle(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleAdd()}
        />
        <button
          className="rounded bg-accent px-4 py-2 text-sm text-white"
          onClick={handleAdd}
        >
          Add
        </button>
      </div>

      {error && <p className="mb-4 text-sm text-accent-dark">{error}</p>}

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {COLUMNS.map((column) => (
          <div key={column.status}>
            <h2 className="mb-2 text-sm font-semibold text-ink-soft">{column.label}</h2>
            <div className="flex flex-col gap-2">
              {items
                .filter((item) => item.status === column.status)
                .map((item) => (
                  <WorkItemCard
                    key={item.id}
                    item={item}
                    onStatusChange={handleStatusChange}
                    onDelete={handleDelete}
                  />
                ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
