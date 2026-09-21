"use client";

import type { WorkItem, WorkItemStatus } from "@/lib/api";

const STATUS_OPTIONS: WorkItemStatus[] = ["todo", "in_progress", "blocked", "done"];

export function WorkItemCard({
  item,
  onStatusChange,
  onDelete,
}: {
  item: WorkItem;
  onStatusChange: (id: number, status: WorkItemStatus) => void;
  onDelete: (id: number) => void;
}) {
  return (
    <div className="rounded-[var(--radius-md)] border border-line bg-paper-raised p-3 shadow-sm">
      <div className="flex items-start justify-between gap-2">
        <span className="text-sm font-medium text-ink">{item.title}</span>
        <span className="rounded-full bg-accent-soft px-2 py-0.5 text-xs text-accent-dark">
          {item.type}
        </span>
      </div>
      {item.description && (
        <p className="mt-1 text-xs text-ink-soft">{item.description}</p>
      )}
      <div className="mt-2 flex items-center justify-between gap-2 text-xs text-ink-soft">
        <span>{item.source}</span>
        {item.jira_url && (
          <a href={item.jira_url} target="_blank" rel="noreferrer" className="underline">
            {item.jira_key}
          </a>
        )}
      </div>
      <div className="mt-2 flex items-center gap-2">
        <select
          className="rounded border border-line bg-paper px-1 py-0.5 text-xs text-ink"
          value={item.status}
          onChange={(e) => onStatusChange(item.id, e.target.value as WorkItemStatus)}
        >
          {STATUS_OPTIONS.map((status) => (
            <option key={status} value={status}>
              {status}
            </option>
          ))}
        </select>
        <button
          className="text-xs text-ink-soft underline"
          onClick={() => onDelete(item.id)}
        >
          delete
        </button>
      </div>
    </div>
  );
}
