import * as React from 'react';

export interface LogoProps {
  /** Wordmark text. @default 'Movoz' */
  wordmark?: string;
  /** Single letter inside the boxed mark. @default 'M' */
  mark?: string;
  /** Tiny caps line under the wordmark, e.g. "LOGO + WORDMARK". @default '' */
  sublabel?: string;
  /** @default 'md' */
  size?: 'sm' | 'md' | 'lg';
  /** Invert for dark surfaces (the tool dock). @default false */
  onDark?: boolean;
  style?: React.CSSProperties;
}

/** Boxed marker initial + italic wordmark — the Movoz lockup. */
export function Logo(props: LogoProps): JSX.Element;
