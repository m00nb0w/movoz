"use client";

import { ArrowRight, Heart } from "lucide-react";

export function Hero() {
  return (
    <section className="min-h-[80vh] flex flex-col justify-center items-center px-6">
      <div className="max-w-5xl mx-auto text-center">
        <h1 className="font-serif text-[2.5rem] md:text-[4rem] lg:text-[5.5rem] font-normal mb-8 leading-[1.1] tracking-[-0.01em]">
          <span className="text-zen-text">welcome to</span>
          <br />
          <span className="text-zen-text">a </span>
          <span className="italic text-accent">calmer</span>
          <span className="text-zen-text"> internet</span>
        </h1>

        <p className="text-zen-text text-base md:text-[1.1rem] leading-[1.6] font-normal max-w-2xl mx-auto mb-10">
          Beautifully designed, privacy-focused, and packed with features.
          <br />
          We care about your experience, not your data.
        </p>

        <div className="flex flex-wrap justify-center gap-4">
          <a
            href="#projects"
            className="inline-flex items-center gap-2 px-6 py-4 bg-zen-text text-zen-bg text-base font-medium rounded-xl hover:opacity-90 transition-opacity duration-200"
          >
            Beta is now available!
            <ArrowRight className="w-5 h-5" />
          </a>
          <a
            href="#contact"
            className="inline-flex items-center gap-2 px-6 py-4 bg-zen-subtle text-zen-text text-base font-medium rounded-xl hover:bg-zen-border transition-colors duration-200"
          >
            Support Us
            <Heart className="w-5 h-5 text-red-500" fill="currentColor" />
          </a>
        </div>
      </div>
    </section>
  );
}
