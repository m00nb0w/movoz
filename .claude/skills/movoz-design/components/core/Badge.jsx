import React from 'react';

/**
 * Movoz Badge / Tag — a small status or category marker.
 * `tone` picks the muted semantic palette; `outline` for a hairline-only chip.
 */
export function Badge({
  children,
  tone = 'neutral',
  outline = false,
  style = {},
  ...rest
}) {
  const tones = {
    neutral:    { fg: 'var(--ink-2)',          bg: 'var(--pencil-light)', bd: 'var(--line)' },
    accent:     { fg: 'var(--terracotta-ink)', bg: 'var(--terracotta-soft)', bd: 'var(--terracotta)' },
    success:    { fg: 'var(--success)',        bg: 'var(--success-soft)', bd: 'var(--success)' },
    warning:    { fg: 'var(--warning)',        bg: 'var(--warning-soft)', bd: 'var(--warning)' },
    danger:     { fg: 'var(--danger)',         bg: 'var(--danger-soft)',  bd: 'var(--danger)' },
    info:       { fg: 'var(--info)',           bg: 'var(--info-soft)',    bd: 'var(--info)' },
  };
  const t = tones[tone] || tones.neutral;
  const base = {
    display: 'inline-flex',
    alignItems: 'center',
    gap: '5px',
    fontFamily: 'var(--font-sans)',
    fontWeight: 'var(--fw-semibold)',
    fontSize: 'var(--text-2xs)',
    letterSpacing: 'var(--ls-wide)',
    textTransform: 'uppercase',
    lineHeight: 1,
    padding: '5px 9px',
    borderRadius: 'var(--radius-sm)',
    color: t.fg,
    background: outline ? 'transparent' : t.bg,
    border: `var(--border-hair) solid ${t.bd}`,
    ...style,
  };
  return <span style={base} {...rest}>{children}</span>;
}
