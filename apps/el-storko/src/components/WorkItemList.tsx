"use client";

import { useEffect, useState } from "react";
import { Badge, Pill, Text } from "@movoz/ui-web";
import { listWorkItems, type WorkItem, type WorkItemSource } from "@/lib/api";

const SOURCE_FILTERS: { label: string; value: WorkItemSource | "" }[] = [
  { label: "All sources", value: "" },
  { label: "Personal", value: "personal" },
  { label: "Jira", value: "jira" },
  { label: "Agent", value: "agent" },
];

export function WorkItemList() {
  const [items, setItems] = useState<WorkItem[]>([]);
  const [epics, setEpics] = useState<WorkItem[]>([]);
  const [source, setSource] = useState<WorkItemSource | "">("");
  const [epicId, setEpicId] = useState<string>("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    listWorkItems({ type: "epic" }).then(setEpics).catch(() => {});
  }, []);

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
      <div className="mb-4 flex flex-wrap items-center gap-2">
        {SOURCE_FILTERS.map((filter) => (
          <Pill key={filter.value} active={source === filter.value} onClick={() => setSource(filter.value)}>
            {filter.label}
          </Pill>
        ))}
        <Pill active={epicId === ""} onClick={() => setEpicId("")}>
          All epics
        </Pill>
        {epics.map((epic) => (
          <Pill key={epic.id} active={epicId === String(epic.id)} onClick={() => setEpicId(String(epic.id))}>
            {epic.title}
          </Pill>
        ))}
      </div>

      {error && (
        <Text size="sm" color="accent" className="mb-4">
          {error}
        </Text>
      )}

      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-line text-left">
            <th className="py-2">
              <Text as="span" size="sm" color="muted" weight="semibold">
                Title
              </Text>
            </th>
            <th className="py-2">
              <Text as="span" size="sm" color="muted" weight="semibold">
                Type
              </Text>
            </th>
            <th className="py-2">
              <Text as="span" size="sm" color="muted" weight="semibold">
                Status
              </Text>
            </th>
            <th className="py-2">
              <Text as="span" size="sm" color="muted" weight="semibold">
                Source
              </Text>
            </th>
          </tr>
        </thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.id} className="border-b border-line">
              <td className="py-2">
                <Text as="span" size="sm">
                  {item.title}
                </Text>
              </td>
              <td className="py-2">
                <Badge variant="outline" color="default" size="sm">
                  {item.type}
                </Badge>
              </td>
              <td className="py-2">
                <Text as="span" size="sm" color="muted">
                  {item.status}
                </Text>
              </td>
              <td className="py-2">
                <Badge variant="subtle" color={item.source === "jira" ? "accent" : "default"} size="sm">
                  {item.source}
                </Badge>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
