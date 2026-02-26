"use client";

import { ArrowRight, Heart } from "lucide-react";
import { Text, Container, cn } from "@movoz/ui-web";

export function Hero() {
  return (
    <section className="min-h-[80vh] flex flex-col justify-center items-center px-6">
      <Container maxWidth="xl" className="text-center">
        <Text
          as="h1"
          font="serif"
          className="text-[2.5rem] md:text-[4rem] lg:text-[5.5rem] mb-8 leading-[1.1] tracking-[-0.01em]"
        >
          <span className="text-zen-text">welcome to</span>
          <br />
          <span className="text-zen-text">a </span>
          <span className="italic text-accent">calmer</span>
          <span className="text-zen-text"> internet</span>
        </Text>

        <Text className="md:text-[1.1rem] leading-[1.6] max-w-2xl mx-auto mb-10">
          Beautifully designed, privacy-focused, and packed with features.
          <br />
          We care about your experience, not your data.
        </Text>

        <div className="flex flex-wrap justify-center gap-4">
          <a
            href="#projects"
            className={cn(
              "inline-flex items-center justify-center font-medium transition-all duration-200",
              "px-6 py-3.5 text-base rounded-xl gap-2",
              "bg-zen-text text-zen-bg hover:opacity-90"
            )}
          >
            Beta is now available!
            <ArrowRight className="w-5 h-5" />
          </a>
          <a
            href="#contact"
            className={cn(
              "inline-flex items-center justify-center font-medium transition-all duration-200",
              "px-6 py-3.5 text-base rounded-xl gap-2",
              "bg-zen-subtle text-zen-text border border-zen-border hover:bg-zen-border"
            )}
          >
            Support Us
            <Heart className="w-5 h-5 text-red-500" fill="currentColor" />
          </a>
        </div>
      </Container>
    </section>
  );
}
