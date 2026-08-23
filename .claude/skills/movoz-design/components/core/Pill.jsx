import React from 'react';

/**
 * Movoz Pill — the rounded filter / chip control from the wireframe's filter row.
 * Active pills get a soft pencil fill; inactive are outlined.
 */
export function Pill({
  children,
  active = false,
  as = 'button',
  onClick,
  style = {},
  ...rest
}) {
  const Tag = as;
  const base = {
    display: 'inline-flex',
    alignItems: 'center',
    gap: '6px',
    fontFamily: 'var(--font-marker)',
    fontWeight: 'var(--fw-medium)',
    fontSize: 'var(--text-md)',
    lineHeight: 1,
    padding: '8px 18px',
    borderRadius: 'var(--radius-pill)',
    cursor: 'pointer',
    color: 'var(--ink)',
    background: active ? 'var(--pencil-light)' : 'var(--paper-raised)',
    border: `var(--border-width) solid ${active ? 'var(--line-strong)' : 'var(--line)'}`,
    transition: 'background var(--dur-fast), border-color var(--dur-fast)',
    ...style,
  };
  return (
    <Tag onClick={onClick} style={base} {...rest}>
      {children}
    </Tag>
  );
}
