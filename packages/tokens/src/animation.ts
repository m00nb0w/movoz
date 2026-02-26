export const durations = {
  fast: "150ms",
  normal: "200ms",
  slow: "300ms",
  slower: "600ms",
} as const;

export const easings = {
  ease: "ease",
  easeIn: "ease-in",
  easeOut: "ease-out",
  easeInOut: "ease-in-out",
  cubic: "cubic-bezier(0.4, 0, 0.2, 1)",
} as const;

export const keyframes = {
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
} as const;

export const animations = {
  "fade-in": "fadeIn 0.6s ease-out forwards",
  "slide-up": "slideUp 0.6s ease-out forwards",
  "slide-in-right": "slideInRight 0.6s ease-out forwards",
} as const;
