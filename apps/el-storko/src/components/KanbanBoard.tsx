"use client";

import { useEffect, useState } from "react";
import { Badge, Button, Input, Text } from "@movoz/ui-web";
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
        <Input
          className="flex-1"
          placeholder="Quick add a task…"
          value={newTitle}
          onChange={(e) => setNewTitle(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleAdd()}
        />
        <Button variant="primary" onClick={handleAdd}>
          Add
        </Button>
      </div>

      {error && (
        <Text size="sm" color="accent" className="mb-4">
          {error}
        </Text>
      )}

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {COLUMNS.map((column) => {
          const columnItems = items.filter((item) => item.status === column.status);
          return (
            <div key={column.status}>
              <div className="mb-2 flex items-center gap-2">
                <Text as="h2" font="marker" weight="semibold" size="sm" color="muted">
                  {column.label}
                </Text>
                <Badge variant="outline" color="default" size="sm">
                  {columnItems.length}
                </Badge>
              </div>
              <div className="flex flex-col gap-2">
                {columnItems.map((item) => (
                  <WorkItemCard
                    key={item.id}
                    item={item}
                    onStatusChange={handleStatusChange}
                    onDelete={handleDelete}
                  />
                ))}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
