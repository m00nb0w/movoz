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
        className="relative p-2 rounded-lg hover:bg-paper-sunken theme-transition"
        aria-label="Toggle theme"
      >
        <div className="relative w-5 h-5">
          <Sun className="w-5 h-5 text-ink" />
        </div>
      </button>
    );
  }

  return (
    <button
      onClick={cycleTheme}
      className="relative p-2 rounded-lg hover:bg-paper-sunken theme-transition"
      aria-label="Toggle theme"
    >
      <div className="relative w-5 h-5">
        {theme === "light" && (
          <Sun className="w-5 h-5 text-ink" />
        )}
        {theme === "dark" && (
          <Moon className="w-5 h-5 text-ink" />
        )}
        {theme === "system" && (
          <Monitor className="w-5 h-5 text-ink" />
        )}
      </div>
    </button>
  );
}
