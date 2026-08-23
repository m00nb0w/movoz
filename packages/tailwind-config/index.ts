import type { Config } from "tailwindcss";
import { fontFamilies, keyframes, animations } from "@movoz/tokens";

const config: Partial<Config> = {
  darkMode: "class",
  theme: {
    extend: {
      fontFamily: {
        sans: [...fontFamilies.body],
        marker: [...fontFamilies.marker],
        body: [...fontFamilies.body],
      },
      colors: {
        accent: {
          DEFAULT: "var(--terracotta)",
          light: "var(--terracotta-light)",
          dark: "var(--terracotta-ink)",
        },
        "accent-soft": "var(--terracotta-soft)",
        paper: "var(--paper)",
        "paper-raised": "var(--paper-raised)",
        "paper-sunken": "var(--paper-sunken)",
        ink: "var(--ink)",
        "ink-soft": "var(--ink-soft)",
        line: "var(--line)",
        pencil: "var(--pencil)",
        "pencil-light": "var(--pencil-light)",
        dock: "var(--dock)",
        "dock-line": "var(--dock-line)",
      },
      animation: { ...animations },
      keyframes: { ...keyframes },
    },
  },
};

export default config;
