"use client";

import { forwardRef, type ComponentPropsWithoutRef } from "react";
import { cn } from "../../utils/cn";

export interface PillProps extends ComponentPropsWithoutRef<"button"> {
  active?: boolean;
}

export const Pill = forwardRef<HTMLButtonElement, PillProps>(
  ({ active = false, className, children, ...props }, ref) => {
    return (
      <button
        ref={ref}
        type="button"
        aria-pressed={active}
        className={cn(
          "inline-flex items-center gap-1.5 font-marker font-medium text-base leading-none",
          "px-[18px] py-2 rounded-full border transition-colors duration-150",
          active
            ? "bg-paper-sunken border-line text-ink"
            : "bg-paper-raised border-line text-ink hover:bg-paper-sunken",
          className
        )}
        {...props}
      >
        {children}
      </button>
    );
  }
);

Pill.displayName = "Pill";
