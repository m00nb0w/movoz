"use client";

import { useEffect, useState } from "react";
import { Button, Input, Text } from "@movoz/ui-web";
import { listWorkItems, updateWorkItem, type WorkItem } from "@/lib/api";

export function EpicPicker({
  taskId,
  currentParentId,
  onAssigned,
}: {
  taskId: number;
  currentParentId: number | null;
  onAssigned: (item: WorkItem) => void;
}) {
  const [epics, setEpics] = useState<WorkItem[]>([]);
  const [query, setQuery] = useState("");

  useEffect(() => {
    listWorkItems({ type: "epic" }).then(setEpics).catch(() => {});
  }, []);

  const currentEpic = epics.find((epic) => epic.id === currentParentId);
  const matches = query.trim()
    ? epics.filter((epic) => epic.title.toLowerCase().includes(query.trim().toLowerCase()))
    : epics;

  async function assign(epicId: number | null) {
    const updated = await updateWorkItem(taskId, { parent_id: epicId });
    onAssigned(updated);
    setQuery("");
  }

  return (
    <div>
      <Text as="span" size="sm" weight="semibold" className="mb-1 block">
        Epic
      </Text>
      {currentEpic && (
        <div className="mb-2 flex items-center gap-2">
          <Text as="span" size="sm">
            {currentEpic.title}
          </Text>
          <Button variant="ghost" size="sm" onClick={() => assign(null)}>
            Clear
          </Button>
        </div>
      )}
      <Input
        placeholder="Search epics…"
        size="sm"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
      />
      {query.trim() && (
        <div className="mt-1 flex max-h-32 flex-col overflow-auto rounded-[var(--radius-sm)] border border-line">
          {matches.length === 0 ? (
            <Text size="xs" color="muted" className="p-2">
              No matching epics
            </Text>
          ) : (
            matches.map((epic) => (
              <button
                key={epic.id}
                type="button"
                onClick={() => assign(epic.id)}
                className="px-3 py-1.5 text-left text-sm text-ink hover:bg-paper-sunken"
              >
                {epic.title}
              </button>
            ))
          )}
        </div>
      )}
    </div>
  );
}
