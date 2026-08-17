interface StatTileProps {
  label: string;
  value: number | string;
}

export function StatTile({ label, value }: StatTileProps) {
  return (
    <div className="border border-zen-border bg-zen-subtle p-4">
      <div className="text-3xl font-bold text-zen-text">{value}</div>
      <div className="text-sm text-zen-muted">{label}</div>
    </div>
  );
}
