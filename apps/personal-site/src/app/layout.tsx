import type { Metadata } from "next";
import "./globals.css";
import { ThemeProvider } from "@movoz/theme";

export const metadata: Metadata = {
  title: "To Ngoc Long - Software Engineer",
  description:
    "A passionate developer crafting elegant solutions with modern technologies. Focused on creating impactful, user-centered products.",
  keywords: [
    "software engineer",
    "web developer",
    "full-stack",
    "React",
    "Next.js",
    "TypeScript",
  ],
  authors: [{ name: "To Ngoc Long" }],
  openGraph: {
    title: "To Ngoc Long - Software Engineer",
    description:
      "A passionate developer crafting elegant solutions with modern technologies.",
    type: "website",
  },
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
