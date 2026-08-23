import * as React from 'react';

export interface InputProps {
  value?: string;
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void;
  placeholder?: string;
  type?: string;
  /** Leading icon node (e.g. a search glyph). */
  iconLeft?: React.ReactNode;
  /** Fully rounded ends. @default false */
  pill?: boolean;
  disabled?: boolean;
  /** @default 'md' */
  size?: 'sm' | 'md' | 'lg';
  style?: React.CSSProperties;
  wrapStyle?: React.CSSProperties;
}

/** Felt-tip outlined text field with an optional leading icon. */
export function Input(props: InputProps): JSX.Element;
