import React from 'react';

/**
 * Movoz Annotation — the terracotta hand-written margin note from the wireframe
 * ("↳ hero — featured / latest post"). Use it to label regions in mockups,
 * blueprints, and spec sheets.
 */
export function Annotation({
  children,
  arrow = true,
  align = 'left',
  style = {},
  ...rest
}) {
  const base = {
    display: 'inline-flex',
    alignItems: 'baseline',
    gap: '7px',
    fontFamily: 'var(--font-marker)',
    fontStyle: 'italic',
    fontWeight: 'var(--fw-medium)',
    fontSize: 'var(--text-md)',
    color: 'var(--terracotta)',
    justifyContent: align === 'right' ? 'flex-end' : 'flex-start',
    ...style,
  };
  return (
    <span style={base} {...rest}>
      {arrow && <span style={{ fontStyle: 'normal' }}>{align === 'right' ? '↳' : '↳'}</span>}
      {children}
    </span>
  );
}
