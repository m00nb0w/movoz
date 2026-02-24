"use client";

import { useState } from "react";
import { ThemeToggle } from "./ThemeToggle";
import { Menu, X, ChevronDown } from "lucide-react";

const navLinks = [
  { href: "#projects", label: "Projects" },
  { href: "#about", label: "About" },
  { href: "#contact", label: "Contact" },
];

export function Navigation() {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <header className="relative z-50">
      <nav className="max-w-6xl mx-auto px-4 md:px-6 py-6 flex items-center justify-between">
        {/* Logo */}
        <a
          href="#"
          className="text-zen-text hover:opacity-80 transition-opacity"
        >
          <span className="text-xl font-semibold">To Ngoc Long</span>
        </a>

        {/* Desktop Navigation - Centered */}
        <div className="hidden md:flex items-center gap-2 absolute left-1/2 -translate-x-1/2">
          {navLinks.map((link) => (
            <a
              key={link.href}
              href={link.href}
              className="px-5 py-2 text-lg font-medium text-zen-text hover:bg-zen-subtle rounded-lg transition-colors duration-200"
            >
              {link.label}
            </a>
          ))}
        </div>

        {/* Right side - Theme Toggle */}
        <div className="hidden md:flex items-center">
          <ThemeToggle />
        </div>

        {/* Mobile Menu Button */}
        <div className="flex md:hidden items-center gap-1">
          <ThemeToggle />
          <button
            onClick={() => setIsOpen(!isOpen)}
            className="p-2 rounded-lg hover:bg-zen-subtle transition-colors"
            aria-label="Toggle menu"
          >
            {isOpen ? (
              <X className="w-5 h-5 text-zen-text" />
            ) : (
              <Menu className="w-5 h-5 text-zen-text" />
            )}
          </button>
        </div>
      </nav>

      {/* Mobile Menu */}
      {isOpen && (
        <div className="md:hidden bg-zen-bg border-t border-zen-border">
          <div className="px-4 py-3 space-y-1">
            {navLinks.map((link) => (
              <a
                key={link.href}
                href={link.href}
                onClick={() => setIsOpen(false)}
                className="block px-4 py-2.5 text-sm text-zen-text hover:bg-zen-subtle rounded-lg transition-colors duration-200"
              >
                {link.label}
              </a>
            ))}
          </div>
        </div>
      )}
    </header>
  );
}
