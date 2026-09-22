"use client";

import { useMemo, useState } from "react";
import { Badge, Button, Divider, Dropdown, Input, Modal, Text } from "@movoz/ui-web";
import { BucketChip } from "@/components/BucketChip";
import { fmtVND } from "@/lib/format";
import type { Bucket } from "@/data/mock";

type Parsed = {
  text: string;
  amount: number;
  bucket: Bucket;
  category: string;
  tag?: string;
};

function parsePlainLanguage(input: string): Parsed | null {
  const m = input.match(/^([^\d]+?)\s+([\d.]+)([kmM]?)\s*(#\w+)?/);
  if (!m) return null;
  const [, text, num, suffix, tag] = m;
  let amount = parseFloat(num);
  if (suffix === "k") amount *= 1_000;
  if (suffix === "m" || suffix === "M") amount *= 1_000_000;

  const t = text.trim().toLowerCase();
  let bucket: Bucket = "lifestyle";
  let category = "Misc";
  if (
    t.includes("lunch") ||
    t.includes("dinner") ||
    t.includes("pho") ||
    t.includes("com")
  ) {
    bucket = "essentials";
    category = "Food";
  } else if (t.includes("coffee") || t.includes("cà phê")) {
    bucket = "lifestyle";
    category = "Coffee";
  } else if (t.includes("grab") || t.includes("taxi")) {
    bucket = "essentials";
    category = "Transport";
  } else if (t.includes("flight") || t.includes("visa")) {
    bucket = "irregular";
    category = "Trip";
  }
  return { text: text.trim(), amount, bucket, category, tag };
}

const BUCKET_OPTIONS: { label: string; value: Bucket }[] = [
  { label: "Essentials", value: "essentials" },
  { label: "Lifestyle", value: "lifestyle" },
  { label: "Irregular", value: "irregular" },
];

type Props = {
  parseInput: string;
  onParseInputChange: (s: string) => void;
  onCancel: () => void;
  onSave: (parsed: Parsed | null) => void;
};

export function AddExpenseModal({
  parseInput,
  onParseInputChange,
  onCancel,
  onSave,
}: Props) {
  const parsed = useMemo(() => parsePlainLanguage(parseInput), [parseInput]);
  const [manualBucket, setManualBucket] = useState<Bucket>(parsed?.bucket ?? "lifestyle");

  return (
    <Modal open onClose={onCancel} title="Add expense" size="sm">
      <Text size="xs" color="muted" className="mb-2 uppercase tracking-widest">
        Plain language
      </Text>
      <Input
        autoFocus
        value={parseInput}
        onChange={(e) => onParseInputChange(e.target.value)}
        placeholder='e.g. "lunch 120k", "flight 8M #ustravel2026"'
      />

      <div className="mt-3.5 rounded-[var(--radius-md)] border border-dashed border-line bg-paper-sunken p-3.5">
        <Text size="xs" color="muted" className="mb-2 uppercase tracking-widest">
          Drunken Dolphin parsed this as
        </Text>
        {parsed ? (
          <div>
            <div className="mb-1 flex items-baseline justify-between">
              <span className="text-base text-ink">{parsed.text}</span>
              <span className="font-mono-tabular text-xl font-medium text-ink">
                {fmtVND(parsed.amount)}
                <span className="text-sm text-ink-soft"> ₫</span>
              </span>
            </div>
            <div className="mt-2 flex gap-2">
              <BucketChip bucket={parsed.bucket} />
              <Badge variant="outline" size="sm" className="normal-case">
                {parsed.category}
              </Badge>
              {parsed.tag && (
                <Badge variant="outline" color="accent" size="sm" className="normal-case">
                  {parsed.tag}
                </Badge>
              )}
            </div>
          </div>
        ) : (
          <Text size="sm" color="muted">
            Couldn&apos;t parse. Use form below.
          </Text>
        )}
      </div>

      <div className="my-4 flex items-center gap-2.5">
        <Divider className="flex-1" />
        <span className="font-mono-tabular text-[9px] tracking-[0.18em] text-ink-soft">
          OR USE FORM
        </span>
        <Divider className="flex-1" />
      </div>

      <div className="grid grid-cols-[1fr_130px] gap-2.5">
        <Input placeholder="Note" defaultValue={parsed?.text || ""} />
        <Input
          className="font-mono-tabular"
          placeholder="Amount"
          defaultValue={parsed?.amount || ""}
        />
      </div>
      <div className="mt-2.5 grid grid-cols-3 gap-2.5">
        <Dropdown
          trigger={
            <Button variant="secondary" size="sm" className="w-full justify-between">
              {BUCKET_OPTIONS.find((b) => b.value === manualBucket)?.label}
            </Button>
          }
          items={BUCKET_OPTIONS}
          onSelect={(value) => setManualBucket(value as Bucket)}
        />
        <Input placeholder="Category" defaultValue={parsed?.category || ""} />
        <Input placeholder="#tag (optional)" defaultValue={parsed?.tag || ""} />
      </div>

      <div className="mt-5 flex justify-end gap-2">
        <Button variant="ghost" size="sm" onClick={onCancel}>
          Cancel
        </Button>
        <Button variant="primary" size="sm" onClick={() => onSave(parsed)} disabled={!parsed}>
          Save expense
        </Button>
      </div>
    </Modal>
  );
}
