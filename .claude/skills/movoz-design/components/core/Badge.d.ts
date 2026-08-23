import * as React from 'react';

export interface BadgeProps {
  children?: React.ReactNode;
  /** Muted semantic palette. @default 'neutral' */
  tone?: 'neutral' | 'accent' | 'success' | 'warning' | 'danger' | 'info';
  /** Hairline-only, transparent fill. @default false */
  outline?: boolean;
  style?: React.CSSProperties;
}

/** Small uppercase status / category marker in the grotesque label font. */
export function Badge(props: BadgeProps): JSX.Element;
