import * as React from 'react';

export interface PlaceholderBoxProps {
  /** Centered italic label, e.g. "cover image". @default '' */
  label?: string;
  /** CSS aspect-ratio when no fixed height. @default '16 / 9' */
  ratio?: string;
  /** Draw the diagonal X. @default true */
  cross?: boolean;
  /** Dashed instead of solid border. @default false */
  dashed?: boolean;
  /** Fixed height (overrides ratio). */
  height?: string | number | null;
  style?: React.CSSProperties;
}

/** The crossed image/content placeholder frame used throughout Movoz wireframes. */
export function PlaceholderBox(props: PlaceholderBoxProps): JSX.Element;
