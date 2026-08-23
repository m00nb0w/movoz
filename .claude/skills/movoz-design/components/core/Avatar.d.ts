import * as React from 'react';

export interface AvatarProps {
  src?: string | null;
  /** Fallback initials when no image. */
  initials?: string;
  /** Diameter in px. @default 40 */
  size?: number;
  style?: React.CSSProperties;
}

/** Circular outlined author/user bubble — image, initials, or bare pencil ring. */
export function Avatar(props: AvatarProps): JSX.Element;
