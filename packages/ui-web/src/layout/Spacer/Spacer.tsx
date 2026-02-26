"use client";

import { cn } from "../../utils/cn";

export interface SpacerProps {
  size?: number;
  flex?: boolean;
  className?: string;
}

const sizeMap: Record<number, string> = {
  1: "h-1 w-1",
  2: "h-2 w-2",
  3: "h-3 w-3",
  4: "h-4 w-4",
  5: "h-5 w-5",
  6: "h-6 w-6",
  8: "h-8 w-8",
  10: "h-10 w-10",
  12: "h-12 w-12",
  16: "h-16 w-16",
  20: "h-20 w-20",
  24: "h-24 w-24",
};

export function Spacer({ size, flex = false, className }: SpacerProps) {
  if (flex) {
    return <div className={cn("flex-1", className)} />;
  }

  return (
    <div
      className={cn(size && sizeMap[size], className)}
      aria-hidden="true"
    />
  );
}
