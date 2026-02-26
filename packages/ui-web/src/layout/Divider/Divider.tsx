"use client";

import { forwardRef, type ComponentPropsWithoutRef } from "react";
import { cn } from "../../utils/cn";

export interface DividerProps extends ComponentPropsWithoutRef<"hr"> {
  orientation?: "horizontal" | "vertical";
  color?: "default" | "subtle";
}

export const Divider = forwardRef<HTMLHRElement, DividerProps>(
  (
    {
      orientation = "horizontal",
      color = "default",
      className,
      ...props
    },
    ref
  ) => {
    const colorStyle = color === "subtle" ? "border-zen-subtle" : "border-zen-border";

    if (orientation === "vertical") {
      return (
        <hr
          ref={ref}
          className={cn(
            "border-l h-full border-t-0",
            colorStyle,
            className
          )}
          {...props}
        />
      );
    }

    return (
      <hr
        ref={ref}
        className={cn("border-t w-full", colorStyle, className)}
        {...props}
      />
    );
  }
);

Divider.displayName = "Divider";
