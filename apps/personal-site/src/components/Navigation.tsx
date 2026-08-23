"use client";

import { useState } from "react";
import { ThemeToggle } from "@movoz/theme";
import { Menu, X } from "lucide-react";
import { IconButton, Container, Stack } from "@movoz/ui-web";

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
          className="text-ink hover:opacity-80 transition-opacity"
        >
          <span className="text-xl font-semibold">To Ngoc Long</span>
        </a>

        {/* Desktop Navigation - Centered */}
        <Stack direction="horizontal" gap={2} className="hidden md:flex absolute left-1/2 -translate-x-1/2">
          {navLinks.map((link) => (
            <a
              key={link.href}
              href={link.href}
              className="px-5 py-2 text-lg font-medium text-ink hover:bg-paper-sunken rounded-lg transition-colors duration-200"
            >
              {link.label}
            </a>
          ))}
        </Stack>

        {/* Right side - Theme Toggle */}
        <div className="hidden md:flex items-center">
          <ThemeToggle />
        </div>

        {/* Mobile Menu Button */}
        <Stack direction="horizontal" gap={1} className="flex md:hidden" align="center">
          <ThemeToggle />
          <IconButton
            icon={
              isOpen ? (
                <X className="w-5 h-5 text-ink" />
              ) : (
                <Menu className="w-5 h-5 text-ink" />
              )
            }
            label="Toggle menu"
            onClick={() => setIsOpen(!isOpen)}
          />
        </Stack>
      </nav>

      {/* Mobile Menu */}
      {isOpen && (
        <div className="md:hidden bg-paper border-t border-line">
          <Stack gap={1} className="px-4 py-3">
            {navLinks.map((link) => (
              <a
                key={link.href}
                href={link.href}
                onClick={() => setIsOpen(false)}
                className="block px-4 py-2.5 text-sm text-ink hover:bg-paper-sunken rounded-lg transition-colors duration-200"
              >
                {link.label}
              </a>
            ))}
          </Stack>
        </div>
      )}
    </header>
  );
}
