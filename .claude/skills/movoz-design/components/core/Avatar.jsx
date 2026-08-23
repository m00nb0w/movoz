import React from 'react';

/**
 * Movoz Avatar — a circular outlined frame. Shows an image, initials, or an
 * empty pencil outline (the wireframe's bare author bubble).
 */
export function Avatar({
  src = null,
  initials = '',
  size = 40,
  style = {},
  ...rest
}) {
  const base = {
    width: `${size}px`,
    height: `${size}px`,
    borderRadius: '50%',
    border: 'var(--border-width) solid var(--line-strong)',
    background: src ? `center/cover no-repeat url(${src})` : 'var(--paper-raised)',
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    color: 'var(--ink-soft)',
    fontFamily: 'var(--font-marker)',
    fontWeight: 'var(--fw-semibold)',
    fontSize: `${Math.round(size * 0.4)}px`,
    flex: '0 0 auto',
    overflow: 'hidden',
    ...style,
  };
  return (
    <span style={base} {...rest}>
      {!src && initials}
    </span>
  );
}
