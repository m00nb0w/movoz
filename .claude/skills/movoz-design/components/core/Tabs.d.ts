import * as React from 'react';

export interface TabItem {
  value: string;
  label: React.ReactNode;
  /** Render as a terracotta toggle (e.g. "Annotations", "Grid"). */
  accent?: boolean;
}

export interface TabsProps {
  /** Strings or {value,label,accent} objects. */
  items?: (string | TabItem)[];
  value?: string;
  onChange?: (value: string) => void;
  /** 'light' in-page tabs, or 'dock' for the dark tool bar. @default 'light' */
  tone?: 'light' | 'dock';
  style?: React.CSSProperties;
}

/** Segmented control / view switcher, including the dark wireframe tool dock. */
export function Tabs(props: TabsProps): JSX.Element;
