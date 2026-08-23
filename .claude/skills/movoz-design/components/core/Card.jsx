import React from 'react';

/**
 * Movoz Card — a paper sheet with a felt-tip hairline border.
 * `lift` adds the hard sketch drop-shadow; `interactive` raises it on hover.
 */
export function Card({
  children,
  lift = false,
  interactive = false,
  padding = 'var(--space-5)',
  as = 'div',
  style = {},
  ...rest
}) {
  const Tag = as;
  const [hover, setHover] = React.useState(false);
  const base = {
    background: 'var(--paper-raised)',
    border: 'var(--border-width) solid var(--line)',
    borderRadius: 'var(--radius-md)',
    padding,
    boxShadow: lift ? 'var(--shadow-sketch)' : 'var(--shadow-none)',
    transition: 'transform var(--dur-base) var(--ease-out), box-shadow var(--dur-base) var(--ease-out), border-color var(--dur-base)',
    transform: interactive && hover ? 'translate(-2px,-2px)' : 'translate(0,0)',
    ...(interactive && hover
      ? { boxShadow: 'var(--shadow-sketch)', borderColor: 'var(--ink)' }
      : null),
    cursor: interactive ? 'pointer' : 'default',
    ...style,
  };
  return (
    <Tag
      style={base}
      onMouseEnter={() => interactive && setHover(true)}
      onMouseLeave={() => interactive && setHover(false)}
      {...rest}
    >
      {children}
    </Tag>
  );
}
