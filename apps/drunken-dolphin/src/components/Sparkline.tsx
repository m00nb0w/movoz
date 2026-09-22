type SparklineProps = {
  data: number[];
  width?: number;
  height?: number;
  stroke?: string;
  fill?: string;
  marker?: number | null;
};

export function Sparkline({
  data,
  width = 220,
  height = 44,
  stroke = "var(--ink)",
  fill,
  marker = null,
}: SparklineProps) {
  if (!data?.length) return null;
  const max = Math.max(...data);
  const min = Math.min(...data);
  const range = Math.max(max - min, 1);
  const stepX = width / (data.length - 1 || 1);
  const pts: [number, number][] = data.map((v, i) => [
    i * stepX,
    height - ((v - min) / range) * (height - 6) - 3,
  ]);
  const d = pts.map(([x, y], i) => (i ? `L${x},${y}` : `M${x},${y}`)).join(" ");
  const dArea = `${d} L${width},${height} L0,${height} Z`;
  return (
    <svg width={width} height={height} style={{ display: "block" }}>
      {fill && <path d={dArea} fill={fill} opacity="0.3" />}
      <path
        d={d}
        fill="none"
        stroke={stroke}
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      {marker !== null && pts[marker] && (
        <circle cx={pts[marker][0]} cy={pts[marker][1]} r="3" fill={stroke} />
      )}
    </svg>
  );
}
