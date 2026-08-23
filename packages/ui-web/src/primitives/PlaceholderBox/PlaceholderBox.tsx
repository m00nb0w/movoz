import { cn } from "../../utils/cn";

export interface PlaceholderBoxProps {
  label?: string;
  ratio?: string;
  cross?: boolean;
  dashed?: boolean;
  height?: string | number;
  className?: string;
}

export function PlaceholderBox({
  label = "",
  ratio = "16 / 9",
  cross = true,
  dashed = false,
  height,
  className,
}: PlaceholderBoxProps) {
  return (
    <div
      className={cn(
        "relative flex items-center justify-center overflow-hidden w-full",
        "bg-paper-raised border-line rounded-[var(--radius-sm)]",
        "border-[length:var(--border-width)]",
        dashed ? "border-dashed" : "border-solid",
        "text-ink-soft font-marker italic",
        className
      )}
      style={{ aspectRatio: height ? undefined : ratio, height: height ?? undefined }}
    >
      {cross && (
        <svg
          className="absolute inset-0 text-pencil"
          preserveAspectRatio="none"
          viewBox="0 0 100 100"
        >
          <line x1="0" y1="0" x2="100" y2="100" stroke="currentColor" strokeWidth="0.6" vectorEffect="non-scaling-stroke" />
          <line x1="100" y1="0" x2="0" y2="100" stroke="currentColor" strokeWidth="0.6" vectorEffect="non-scaling-stroke" />
        </svg>
      )}
      {label && <span className="relative bg-paper-raised px-2">{label}</span>}
    </div>
  );
}
