import * as React from 'react';

export interface AnnotationProps {
  children?: React.ReactNode;
  /** Show the leading ↳ hook. @default true */
  arrow?: boolean;
  /** @default 'left' */
  align?: 'left' | 'right';
  style?: React.CSSProperties;
}

/** Terracotta hand-written margin note used to label regions in mockups & specs. */
export function Annotation(props: AnnotationProps): JSX.Element;
