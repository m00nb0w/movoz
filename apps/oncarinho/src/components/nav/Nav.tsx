"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { ThemeToggle } from "@movoz/theme";
import { LanguageToggle } from "./LanguageToggle";

export function Nav() {
  const t = useTranslations("nav");
  const tc = useTranslations("common");
  const pathname = usePathname();

  const links = [
    { href: "/", label: t("dashboard") },
    { href: "/leaderboard", label: t("leaderboard") },
  ];

  return (
    <nav className="border-b border-zen-border">
      <div className="mx-auto flex max-w-5xl flex-wrap items-center justify-between gap-4 px-4 py-4">
        <Link href="/" className="font-serif text-lg font-bold text-zen-text">
          {tc("appName")}
        </Link>
        <div className="flex flex-wrap items-center gap-4">
          {links.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              aria-current={pathname === link.href ? "page" : undefined}
              className={
                pathname === link.href
                  ? "text-sm font-medium text-zen-text"
                  : "text-sm text-zen-muted hover:text-zen-text"
              }
            >
              {link.label}
            </Link>
          ))}
          <LanguageToggle />
          <ThemeToggle />
          <Link href="/admin" className="text-xs text-zen-muted hover:text-zen-text">
            {t("admin")}
          </Link>
        </div>
      </div>
    </nav>
  );
}
