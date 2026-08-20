import type { Metadata } from "next";
import "./globals.css";
import { ThemeProvider } from "@movoz/theme";

export const metadata: Metadata = {
  title: "Scout",
  description: "Engineer performance tracking — FIFA-style attribute cards, biweekly rankings, and AI-assisted reviews.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="antialiased">
        <ThemeProvider>{children}</ThemeProvider>
      </body>
    </html>
  );
}
