import React from 'react';

/**
 * Movoz Input — a felt-tip outlined text field. Optional leading icon (e.g. a
 * search glyph) and pill or default rounding.
 */
export function Input({
  value,
  onChange,
  placeholder = '',
  type = 'text',
  iconLeft = null,
  pill = false,
  disabled = false,
  size = 'md',
  style = {},
  wrapStyle = {},
  ...rest
}) {
  const sizes = {
    sm: { h: 34, fs: 'var(--text-sm)', px: 12 },
    md: { h: 42, fs: 'var(--text-md)', px: 14 },
    lg: { h: 50, fs: 'var(--text-lg)', px: 16 },
  };
  const s = sizes[size];
  const wrap = {
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
    height: `${s.h}px`,
    padding: `0 ${s.px}px`,
    background: 'var(--paper-raised)',
    border: 'var(--border-width) solid var(--line)',
    borderRadius: pill ? 'var(--radius-pill)' : 'var(--radius-md)',
    opacity: disabled ? 0.5 : 1,
    transition: 'border-color var(--dur-fast)',
    ...wrapStyle,
  };
  const input = {
    flex: 1,
    minWidth: 0,
    border: 'none',
    outline: 'none',
    background: 'transparent',
    color: 'var(--ink)',
    fontFamily: 'var(--font-sans)',
    fontSize: s.fs,
    ...style,
  };
  return (
    <label
      style={wrap}
      onFocus={(e) => (e.currentTarget.style.borderColor = 'var(--ink)')}
      onBlur={(e) => (e.currentTarget.style.borderColor = 'var(--line)')}
    >
      {iconLeft && <span style={{ color: 'var(--ink-faint)', display: 'inline-flex' }}>{iconLeft}</span>}
      <input
        type={type}
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        disabled={disabled}
        style={input}
        {...rest}
      />
    </label>
  );
}
