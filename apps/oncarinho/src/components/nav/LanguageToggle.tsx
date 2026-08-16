"use client";

import { useLocale } from "next-intl";
import { useRouter } from "next/navigation";
import { locales, localeCookieName, type Locale } from "@/i18n/config";

export function LanguageToggle() {
  const locale = useLocale() as Locale;
  const router = useRouter();

  function switchTo(next: Locale) {
    document.cookie = `${localeCookieName}=${next};path=/;max-age=${60 * 60 * 24 * 365}`;
    router.refresh();
  }

  return (
    <div className="flex items-center gap-1 text-sm" role="group" aria-label="Language">
      {locales.map((option) => (
        <button
          key={option}
          onClick={() => switchTo(option)}
          aria-pressed={option === locale}
          className={
            option === locale
              ? "px-2 py-1 font-medium text-zen-text"
              : "px-2 py-1 text-zen-muted hover:text-zen-text"
          }
        >
          {option.toUpperCase()}
        </button>
      ))}
    </div>
  );
}
