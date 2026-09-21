import type { Metadata } from "next";
import { ThemeProvider } from "@movoz/theme";
import { NavTabs } from "@/components/NavTabs";
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
          <nav className="border-b border-line px-4 py-3">
            <NavTabs />
          </nav>
          {children}
        </ThemeProvider>
      </body>
    </html>
  );
}
