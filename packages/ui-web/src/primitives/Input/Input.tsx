"use client";

import { forwardRef, type ComponentPropsWithoutRef, type ReactNode } from "react";
import { cn } from "../../utils/cn";

export interface InputProps
  extends Omit<ComponentPropsWithoutRef<"input">, "size"> {
  label?: string;
  error?: string;
  helperText?: string;
  icon?: ReactNode;
  size?: "sm" | "md" | "lg";
}

const sizeStyles = {
  sm: "px-3 py-2 text-sm",
  md: "px-4 py-3 text-base",
  lg: "px-5 py-4 text-base",
};

export const Input = forwardRef<HTMLInputElement, InputProps>(
  (
    {
      label,
      error,
      helperText,
      icon,
      size = "md",
      className,
      id,
      ...props
    },
    ref
  ) => {
    const inputId = id || label?.toLowerCase().replace(/\s+/g, "-");

    return (
      <div className="w-full">
        {label && (
          <label
            htmlFor={inputId}
            className="block font-semibold text-zen-text text-base mb-2"
          >
            {label}
          </label>
        )}
        <div className="relative">
          {icon && (
            <div className="absolute left-3 top-1/2 -translate-y-1/2 text-zen-muted">
              {icon}
            </div>
          )}
          <input
            ref={ref}
            id={inputId}
            className={cn(
              "w-full bg-zen-subtle border rounded-lg transition-all",
              "focus:outline-none focus:ring-2 focus:ring-zen-text/20 focus:border-zen-text",
              "text-zen-text placeholder-zen-muted",
              error
                ? "border-red-500 focus:ring-red-500/20 focus:border-red-500"
                : "border-zen-border",
              sizeStyles[size],
              icon && "pl-10",
              className
            )}
            {...props}
          />
        </div>
        {error && (
          <p className="mt-1.5 text-sm text-red-500">{error}</p>
        )}
        {helperText && !error && (
          <p className="mt-1.5 text-sm text-zen-muted">{helperText}</p>
        )}
      </div>
    );
  }
);

Input.displayName = "Input";
