interface StatTileProps {
  label: string;
  value: number | string;
}

export function StatTile({ label, value }: StatTileProps) {
  return (
    <div className="border border-line bg-paper-sunken p-4">
      <div className="text-3xl font-bold text-ink">{value}</div>
      <div className="text-sm text-ink-soft">{label}</div>
    </div>
  );
}
