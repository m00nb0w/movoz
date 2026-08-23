import * as React from 'react';

export interface SkeletonProps {
  /** @default '100%' */
  width?: string | number;
  /** @default 12 */
  height?: string | number;
  /** Number of stacked lines; last line is shortened. @default 1 */
  lines?: number;
  /** Gap between lines in px. @default 10 */
  gap?: number;
  /** Pill-rounded ends. @default true */
  rounded?: boolean;
  style?: React.CSSProperties;
}

/** Pencil-gray placeholder bars standing in for not-yet-real copy. */
export function Skeleton(props: SkeletonProps): JSX.Element;
