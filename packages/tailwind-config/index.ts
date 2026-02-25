import type { Config } from "tailwindcss";

const config: Partial<Config> = {
  darkMode: "class",
  theme: {
    extend: {
      fontFamily: {
        sans: ["Rubik", "system-ui", "sans-serif"],
        serif: ["Libre Baskerville", "Georgia", "serif"],
        ui: ["Inter", "system-ui", "sans-serif"],
      },
      colors: {
        accent: {
          DEFAULT: "#d4775c",
          light: "#e08a70",
          dark: "#c46448",
        },
        paper: "var(--zen-paper)",
        zen: {
          subtle: "var(--zen-subtle)",
          bg: "var(--zen-bg)",
          text: "var(--zen-text)",
          muted: "var(--zen-muted)",
          border: "var(--zen-border)",
        },
      },
      animation: {
        "fade-in": "fadeIn 0.6s ease-out forwards",
        "slide-up": "slideUp 0.6s ease-out forwards",
        "slide-in-right": "slideInRight 0.6s ease-out forwards",
      },
      keyframes: {
        fadeIn: {
          "0%": { opacity: "0" },
          "100%": { opacity: "1" },
        },
        slideUp: {
          "0%": { opacity: "0", transform: "translateY(20px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        slideInRight: {
          "0%": { opacity: "0", transform: "translateX(-20px)" },
          "100%": { opacity: "1", transform: "translateX(0)" },
        },
      },
    },
  },
};

export default config;
