export function TrendLine({
  values,
  color = "var(--terracotta)",
  height = 72,
}: {
  values: number[];
  color?: string;
  height?: number;
}) {
  const width = 300;
  const max = Math.max(1, ...values);
  const stepX = values.length > 1 ? width / (values.length - 1) : width;
  const points = values
    .map((v, i) => `${i * stepX},${height - (v / max) * height}`)
    .join(" ");

  return (
    <svg viewBox={`0 0 ${width} ${height}`} className="h-16 w-full" preserveAspectRatio="none">
      <polyline points={points} fill="none" stroke={color} strokeWidth={2} strokeLinejoin="round" />
    </svg>
  );
}
