import { type ComponentPropsWithoutRef } from "react";
import { cn } from "../../utils/cn";

export interface SkeletonProps extends ComponentPropsWithoutRef<"div"> {
  lines?: number;
  width?: string | number;
  height?: string | number;
}

export function Skeleton({
  lines = 1,
  width = "100%",
  height = 12,
  className,
  style,
  ...props
}: SkeletonProps) {
  const barHeight = typeof height === "number" ? `${height}px` : height;
  const widths = Array.from({ length: lines }, (_, i) =>
    lines > 1 && i === lines - 1
      ? "62%"
      : typeof width === "number"
        ? `${width}px`
        : width
  );

  return (
    <div className={cn("flex flex-col gap-2.5", className)} style={style} {...props}>
      {widths.map((w, i) => (
        <div key={i} className="rounded-full bg-pencil" style={{ width: w, height: barHeight }} />
      ))}
    </div>
  );
}
