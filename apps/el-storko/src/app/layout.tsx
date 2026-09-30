import type { Metadata } from "next";
import Link from "next/link";
import { Text } from "@movoz/ui-web";
import { ThemeProvider } from "@movoz/theme";
import { NavTabs } from "@/components/NavTabs";
import { ScopeToggle } from "@/components/ScopeToggle";
import { ScopeProvider } from "@/lib/scope";
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
          <ScopeProvider>
            <nav className="flex items-center justify-between gap-6 border-b border-line px-4 py-3">
              <div className="flex items-center gap-6">
                <Link href="/">
                  <Text as="span" font="marker" weight="semibold" size="lg" color="accent">
                    el-storko
                  </Text>
                </Link>
                <NavTabs />
              </div>
              <ScopeToggle />
            </nav>
            {children}
          </ScopeProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
