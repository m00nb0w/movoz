// src/colors.ts
var palette = {
  accent: {
    DEFAULT: "#C25A36",
    light: "#E07A50",
    dark: "#A1462A"
  }
};
var semantic = {
  light: {
    paper: "#EFE9DC",
    paperRaised: "#FAF5EA",
    paperSunken: "#E6DECE",
    ink: "#25231E",
    inkSoft: "#6B675F",
    line: "#D4D0C6"
  },
  dark: {
    paper: "#21201B",
    paperRaised: "#2B2A24",
    paperSunken: "#1A1916",
    ink: "#F2EEE4",
    inkSoft: "#AFA99B",
    line: "#3B382F"
  }
};

// src/typography.ts
var fontFamilies = {
  marker: ["Shantell Sans", "Comic Sans MS", "cursive"],
  body: ["Space Grotesk", "ui-sans-serif", "system-ui", "sans-serif"]
};
var fontSizes = {
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
  "7xl": "4.5rem"
};
var fontWeights = {
  light: 300,
  normal: 400,
  medium: 500,
  semibold: 600,
  bold: 700,
  extrabold: 800
};
var lineHeights = {
  none: 1,
  tight: 1.1,
  snug: 1.3,
  normal: 1.5,
  relaxed: 1.6,
  loose: 1.7
};
var letterSpacings = {
  tighter: "-0.02em",
  tight: "-0.01em",
  normal: "0",
  wide: "0.01em"
};

// src/spacing.ts
var spacing = {
  0: "0",
  0.5: "0.125rem",
  1: "0.25rem",
  1.5: "0.375rem",
  2: "0.5rem",
  2.5: "0.625rem",
  3: "0.75rem",
  4: "1rem",
  5: "1.25rem",
  6: "1.5rem",
  8: "2rem",
  10: "2.5rem",
  12: "3rem",
  16: "4rem",
  20: "5rem",
  24: "6rem",
  32: "8rem"
};

// src/radii.ts
var radii = {
  none: "0",
  sm: "7px",
  DEFAULT: "11px",
  md: "11px",
  lg: "16px",
  xl: "22px",
  "2xl": "30px",
  full: "9999px",
  sketch: "14px 11px 13px 12px"
};

// src/shadows.ts
var shadows = {
  none: "none",
  sm: "0 1px 2px rgba(37, 35, 30, 0.05)",
  DEFAULT: "0 1px 3px rgba(37, 35, 30, 0.05), 0 4px 12px rgba(37, 35, 30, 0.04)",
  md: "0 4px 6px -1px rgba(37, 35, 30, 0.07), 0 2px 4px -2px rgba(37, 35, 30, 0.05)",
  lg: "0 10px 15px -3px rgba(37, 35, 30, 0.08), 0 4px 6px -4px rgba(37, 35, 30, 0.04)",
  xl: "0 20px 25px -5px rgba(37, 35, 30, 0.1), 0 8px 10px -6px rgba(37, 35, 30, 0.04)",
  "2xl": "0 20px 40px -12px rgba(37, 35, 30, 0.15)",
  sketch: "3px 3px 0 var(--ink)",
  sketchAccent: "3px 3px 0 var(--terracotta)"
};
var darkShadows = {
  DEFAULT: "0 1px 3px rgba(0, 0, 0, 0.2), 0 4px 12px rgba(0, 0, 0, 0.15)",
  "2xl": "0 20px 40px -12px rgba(0, 0, 0, 0.4)"
};

// src/breakpoints.ts
var breakpoints = {
  sm: "640px",
  md: "768px",
  lg: "1024px",
  xl: "1280px",
  "2xl": "1536px"
};

// src/animation.ts
var durations = {
  fast: "150ms",
  normal: "200ms",
  slow: "300ms",
  slower: "600ms"
};
var easings = {
  ease: "ease",
  easeIn: "ease-in",
  easeOut: "ease-out",
  easeInOut: "ease-in-out",
  cubic: "cubic-bezier(0.4, 0, 0.2, 1)"
};
var keyframes = {
  fadeIn: {
    "0%": { opacity: "0" },
    "100%": { opacity: "1" }
  },
  slideUp: {
    "0%": { opacity: "0", transform: "translateY(20px)" },
    "100%": { opacity: "1", transform: "translateY(0)" }
  },
  slideInRight: {
    "0%": { opacity: "0", transform: "translateX(-20px)" },
    "100%": { opacity: "1", transform: "translateX(0)" }
  }
};
var animations = {
  "fade-in": "fadeIn 0.6s ease-out forwards",
  "slide-up": "slideUp 0.6s ease-out forwards",
  "slide-in-right": "slideInRight 0.6s ease-out forwards"
};
export {
  animations,
  breakpoints,
  darkShadows,
  durations,
  easings,
  fontFamilies,
  fontSizes,
  fontWeights,
  keyframes,
  letterSpacings,
  lineHeights,
  palette,
  radii,
  semantic,
  shadows,
  spacing
};
