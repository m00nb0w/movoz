import { fmtVND } from "@/lib/format";

type BudgetGaugeProps = {
  spent: number;
  budget: number;
  daysIn: number;
  daysTotal: number;
  size?: number;
};

export function BudgetGauge({
  spent,
  budget,
  daysIn,
  daysTotal,
  size = 240,
}: BudgetGaugeProps) {
  const pct = spent / budget;
  const expectedPct = daysIn / daysTotal;
  const overspendPace = pct > expectedPct + 0.03;

  const r = size * 0.42;
  const cx = size / 2;
  const cy = size * 0.62;
  const c = Math.PI * r;
  const filled = Math.min(pct, 1) * c;
  const expectedAngle = -180 + expectedPct * 180;
  const ex = cx + Math.cos((expectedAngle * Math.PI) / 180) * r;
  const ey = cy + Math.sin((expectedAngle * Math.PI) / 180) * r;
  const remaining = budget - spent;

  return (
    <div style={{ position: "relative", width: size, height: size * 0.74 }}>
      <svg width={size} height={size * 0.74} style={{ display: "block" }}>
        <path
          d={`M ${cx - r} ${cy} A ${r} ${r} 0 0 1 ${cx + r} ${cy}`}
          fill="none"
          stroke="var(--line)"
          strokeWidth="14"
          strokeLinecap="round"
        />
        <path
          d={`M ${cx - r} ${cy} A ${r} ${r} 0 0 1 ${cx + r} ${cy}`}
          fill="none"
          stroke={overspendPace ? "var(--terracotta)" : "var(--ink)"}
          strokeWidth="14"
          strokeLinecap="round"
          strokeDasharray={`${filled} ${c}`}
        />
        <circle cx={ex} cy={ey} r="3" fill="var(--paper)" stroke="var(--ink-soft)" strokeWidth="1.2" />
        <text
          x={cx - r}
          y={cy + 18}
          fontFamily="var(--font-jetbrains)"
          fontSize="9"
          fill="var(--ink-soft)"
          letterSpacing="0.08em"
          textAnchor="middle"
        >
          0
        </text>
        <text
          x={cx + r}
          y={cy + 18}
          fontFamily="var(--font-jetbrains)"
          fontSize="9"
          fill="var(--ink-soft)"
          letterSpacing="0.08em"
          textAnchor="middle"
        >
          {fmtVND(budget)}
        </text>
      </svg>
      <div
        style={{
          position: "absolute",
          top: "40%",
          left: 0,
          right: 0,
          textAlign: "center",
          padding: `0 ${size * 0.18}px`,
        }}
      >
        <div
          className="font-mono-tabular"
          style={{
            fontSize: Math.max(9, size * 0.045),
            color: "var(--ink-soft)",
            letterSpacing: "0.12em",
            textTransform: "uppercase",
            marginBottom: 2,
          }}
        >
          Day {daysIn}/{daysTotal}
        </div>
        <div
          className="font-marker"
          style={{
            fontSize: size * 0.13,
            color: "var(--ink)",
            letterSpacing: "-0.02em",
            lineHeight: 1,
            whiteSpace: "nowrap",
          }}
        >
          {fmtVND(spent)}
          <span
            style={{
              color: "var(--ink-soft)",
              fontSize: size * 0.08,
              marginLeft: 3,
            }}
          >
            ₫
          </span>
        </div>
        <div
          className="font-mono-tabular"
          style={{
            fontSize: Math.max(10, size * 0.048),
            color: overspendPace ? "var(--terracotta)" : "var(--ink-soft)",
            marginTop: 6,
            letterSpacing: "0.04em",
            whiteSpace: "nowrap",
          }}
        >
          {remaining > 0 ? `${fmtVND(remaining)}₫ left` : `${fmtVND(-remaining)}₫ over`} · {Math.round(pct * 100)}%
        </div>
      </div>
    </div>
  );
}
