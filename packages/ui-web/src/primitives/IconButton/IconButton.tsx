"use client";

import { forwardRef, type ComponentPropsWithoutRef, type ReactNode } from "react";
import { cn } from "../../utils/cn";

export interface IconButtonProps
  extends Omit<ComponentPropsWithoutRef<"button">, "children"> {
  icon: ReactNode;
  variant?: "primary" | "secondary" | "ghost" | "danger";
  size?: "sm" | "md" | "lg";
  label: string;
}

const variantStyles = {
  primary: "bg-ink text-paper-raised hover:opacity-90",
  secondary: "bg-paper-sunken text-ink hover:bg-line",
  ghost: "bg-transparent text-ink hover:bg-paper-sunken",
  danger: "bg-red-500 text-white hover:bg-red-600",
};

const sizeStyles = {
  sm: "p-1.5 rounded-[var(--radius-sm)]",
  md: "p-2.5 rounded-[var(--radius-sm)]",
  lg: "p-3 rounded-[var(--radius-md)]",
};

export const IconButton = forwardRef<HTMLButtonElement, IconButtonProps>(
  (
    {
      icon,
      variant = "ghost",
      size = "md",
      label,
      disabled,
      className,
      ...props
    },
    ref
  ) => {
    return (
      <button
        ref={ref}
        aria-label={label}
        disabled={disabled}
        className={cn(
          "inline-flex items-center justify-center transition-all duration-200",
          "disabled:opacity-45 disabled:pointer-events-none",
          variantStyles[variant],
          sizeStyles[size],
          className
        )}
        {...props}
      >
        {icon}
      </button>
    );
  }
);

IconButton.displayName = "IconButton";
