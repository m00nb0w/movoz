export const palette = {
  accent: {
    DEFAULT: "#d4775c",
    light: "#e08a70",
    dark: "#c46448",
  },
} as const;

export const semantic = {
  light: {
    bg: "#fdf6e3",
    text: "#3d3520",
    muted: "#857a5c",
    subtle: "#f5edd8",
    border: "#e8dfc5",
    paper: "#fffaed",
  },
  dark: {
    bg: "#1a1a1a",
    text: "#e8e4dc",
    muted: "#a09080",
    subtle: "#252525",
    border: "#3a3a3a",
    paper: "#222222",
  },
} as const;

export type SemanticColorKey = keyof (typeof semantic)["light"];
export type ColorMode = keyof typeof semantic;
