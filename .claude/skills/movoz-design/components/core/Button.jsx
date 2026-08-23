import React from 'react';

/**
 * Movoz Button — a felt-tip outlined control.
 * Primary = ink fill, secondary = paper with ink hairline, ghost = bare.
 * The "lift" variant carries the hard sketch drop-shadow.
 */
export function Button({
  children,
  variant = 'secondary',
  size = 'md',
  lift = false,
  iconRight = null,
  iconLeft = null,
  disabled = false,
  type = 'button',
  onClick,
  style = {},
  ...rest
}) {
  const sizes = {
    sm: { padding: '6px 12px', fontSize: 'var(--text-sm)', gap: '6px' },
    md: { padding: '10px 18px', fontSize: 'var(--text-md)', gap: '8px' },
    lg: { padding: '14px 26px', fontSize: 'var(--text-lg)', gap: '10px' },
  };

  const variants = {
    primary: {
      background: 'var(--ink)',
      color: 'var(--paper-raised)',
      border: 'var(--border-width) solid var(--ink)',
    },
    secondary: {
      background: 'var(--paper-raised)',
      color: 'var(--ink)',
      border: 'var(--border-width) solid var(--ink)',
    },
    accent: {
      background: 'var(--terracotta)',
      color: '#fff',
      border: 'var(--border-width) solid var(--terracotta)',
    },
    ghost: {
      background: 'transparent',
      color: 'var(--ink)',
      border: 'var(--border-width) solid transparent',
    },
  };

  const base = {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    gap: sizes[size].gap,
    fontFamily: 'var(--font-marker)',
    fontWeight: 'var(--fw-semibold)',
    fontSize: sizes[size].fontSize,
    lineHeight: 1,
    padding: sizes[size].padding,
    borderRadius: 'var(--radius-md)',
    cursor: disabled ? 'not-allowed' : 'pointer',
    opacity: disabled ? 0.45 : 1,
    boxShadow: lift ? 'var(--shadow-sketch)' : 'none',
    transition: 'transform var(--dur-fast) var(--ease-out), box-shadow var(--dur-fast) var(--ease-out), background var(--dur-fast)',
    transform: 'translate(0,0)',
    ...variants[variant],
    ...style,
  };

  const onDown = (e) => { if (!disabled && lift) { e.currentTarget.style.transform = 'translate(3px,3px)'; e.currentTarget.style.boxShadow = 'none'; } };
  const onUp = (e) => { if (!disabled && lift) { e.currentTarget.style.transform = 'translate(0,0)'; e.currentTarget.style.boxShadow = 'var(--shadow-sketch)'; } };

  return (
    <button
      type={type}
      disabled={disabled}
      onClick={onClick}
      onMouseDown={onDown}
      onMouseUp={onUp}
      onMouseLeave={onUp}
      style={base}
      {...rest}
    >
      {iconLeft}
      {children}
      {iconRight}
    </button>
  );
}
