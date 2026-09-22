"use client";

import { Badge, Button, Card, Dropdown, Text } from "@movoz/ui-web";
import type { WorkItem, WorkItemStatus } from "@/lib/api";

const STATUS_OPTIONS: { label: string; value: WorkItemStatus }[] = [
  { label: "To Do", value: "todo" },
  { label: "In Progress", value: "in_progress" },
  { label: "Blocked", value: "blocked" },
  { label: "Done", value: "done" },
];

export function WorkItemCard({
  item,
  onStatusChange,
  onDelete,
}: {
  item: WorkItem;
  onStatusChange: (id: number, status: WorkItemStatus) => void;
  onDelete: (id: number) => void;
}) {
  const currentStatus = STATUS_OPTIONS.find((s) => s.value === item.status);

  const isJira = item.source === "jira";

  return (
    <Card
      variant="outlined"
      padding="sm"
      style={isJira ? { borderLeftWidth: 3, borderLeftColor: "var(--terracotta)" } : undefined}
    >
      <div className="flex items-start justify-between gap-2">
        <Text as="h3" font="marker" weight="semibold" size="base">
          {item.title}
        </Text>
        <Badge variant="outline" color="default" size="sm">
          {item.type}
        </Badge>
      </div>

      {item.description && (
        <Text size="sm" color="muted" className="mt-1">
          {item.description}
        </Text>
      )}

      <div className="mt-2 flex items-center justify-between gap-2">
        <Badge variant="subtle" color={isJira ? "accent" : "default"} size="sm">
          {item.source}
        </Badge>
        {item.jira_url && (
          <Text as="a" href={item.jira_url} target="_blank" rel="noreferrer" size="xs" color="accent">
            {item.jira_key}
          </Text>
        )}
      </div>

      <div className="mt-3 flex items-center gap-2">
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
