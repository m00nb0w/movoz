"use client";

import { useEffect, useState } from "react";
import { Badge, Dropdown, Input, Modal, Text } from "@movoz/ui-web";
import { updateWorkItem, type UpdateWorkItemInput, type WorkItem, type WorkItemStatus } from "@/lib/api";
import { getDueDateColor } from "@/lib/dueDate";
import { EpicPicker } from "./EpicPicker";

const STATUS_OPTIONS: { label: string; value: WorkItemStatus }[] = [
  { label: "Backlog", value: "backlog" },
  { label: "Picked for today", value: "picked_for_today" },
  { label: "In progress", value: "in_progress" },
  { label: "Blocked", value: "blocked" },
  { label: "Done", value: "done" },
];

export function ItemDrawer({
  item,
  onClose,
  onUpdated,
}: {
  item: WorkItem | null;
  onClose: () => void;
  onUpdated: (item: WorkItem) => void;
}) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [estimateHours, setEstimateHours] = useState("");
  const [dueDate, setDueDate] = useState("");

  useEffect(() => {
    setTitle(item?.title ?? "");
    setDescription(item?.description ?? "");
    setEstimateHours(item?.estimate_hours != null ? String(item.estimate_hours) : "");
    setDueDate(item?.due_date ?? "");
  }, [item]);

  if (!item) return null;

  const currentItem = item;

  async function save(fields: UpdateWorkItemInput) {
    const updated = await updateWorkItem(currentItem.id, fields);
    onUpdated(updated);
  }

  const currentStatus = STATUS_OPTIONS.find((s) => s.value === item.status);
  const dueDateColor = getDueDateColor(item.due_date);
  const dueDateColorClass =
    dueDateColor === "red"
      ? "border-red-500 text-red-600"
      : dueDateColor === "yellow"
        ? "border-amber-500 text-amber-600"
        : undefined;

  return (
    <Modal open onClose={onClose} title={item.reference_key} size="lg">
      <div className="flex flex-col gap-3">
        <Input
          label="Title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          onBlur={() => title !== item.title && save({ title })}
        />
        <Input
          label="Description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          onBlur={() => description !== item.description && save({ description })}
        />

        <div>
          <Text as="span" size="sm" weight="semibold" className="mb-1 block">
            Status
          </Text>
          <Dropdown
            trigger={
              <Badge variant="outline" color="default" size="sm">
                {currentStatus?.label ?? item.status}
              </Badge>
            }
            items={STATUS_OPTIONS}
            onSelect={(value) => save({ status: value as WorkItemStatus })}
          />
        </div>

        <div className="grid grid-cols-2 gap-3">
          <Input
            label="Estimate (hours)"
            type="number"
            min="0"
            step="0.5"
            value={estimateHours}
            onChange={(e) => setEstimateHours(e.target.value)}
            onBlur={() => {
              const parsed = estimateHours.trim() === "" ? null : Number(estimateHours);
              if (parsed !== item.estimate_hours) save({ estimate_hours: parsed });
            }}
          />
          <Input
            label="Due date"
            type="date"
            value={dueDate}
            onChange={(e) => setDueDate(e.target.value)}
            onBlur={() => {
              const parsed = dueDate.trim() === "" ? null : dueDate;
              if (parsed !== item.due_date) save({ due_date: parsed });
            }}
            className={dueDateColorClass}
          />
        </div>

        {item.type === "task" && (
          <EpicPicker taskId={item.id} currentParentId={item.parent_id} onAssigned={onUpdated} />
        )}

        <div className="flex items-center gap-2 border-t border-line pt-3">
          <Text as="span" size="xs" color="muted">
            Reference
          </Text>
          <Badge variant="outline" color="default" size="sm">
            {item.reference_key}
          </Badge>
          {item.source === "jira" && item.jira_key && (
            <a href={item.jira_url ?? undefined} target="_blank" rel="noreferrer">
              <Badge variant="subtle" color="accent" size="sm">
                {item.jira_key}
              </Badge>
            </a>
          )}
        </div>
      </div>
    </Modal>
  );
}
