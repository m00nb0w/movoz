import React from 'react';

/**
 * Movoz PlaceholderBox — the crossed "cover image / code sample" frame from the
 * wireframe. A dashed-or-solid rectangle with an X and an optional centered label.
 * Use it anywhere real imagery/content isn't available yet.
 */
export function PlaceholderBox({
  label = '',
  ratio = '16 / 9',
  cross = true,
  dashed = false,
  height = null,
  style = {},
  ...rest
}) {
  const wrap = {
    position: 'relative',
    width: '100%',
    aspectRatio: height ? undefined : ratio,
    height: height || undefined,
    background: 'var(--paper-raised)',
    border: `var(--border-width) ${dashed ? 'dashed' : 'solid'} var(--line-strong)`,
    borderRadius: 'var(--radius-sm)',
    overflow: 'hidden',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    color: 'var(--ink-faint)',
    fontFamily: 'var(--font-marker)',
    fontStyle: 'italic',
    fontSize: 'var(--text-md)',
    ...style,
  };
  const line = {
    position: 'absolute',
    top: 0, left: 0, right: 0, bottom: 0,
    color: 'var(--pencil)',
  };
  return (
    <div style={wrap} {...rest}>
      {cross && (
        <svg style={line} preserveAspectRatio="none" viewBox="0 0 100 100">
          <line x1="0" y1="0" x2="100" y2="100" stroke="currentColor" strokeWidth="0.6" vectorEffect="non-scaling-stroke" />
          <line x1="100" y1="0" x2="0" y2="100" stroke="currentColor" strokeWidth="0.6" vectorEffect="non-scaling-stroke" />
        </svg>
      )}
      {label && <span style={{ position: 'relative', background: 'var(--paper-raised)', padding: '0 8px' }}>{label}</span>}
    </div>
  );
}
