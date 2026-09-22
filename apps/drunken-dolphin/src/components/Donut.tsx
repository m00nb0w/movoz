type Segment = { value: number; color: string };

type DonutProps = {
  segments: Segment[];
  size?: number;
  thickness?: number;
  label?: string;
  sub?: string;
};

export function Donut({
  segments,
  size = 180,
  thickness = 22,
  label,
  sub,
}: DonutProps) {
  const cx = size / 2;
  const cy = size / 2;
  const r = (size - thickness) / 2;
  const c = 2 * Math.PI * r;
  const total = segments.reduce((s, x) => s + x.value, 0);
  let offset = 0;

  return (
    <svg width={size} height={size} style={{ display: "block" }}>
      <circle
        cx={cx}
        cy={cy}
        r={r}
        fill="none"
        stroke="var(--line)"
        strokeWidth={thickness}
      />
      {segments.map((s, i) => {
        const len = (s.value / total) * c;
        const dasharray = `${len} ${c - len}`;
        const dashoffset = -offset;
        offset += len;
        return (
          <circle
            key={i}
            cx={cx}
            cy={cy}
            r={r}
            fill="none"
            stroke={s.color}
            strokeWidth={thickness}
            strokeDasharray={dasharray}
            strokeDashoffset={dashoffset}
            transform={`rotate(-90 ${cx} ${cy})`}
          />
        );
      })}
      {label && (
        <g>
          <text
            x={cx}
            y={cy - 2}
            textAnchor="middle"
            fontFamily="Shantell Sans, cursive"
            fontSize="22"
            fill="var(--ink)"
          >
            {label}
          </text>
          {sub && (
            <text
              x={cx}
              y={cy + 16}
              textAnchor="middle"
              fontFamily="var(--font-jetbrains)"
              fontSize="9"
              letterSpacing="0.1em"
              fill="var(--ink-soft)"
            >
              {sub}
            </text>
          )}
        </g>
      )}
    </svg>
  );
}
