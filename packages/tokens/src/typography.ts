export const fontFamilies = {
  sans: ["Rubik", "system-ui", "sans-serif"],
  serif: ["Libre Baskerville", "Georgia", "serif"],
  ui: ["Inter", "system-ui", "sans-serif"],
} as const;

export const fontSizes = {
  xs: "0.75rem",
  sm: "0.875rem",
  base: "1rem",
  lg: "1.125rem",
  xl: "1.25rem",
  "2xl": "1.5rem",
  "3xl": "1.875rem",
  "4xl": "2.25rem",
  "5xl": "3rem",
  "6xl": "3.75rem",
  "7xl": "4.5rem",
} as const;

export const fontWeights = {
  light: 300,
  normal: 400,
  medium: 500,
  semibold: 600,
  bold: 700,
  extrabold: 800,
} as const;

export const lineHeights = {
  none: 1,
  tight: 1.1,
  snug: 1.3,
  normal: 1.5,
  relaxed: 1.6,
  loose: 1.7,
} as const;

export const letterSpacings = {
  tighter: "-0.02em",
  tight: "-0.01em",
  normal: "0",
  wide: "0.01em",
} as const;
