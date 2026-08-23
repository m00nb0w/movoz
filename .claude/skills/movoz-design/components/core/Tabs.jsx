import React from 'react';

/**
 * Movoz Tabs — the segmented control from the wireframe's tool dock.
 * `tone="dock"` renders the dark pill-bar; `tone="light"` for in-page tabs.
 */
export function Tabs({
  items = [],
  value,
  onChange,
  tone = 'light',
  style = {},
  ...rest
}) {
  const dark = tone === 'dock';
  const bar = {
    display: 'inline-flex',
    alignItems: 'center',
    gap: '6px',
    padding: '6px',
    borderRadius: 'var(--radius-lg)',
    background: dark ? 'var(--dock)' : 'var(--paper-raised)',
    border: `var(--border-width) solid ${dark ? 'var(--dock-line)' : 'var(--line)'}`,
    ...style,
  };
  return (
    <div style={bar} role="tablist" {...rest}>
      {items.map((it) => {
        const key = typeof it === 'string' ? it : it.value;
        const label = typeof it === 'string' ? it : it.label;
        const accent = typeof it === 'object' && it.accent;
        const active = key === value;
        const tab = {
          fontFamily: 'var(--font-marker)',
          fontWeight: 'var(--fw-semibold)',
          fontSize: 'var(--text-md)',
          lineHeight: 1,
          padding: '8px 16px',
          borderRadius: 'var(--radius-md)',
          cursor: 'pointer',
          border: 'var(--border-hair) solid transparent',
          transition: 'background var(--dur-fast), color var(--dur-fast)',
          background: active
            ? (accent ? 'var(--terracotta)' : (dark ? 'var(--paper)' : 'var(--pencil-light)'))
            : 'transparent',
          color: active
            ? (accent ? '#fff' : (dark ? 'var(--ink)' : 'var(--ink)'))
            : (accent ? 'var(--terracotta)' : (dark ? 'var(--pencil)' : 'var(--ink-soft)')),
          borderColor: !active && accent ? 'var(--terracotta)' : 'transparent',
        };
        return (
          <button key={key} role="tab" aria-selected={active} onClick={() => onChange && onChange(key)} style={tab}>
            {label}
          </button>
        );
      })}
    </div>
  );
}
