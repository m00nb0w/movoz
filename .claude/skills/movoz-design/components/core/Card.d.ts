import * as React from 'react';

export interface CardProps {
  children?: React.ReactNode;
  /** Hard offset sketch drop-shadow. @default false */
  lift?: boolean;
  /** Raise + darken border on hover (for clickable cards). @default false */
  interactive?: boolean;
  /** @default 'var(--space-5)' */
  padding?: string;
  as?: any;
  style?: React.CSSProperties;
}

/**
 * Paper sheet with a felt-tip hairline border — the base surface for article
 * cards, panels, and tiles.
 * @startingPoint section="Core" subtitle="Paper sheet surface" viewport="700x150"
 */
export function Card(props: CardProps): JSX.Element;
