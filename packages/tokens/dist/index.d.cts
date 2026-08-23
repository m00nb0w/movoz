declare const palette: {
    readonly accent: {
        readonly DEFAULT: "#C25A36";
        readonly light: "#E07A50";
        readonly dark: "#A1462A";
    };
};
declare const semantic: {
    readonly light: {
        readonly paper: "#EFE9DC";
        readonly paperRaised: "#FAF5EA";
        readonly paperSunken: "#E6DECE";
        readonly ink: "#25231E";
        readonly inkSoft: "#6B675F";
        readonly line: "#D4D0C6";
    };
    readonly dark: {
        readonly paper: "#21201B";
        readonly paperRaised: "#2B2A24";
        readonly paperSunken: "#1A1916";
        readonly ink: "#F2EEE4";
        readonly inkSoft: "#AFA99B";
        readonly line: "#3B382F";
    };
};
type SemanticColorKey = keyof (typeof semantic)["light"];
type ColorMode = keyof typeof semantic;

declare const fontFamilies: {
    readonly marker: readonly ["Shantell Sans", "Comic Sans MS", "cursive"];
    readonly body: readonly ["Space Grotesk", "ui-sans-serif", "system-ui", "sans-serif"];
};
declare const fontSizes: {
    readonly xs: "0.75rem";
    readonly sm: "0.875rem";
    readonly base: "1rem";
    readonly lg: "1.125rem";
    readonly xl: "1.25rem";
    readonly "2xl": "1.5rem";
    readonly "3xl": "1.875rem";
    readonly "4xl": "2.25rem";
    readonly "5xl": "3rem";
    readonly "6xl": "3.75rem";
    readonly "7xl": "4.5rem";
};
declare const fontWeights: {
    readonly light: 300;
    readonly normal: 400;
    readonly medium: 500;
    readonly semibold: 600;
    readonly bold: 700;
    readonly extrabold: 800;
};
declare const lineHeights: {
    readonly none: 1;
    readonly tight: 1.1;
    readonly snug: 1.3;
    readonly normal: 1.5;
    readonly relaxed: 1.6;
    readonly loose: 1.7;
};
declare const letterSpacings: {
    readonly tighter: "-0.02em";
    readonly tight: "-0.01em";
    readonly normal: "0";
    readonly wide: "0.01em";
};

declare const spacing: {
    readonly 0: "0";
    readonly 0.5: "0.125rem";
    readonly 1: "0.25rem";
    readonly 1.5: "0.375rem";
    readonly 2: "0.5rem";
    readonly 2.5: "0.625rem";
    readonly 3: "0.75rem";
    readonly 4: "1rem";
    readonly 5: "1.25rem";
    readonly 6: "1.5rem";
    readonly 8: "2rem";
    readonly 10: "2.5rem";
    readonly 12: "3rem";
    readonly 16: "4rem";
    readonly 20: "5rem";
    readonly 24: "6rem";
    readonly 32: "8rem";
};

declare const radii: {
    readonly none: "0";
    readonly sm: "7px";
    readonly DEFAULT: "11px";
    readonly md: "11px";
    readonly lg: "16px";
    readonly xl: "22px";
    readonly "2xl": "30px";
    readonly full: "9999px";
    readonly sketch: "14px 11px 13px 12px";
};

declare const shadows: {
    readonly none: "none";
    readonly sm: "0 1px 2px rgba(37, 35, 30, 0.05)";
    readonly DEFAULT: "0 1px 3px rgba(37, 35, 30, 0.05), 0 4px 12px rgba(37, 35, 30, 0.04)";
    readonly md: "0 4px 6px -1px rgba(37, 35, 30, 0.07), 0 2px 4px -2px rgba(37, 35, 30, 0.05)";
    readonly lg: "0 10px 15px -3px rgba(37, 35, 30, 0.08), 0 4px 6px -4px rgba(37, 35, 30, 0.04)";
    readonly xl: "0 20px 25px -5px rgba(37, 35, 30, 0.1), 0 8px 10px -6px rgba(37, 35, 30, 0.04)";
    readonly "2xl": "0 20px 40px -12px rgba(37, 35, 30, 0.15)";
    readonly sketch: "3px 3px 0 var(--ink)";
    readonly sketchAccent: "3px 3px 0 var(--terracotta)";
};
declare const darkShadows: {
    readonly DEFAULT: "0 1px 3px rgba(0, 0, 0, 0.2), 0 4px 12px rgba(0, 0, 0, 0.15)";
    readonly "2xl": "0 20px 40px -12px rgba(0, 0, 0, 0.4)";
};

declare const breakpoints: {
    readonly sm: "640px";
    readonly md: "768px";
    readonly lg: "1024px";
    readonly xl: "1280px";
    readonly "2xl": "1536px";
};

declare const durations: {
    readonly fast: "150ms";
    readonly normal: "200ms";
    readonly slow: "300ms";
    readonly slower: "600ms";
};
declare const easings: {
    readonly ease: "ease";
    readonly easeIn: "ease-in";
    readonly easeOut: "ease-out";
    readonly easeInOut: "ease-in-out";
    readonly cubic: "cubic-bezier(0.4, 0, 0.2, 1)";
};
declare const keyframes: {
    readonly fadeIn: {
        readonly "0%": {
            readonly opacity: "0";
        };
        readonly "100%": {
            readonly opacity: "1";
        };
    };
    readonly slideUp: {
        readonly "0%": {
            readonly opacity: "0";
            readonly transform: "translateY(20px)";
        };
        readonly "100%": {
            readonly opacity: "1";
            readonly transform: "translateY(0)";
        };
    };
    readonly slideInRight: {
        readonly "0%": {
            readonly opacity: "0";
            readonly transform: "translateX(-20px)";
        };
        readonly "100%": {
            readonly opacity: "1";
            readonly transform: "translateX(0)";
        };
    };
};
declare const animations: {
    readonly "fade-in": "fadeIn 0.6s ease-out forwards";
    readonly "slide-up": "slideUp 0.6s ease-out forwards";
    readonly "slide-in-right": "slideInRight 0.6s ease-out forwards";
};

export { type ColorMode, type SemanticColorKey, animations, breakpoints, darkShadows, durations, easings, fontFamilies, fontSizes, fontWeights, keyframes, letterSpacings, lineHeights, palette, radii, semantic, shadows, spacing };
