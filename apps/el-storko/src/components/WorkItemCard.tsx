"use client";

import { Badge, Button, Card, Dropdown, Text } from "@movoz/ui-web";
import type { WorkItem, WorkItemStatus } from "@/lib/api";
import { getDueDateColor } from "@/lib/dueDate";

const STATUS_OPTIONS: { label: string; value: WorkItemStatus }[] = [
  { label: "Backlog", value: "backlog" },
  { label: "Picked for today", value: "picked_for_today" },
  { label: "In progress", value: "in_progress" },
  { label: "Blocked", value: "blocked" },
  { label: "Done", value: "done" },
];

const DUE_DATE_COLOR_CLASS: Record<"red" | "yellow", string> = {
  red: "text-red-600 dark:text-red-400",
  yellow: "text-amber-600 dark:text-amber-400",
};

export function WorkItemCard({
  item,
  onStatusChange,
  onDelete,
  onOpen,
  draggable = false,
}: {
  item: WorkItem;
  onStatusChange: (id: number, status: WorkItemStatus) => void;
  onDelete: (id: number) => void;
  onOpen?: (item: WorkItem) => void;
  draggable?: boolean;
}) {
  const currentStatus = STATUS_OPTIONS.find((s) => s.value === item.status);

  const isJira = item.source === "jira";
  const dueDateColor = getDueDateColor(item.due_date);

  return (
    <Card
      variant="outlined"
      padding="sm"
      draggable={draggable}
      onDragStart={
        draggable
          ? (e) => e.dataTransfer.setData("text/plain", String(item.id))
          : undefined
      }
      onClick={() => onOpen?.(item)}
      className={onOpen ? "cursor-pointer" : undefined}
      style={isJira ? { borderLeftWidth: 3, borderLeftColor: "var(--terracotta)" } : undefined}
    >
      <div className="flex items-start justify-between gap-2">
        <Text as="h3" font="marker" weight="semibold" size="base">
          {item.title}
        </Text>
        <Badge variant="outline" color="default" size="sm">
          {item.reference_key}
        </Badge>
      </div>

      {item.description && (
        <Text size="sm" color="muted" className="mt-1">
          {item.description}
        </Text>
      )}

      <div className="mt-2 flex flex-wrap items-center gap-2">
        {isJira ? (
          <a
            href={item.jira_url ?? undefined}
            target="_blank"
            rel="noreferrer"
            onClick={(e) => e.stopPropagation()}
          >
            <Badge variant="subtle" color="accent" size="sm">
              {item.jira_key}
            </Badge>
          </a>
        ) : (
          <Badge variant="subtle" color="default" size="sm">
            {item.source}
          </Badge>
        )}
        {item.due_date && (
          <Text
            size="xs"
            className={dueDateColor ? DUE_DATE_COLOR_CLASS[dueDateColor] : "text-ink-soft"}
          >
            Due {item.due_date}
          </Text>
        )}
      </div>

      <div className="mt-3 flex items-center gap-2" onClick={(e) => e.stopPropagation()}>
        <Dropdown
          trigger={
            <Badge variant="outline" color="default" size="sm">
              {currentStatus?.label ?? item.status}
            </Badge>
          }
          items={STATUS_OPTIONS}
          onSelect={(value) => onStatusChange(item.id, value as WorkItemStatus)}
        />
        <Button variant="ghost" size="sm" onClick={() => onDelete(item.id)}>
          Delete
        </Button>
      </div>
    </Card>
  );
}
