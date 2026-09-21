import type { Metadata } from "next";
import { ThemeProvider } from "@movoz/theme";
import Link from "next/link";
import "./globals.css";

export const metadata: Metadata = {
  title: "el-storko",
  description: "Personal work-item tracker",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="antialiased">
        <ThemeProvider>
          <nav className="flex gap-4 border-b border-line px-4 py-3">
            <Link href="/" className="text-sm font-semibold text-ink">
              Board
            </Link>
            <Link href="/list" className="text-sm font-semibold text-ink">
              List
            </Link>
          </nav>
          {children}
        </ThemeProvider>
      </body>
    </html>
  );
}
