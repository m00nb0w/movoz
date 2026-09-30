"use client";

import { Badge, Button, Card, Text } from "@movoz/ui-web";
import type { WorkItem } from "@/lib/api";

export function BacklogList({
  items,
  onPickForToday,
  onOpen,
}: {
  items: WorkItem[];
  onPickForToday: (id: number) => void;
  onOpen: (item: WorkItem) => void;
}) {
  return (
    <div className="mt-8">
      <div className="mb-2 flex items-center gap-2">
        <Text as="h2" font="marker" weight="semibold" size="sm" color="muted">
          Backlog
        </Text>
        <Badge variant="outline" color="default" size="sm">
          {items.length}
        </Badge>
      </div>

      {items.length === 0 ? (
        <Text size="sm" color="muted">
          Nothing in the backlog.
        </Text>
      ) : (
        <div className="flex flex-col gap-2">
          {items.map((item) => (
            <Card
              key={item.id}
              variant="outlined"
              padding="sm"
              draggable
              onDragStart={(e) => e.dataTransfer.setData("text/plain", String(item.id))}
              onClick={() => onOpen(item)}
              className="flex cursor-pointer items-center justify-between gap-2"
              style={
                item.source === "jira"
                  ? { borderLeftWidth: 3, borderLeftColor: "var(--terracotta)" }
                  : undefined
              }
            >
              <div className="flex min-w-0 items-center gap-2">
                <Badge variant="outline" color="default" size="sm">
                  {item.reference_key}
                </Badge>
                <Text as="span" size="sm" truncate>
                  {item.title}
                </Text>
              </div>
              <Button
                variant="secondary"
                size="sm"
                onClick={(e) => {
                  e.stopPropagation();
                  onPickForToday(item.id);
                }}
              >
                Pick for today
              </Button>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
