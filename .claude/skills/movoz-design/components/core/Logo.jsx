import React from 'react';

/**
 * Movoz Logo — boxed marker initial + wordmark, as drawn in the wireframe
 * header. The mark is an outlined square holding a serif-italic "M".
 */
export function Logo({
  wordmark = 'Movoz',
  mark = 'M',
  sublabel = '',
  size = 'md',
  onDark = false,
  style = {},
  ...rest
}) {
  const sizes = {
    sm: { box: 28, font: 'var(--text-md)', word: 'var(--text-lg)' },
    md: { box: 40, font: 'var(--text-xl)', word: 'var(--text-2xl)' },
    lg: { box: 56, font: 'var(--text-2xl)', word: 'var(--text-3xl)' },
  };
  const s = sizes[size];
  const fg = onDark ? 'var(--paper)' : 'var(--ink)';
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: '12px', ...style }} {...rest}>
      <span
        style={{
          width: `${s.box}px`,
          height: `${s.box}px`,
          border: `var(--border-width-2) solid ${fg}`,
          borderRadius: 'var(--radius-sm)',
          display: 'inline-flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontFamily: 'var(--font-marker)',
          fontStyle: 'italic',
          fontWeight: 'var(--fw-bold)',
          fontSize: s.font,
          color: fg,
          flex: '0 0 auto',
        }}
      >
        {mark}
      </span>
      <span style={{ display: 'flex', flexDirection: 'column', lineHeight: 1 }}>
        <span
          style={{
            fontFamily: 'var(--font-marker)',
            fontStyle: 'italic',
            fontWeight: 'var(--fw-bold)',
            fontSize: s.word,
            color: fg,
          }}
        >
          {wordmark}
        </span>
        {sublabel && (
          <span
            style={{
              fontFamily: 'var(--font-sans)',
              fontSize: 'var(--text-3xs)',
              letterSpacing: 'var(--ls-label)',
              textTransform: 'uppercase',
              color: onDark ? 'var(--pencil)' : 'var(--ink-faint)',
              marginTop: '4px',
            }}
          >
            {sublabel}
          </span>
        )}
      </span>
    </span>
  );
}
