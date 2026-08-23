"use client";

import { forwardRef, type ComponentPropsWithoutRef, type ReactNode } from "react";
import { cn } from "../../utils/cn";

export interface ButtonProps extends ComponentPropsWithoutRef<"button"> {
  variant?: "primary" | "secondary" | "ghost" | "danger";
  size?: "sm" | "md" | "lg";
  loading?: boolean;
  icon?: ReactNode;
  iconRight?: ReactNode;
}

const variantStyles = {
  primary:
    "bg-ink text-paper-raised border border-ink hover:opacity-90 shadow-[var(--shadow-sketch)] active:translate-x-[3px] active:translate-y-[3px] active:shadow-none",
  secondary:
    "bg-paper-raised text-ink border border-ink hover:bg-paper-sunken",
  ghost:
    "bg-transparent text-ink border border-transparent hover:bg-paper-sunken",
  danger:
    "bg-red-500 text-white border border-red-500 hover:bg-red-600",
};

const sizeStyles = {
  sm: "px-3 py-1.5 text-sm rounded-[var(--radius-sm)] gap-1.5",
  md: "px-5 py-2.5 text-base rounded-[var(--radius-md)] gap-2",
  lg: "px-6 py-3.5 text-base rounded-[var(--radius-md)] gap-2",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      variant = "primary",
      size = "md",
      loading = false,
      icon,
      iconRight,
      disabled,
      className,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <button
        ref={ref}
        disabled={disabled || loading}
        className={cn(
          "inline-flex items-center justify-center font-marker font-semibold transition-all duration-200",
          "disabled:opacity-45 disabled:pointer-events-none",
          variantStyles[variant],
          sizeStyles[size],
          className
        )}
        {...props}
      >
        {loading ? (
          <svg
            className="animate-spin h-4 w-4"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
            />
          </svg>
        ) : (
          icon
        )}
        {children}
        {iconRight}
      </button>
    );
  }
);

Button.displayName = "Button";
