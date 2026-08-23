import React from 'react';

/**
 * Movoz Skeleton — the soft pencil-gray placeholder bars that stand in for copy.
 * Compose several `Skeleton` lines, or use `lines` for a quick paragraph.
 */
export function Skeleton({
  width = '100%',
  height = 12,
  lines = 1,
  gap = 10,
  rounded = true,
  style = {},
  ...rest
}) {
  const bar = (w, key) => (
    <div
      key={key}
      style={{
        width: w,
        height: typeof height === 'number' ? `${height}px` : height,
        background: 'var(--pencil)',
        borderRadius: rounded ? 'var(--radius-pill)' : 'var(--radius-xs)',
      }}
    />
  );

  if (lines <= 1) {
    return <div style={{ ...style }} {...rest}>{bar(width, 0)}</div>;
  }

  // Last line is shorter, like real ragged text.
  const widths = Array.from({ length: lines }, (_, i) =>
    i === lines - 1 ? '62%' : (typeof width === 'string' ? width : `${width}px`)
  );
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: `${gap}px`, ...style }} {...rest}>
      {widths.map((w, i) => bar(w, i))}
    </div>
  );
}
