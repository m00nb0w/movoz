"use client";

import { Heart } from "lucide-react";
import { Text, Container, Divider } from "@movoz/ui-web";

export function Footer() {
  const currentYear = new Date().getFullYear();

  return (
    <footer className="py-12 px-6">
      <Container>
        <Divider className="mb-12" />
        <div className="flex flex-col md:flex-row items-center justify-between gap-4">
          <Text as="p" size="sm" className="flex items-center gap-1">
            Built with <Heart className="w-4 h-4 text-accent" fill="currentColor" /> using Next.js
          </Text>
          <Text as="p" size="sm">
            &copy; {currentYear} To Ngoc Long
          </Text>
        </div>
      </Container>
    </footer>
  );
}
