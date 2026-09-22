import { Badge } from "@movoz/ui-web";
import type { Bucket } from "@/data/mock";

const LABELS: Record<Bucket, string> = {
  essentials: "Essentials",
  lifestyle: "Lifestyle",
  irregular: "Irregular",
};

export const BUCKET_COLOR: Record<Bucket, string> = {
  essentials: "var(--bucket-essentials)",
  lifestyle: "var(--bucket-lifestyle)",
  irregular: "var(--bucket-irregular)",
};

export function BucketDot({ bucket, size = 8 }: { bucket: Bucket; size?: number }) {
  return (
    <span
      className="inline-block flex-none rounded-full"
      style={{ width: size, height: size, backgroundColor: BUCKET_COLOR[bucket] }}
    />
  );
}

export function BucketChip({ bucket, label }: { bucket: Bucket; label?: string }) {
  return (
    <Badge variant="outline" color="default" size="sm" className="gap-1.5 normal-case">
      <BucketDot bucket={bucket} size={7} />
      {label || LABELS[bucket]}
    </Badge>
  );
}
