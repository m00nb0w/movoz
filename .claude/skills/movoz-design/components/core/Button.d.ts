import * as React from 'react';

export interface ButtonProps {
  children?: React.ReactNode;
  /** Visual style. @default 'secondary' */
  variant?: 'primary' | 'secondary' | 'accent' | 'ghost';
  /** @default 'md' */
  size?: 'sm' | 'md' | 'lg';
  /** Adds the hard offset "sketch" drop-shadow that presses in on click. @default false */
  lift?: boolean;
  iconLeft?: React.ReactNode;
  iconRight?: React.ReactNode;
  disabled?: boolean;
  type?: 'button' | 'submit' | 'reset';
  onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void;
  style?: React.CSSProperties;
}

/**
 * The Movoz button — felt-tip outlined control in the brand marker font.
 * @startingPoint section="Core" subtitle="Outlined felt-tip buttons" viewport="700x150"
 */
export function Button(props: ButtonProps): JSX.Element;
