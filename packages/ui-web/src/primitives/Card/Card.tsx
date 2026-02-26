"use client";

import { forwardRef, type ComponentPropsWithoutRef, type ReactNode } from "react";
import { cn } from "../../utils/cn";

export interface CardProps extends ComponentPropsWithoutRef<"div"> {
  variant?: "elevated" | "outlined" | "filled";
  padding?: "none" | "sm" | "md" | "lg";
  header?: ReactNode;
  footer?: ReactNode;
}

const variantStyles = {
  elevated: "bg-paper shadow-[0_1px_3px_rgba(0,0,0,0.05),0_4px_12px_rgba(0,0,0,0.04)] dark:shadow-[0_1px_3px_rgba(0,0,0,0.2),0_4px_12px_rgba(0,0,0,0.15)]",
  outlined: "bg-paper border border-zen-border",
  filled: "bg-zen-subtle",
};

const paddingStyles = {
  none: "",
  sm: "p-4",
  md: "p-6",
  lg: "p-8",
};

export const Card = forwardRef<HTMLDivElement, CardProps>(
  (
    {
      variant = "elevated",
      padding = "md",
      header,
      footer,
      className,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <div
        ref={ref}
        className={cn(
          "rounded-2xl overflow-hidden",
          variantStyles[variant],
          !header && !footer && paddingStyles[padding],
          className
        )}
        {...props}
      >
        {header && (
          <div className={cn("border-b border-zen-border", paddingStyles[padding])}>
            {header}
          </div>
        )}
        {header || footer ? (
          <div className={paddingStyles[padding]}>{children}</div>
        ) : (
          children
        )}
        {footer && (
          <div className={cn("border-t border-zen-border", paddingStyles[padding])}>
            {footer}
          </div>
        )}
      </div>
    );
  }
);

Card.displayName = "Card";
