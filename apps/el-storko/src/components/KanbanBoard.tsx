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
import { BacklogList } from "./BacklogList";
import { ItemDrawer } from "./ItemDrawer";

const COLUMNS: { status: WorkItemStatus; label: string }[] = [
  { status: "picked_for_today", label: "Picked for today" },
  { status: "in_progress", label: "In progress" },
  { status: "blocked", label: "Blocked" },
  { status: "done", label: "Done" },
];

export function KanbanBoard() {
  const [items, setItems] = useState<WorkItem[]>([]);
  const [newTitle, setNewTitle] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [openItem, setOpenItem] = useState<WorkItem | null>(null);
  const [isDragOver, setIsDragOver] = useState(false);

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

  async function handlePickForToday(id: number) {
    await handleStatusChange(id, "picked_for_today");
  }

  function handleDrop(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault();
    setIsDragOver(false);
    const idRaw = e.dataTransfer.getData("text/plain");
    const id = Number(idRaw);
    if (!id) return;
    handlePickForToday(id);
  }

  const backlogItems = items.filter((item) => item.status === "backlog");

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

      <div
        onDragOver={(e) => {
          e.preventDefault();
          setIsDragOver(true);
        }}
        onDragLeave={() => setIsDragOver(false)}
        onDrop={handleDrop}
        className={
          isDragOver
            ? "rounded-[var(--radius-md)] outline-2 outline-dashed outline-accent"
            : undefined
        }
      >
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
                      onOpen={setOpenItem}
                    />
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      <BacklogList items={backlogItems} onPickForToday={handlePickForToday} onOpen={setOpenItem} />

      <ItemDrawer
        item={openItem}
        onClose={() => setOpenItem(null)}
        onUpdated={(updated) => {
          setItems((prev) => prev.map((it) => (it.id === updated.id ? updated : it)));
          setOpenItem(updated);
        }}
      />
    </div>
  );
}
