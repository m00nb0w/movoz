"use client";

import { forwardRef, type ComponentPropsWithoutRef } from "react";
import { cn } from "../../utils/cn";

export interface BadgeProps extends ComponentPropsWithoutRef<"span"> {
  variant?: "solid" | "subtle" | "outline";
  color?: "default" | "accent" | "success" | "warning" | "danger";
  size?: "sm" | "md";
}

const baseStyles =
  "inline-flex items-center font-body font-semibold uppercase tracking-wide rounded-[var(--radius-sm)]";

const sizeStyles = {
  sm: "px-2 py-0.5 text-xs",
  md: "px-2.5 py-1 text-sm",
};

const variantColorStyles = {
  solid: {
    default: "bg-ink text-paper-raised",
    accent: "bg-accent text-white",
    success: "bg-green-600 text-white",
    warning: "bg-amber-500 text-white",
    danger: "bg-red-500 text-white",
  },
  subtle: {
    default: "bg-paper-sunken text-ink",
    accent: "bg-accent-soft text-accent",
    success: "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400",
    warning: "bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400",
    danger: "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400",
  },
  outline: {
    default: "border border-line text-ink",
    accent: "border border-accent text-accent",
    success: "border border-green-500 text-green-600 dark:text-green-400",
    warning: "border border-amber-500 text-amber-600 dark:text-amber-400",
    danger: "border border-red-500 text-red-600 dark:text-red-400",
  },
};

export const Badge = forwardRef<HTMLSpanElement, BadgeProps>(
  (
    {
      variant = "subtle",
      color = "default",
      size = "sm",
      className,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <span
        ref={ref}
        className={cn(
          baseStyles,
          sizeStyles[size],
          variantColorStyles[variant][color],
          className
        )}
        {...props}
      >
        {children}
      </span>
    );
  }
);

Badge.displayName = "Badge";
