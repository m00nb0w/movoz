import * as React from 'react';

export interface PillProps {
  children?: React.ReactNode;
  /** Filled "selected" state. @default false */
  active?: boolean;
  /** Element to render. @default 'button' */
  as?: any;
  onClick?: (e: React.MouseEvent) => void;
  style?: React.CSSProperties;
}

/** Rounded filter chip in the marker font; rows of these drive content filtering. */
export function Pill(props: PillProps): JSX.Element;
