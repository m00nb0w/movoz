export const palette = {
  accent: {
    DEFAULT: "#C25A36",
    light: "#E07A50",
    dark: "#A1462A",
    soft: "#EFDBD0",
  },
} as const;

export const semantic = {
  light: {
    paper: "#EFE9DC",
    paperRaised: "#FAF5EA",
    paperSunken: "#E6DECE",
    ink: "#25231E",
    inkSoft: "#6B675F",
    line: "#D4D0C6",
    pencil: "#C7C3B9",
    pencilLight: "#DEDAD0",
    dock: "#2A2823",
    dockLine: "#45413A",
  },
  dark: {
    paper: "#21201B",
    paperRaised: "#2B2A24",
    paperSunken: "#1A1916",
    ink: "#F2EEE4",
    inkSoft: "#AFA99B",
    line: "#3B382F",
    pencil: "#4B473D",
    pencilLight: "#38362F",
    dock: "#F2EEE4",
    dockLine: "#D8D3C6",
  },
} as const;

export type SemanticColorKey = keyof (typeof semantic)["light"];
export type ColorMode = keyof typeof semantic;
