const STATUS_ORDER = ["backlog", "picked_for_today", "in_progress", "blocked", "done"] as const;

const STATUS_LABELS: Record<(typeof STATUS_ORDER)[number], string> = {
  backlog: "Backlog",
  picked_for_today: "Picked",
  in_progress: "In progress",
  blocked: "Blocked",
  done: "Done",
};

export function StatusBreakdownChart({
  breakdown,
  color = "var(--terracotta)",
  height = 120,
}: {
  breakdown: Record<string, number>;
  color?: string;
  height?: number;
}) {
  const width = 300;
  const barGap = 12;
  const barWidth = (width - barGap * (STATUS_ORDER.length - 1)) / STATUS_ORDER.length;
  const counts = STATUS_ORDER.map((status) => breakdown[status] ?? 0);
  const max = Math.max(1, ...counts);
  const labelHeight = 16;
  const chartHeight = height - labelHeight;

  return (
    <svg
      viewBox={`0 0 ${width} ${height}`}
      className="h-32 w-full"
      preserveAspectRatio="none"
      role="img"
      aria-label="Item count by status"
    >
      {STATUS_ORDER.map((status, i) => {
        const count = counts[i];
        const barHeight = (count / max) * chartHeight;
        const x = i * (barWidth + barGap);
        const y = chartHeight - barHeight;

        return (
          <g key={status}>
            <rect x={x} y={y} width={barWidth} height={Math.max(barHeight, 1)} fill={color} />
            <text
              x={x + barWidth / 2}
              y={y - 4}
              textAnchor="middle"
              fontSize={10}
              fill="var(--ink)"
            >
              {count}
            </text>
            <text
              x={x + barWidth / 2}
              y={height - 2}
              textAnchor="middle"
              fontSize={8}
              fill="var(--ink-soft)"
            >
              {STATUS_LABELS[status]}
            </text>
          </g>
        );
      })}
    </svg>
  );
}
