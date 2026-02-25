"use client";

import { Heart } from "lucide-react";

export function Footer() {
  const currentYear = new Date().getFullYear();

  return (
    <footer className="py-12 px-6 border-t border-zen-border">
      <div className="max-w-6xl mx-auto flex flex-col md:flex-row items-center justify-between gap-4 text-sm text-zen-text">
        <p className="flex items-center gap-1">
          Built with <Heart className="w-4 h-4 text-accent" fill="currentColor" /> using Next.js
        </p>
        <p>
          &copy; {currentYear} To Ngoc Long
        </p>
      </div>
    </footer>
  );
}
