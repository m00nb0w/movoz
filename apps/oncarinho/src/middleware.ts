import { NextRequest, NextResponse } from "next/server";
import { locales, defaultLocale, localeCookieName, type Locale } from "./i18n/config";

function negotiateLocale(acceptLanguage: string | null): Locale {
  if (!acceptLanguage) return defaultLocale;
  const preferred = acceptLanguage
    .split(",")
    .map((part) => part.split(";")[0].trim().toLowerCase());

  for (const lang of preferred) {
    const match = (locales as readonly string[]).find(
      (locale) => lang === locale || lang.startsWith(`${locale}-`)
    );
    if (match) return match as Locale;
  }
  return defaultLocale;
}

export function middleware(request: NextRequest) {
  const response = NextResponse.next();
  const existing = request.cookies.get(localeCookieName)?.value;

  if (!existing || !(locales as readonly string[]).includes(existing)) {
    const locale = negotiateLocale(request.headers.get("accept-language"));
    response.cookies.set(localeCookieName, locale, {
      path: "/",
      maxAge: 60 * 60 * 24 * 365,
    });
  }

  return response;
}

export const config = {
  matcher: ["/((?!_next|api|favicon.ico).*)"],
};
