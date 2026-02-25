"use client";

import { useEffect, useState } from "react";
import { useTheme } from "./ThemeProvider";
import { Sun, Moon, Monitor } from "lucide-react";

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const cycleTheme = () => {
    if (theme === "light") setTheme("dark");
    else if (theme === "dark") setTheme("system");
    else setTheme("light");
  };

  if (!mounted) {
    return (
      <button
        className="relative p-2 rounded-lg hover:bg-zen-subtle theme-transition"
        aria-label="Toggle theme"
      >
        <div className="relative w-5 h-5">
          <Sun className="w-5 h-5 text-zen-text" />
        </div>
      </button>
    );
  }

  return (
    <button
      onClick={cycleTheme}
      className="relative p-2 rounded-lg hover:bg-zen-subtle theme-transition"
      aria-label="Toggle theme"
    >
      <div className="relative w-5 h-5">
        {theme === "light" && (
          <Sun className="w-5 h-5 text-zen-text" />
        )}
        {theme === "dark" && (
          <Moon className="w-5 h-5 text-zen-text" />
        )}
        {theme === "system" && (
          <Monitor className="w-5 h-5 text-zen-text" />
        )}
      </div>
    </button>
  );
}
