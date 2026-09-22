type HBarItem = { label: string; value: number; color: string };

type HBarProps = {
  items: HBarItem[];
  maxValue?: number;
  valueFmt?: (v: number) => string;
};

export function HBar({ items, maxValue, valueFmt = (v) => String(v) }: HBarProps) {
  const max = maxValue ?? Math.max(...items.map((i) => i.value));
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 9 }}>
      {items.map((it, i) => (
        <div key={i} style={{ display: "flex", alignItems: "center", gap: 10 }}>
          <div
            style={{
              width: 110,
              fontSize: 12,
              color: "var(--ink)",
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
            }}
          >
            {it.label}
          </div>
          <div
            style={{
              flex: 1,
              height: 14,
              background: "var(--line)",
              borderRadius: 2,
              position: "relative",
            }}
          >
            <div
              style={{
                position: "absolute",
                inset: 0,
                width: `${(it.value / max) * 100}%`,
                background: it.color,
                borderRadius: 2,
              }}
            />
          </div>
          <div
            className="font-mono-tabular"
            style={{
              fontSize: 11,
              color: "var(--ink-soft)",
              minWidth: 60,
              textAlign: "right",
            }}
          >
            {valueFmt(it.value)}
          </div>
        </div>
      ))}
    </div>
  );
}
