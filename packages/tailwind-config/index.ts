import type { Config } from "tailwindcss";
import { palette, fontFamilies, keyframes, animations } from "@movoz/tokens";

const config: Partial<Config> = {
  darkMode: "class",
  theme: {
    extend: {
      fontFamily: {
        sans: [...fontFamilies.sans],
        serif: [...fontFamilies.serif],
        ui: [...fontFamilies.ui],
      },
      colors: {
        accent: { ...palette.accent },
        paper: "var(--zen-paper)",
        zen: {
          subtle: "var(--zen-subtle)",
          bg: "var(--zen-bg)",
          text: "var(--zen-text)",
          muted: "var(--zen-muted)",
          border: "var(--zen-border)",
        },
      },
      animation: { ...animations },
      keyframes: { ...keyframes },
    },
  },
};

export default config;
