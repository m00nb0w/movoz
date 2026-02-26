"use client";

import { forwardRef, useState, type ComponentPropsWithoutRef } from "react";
import { cn } from "../../utils/cn";

export interface AvatarProps extends ComponentPropsWithoutRef<"div"> {
  src?: string;
  alt?: string;
  fallback?: string;
  size?: "sm" | "md" | "lg" | "xl";
  shape?: "circle" | "square";
}

const sizeStyles = {
  sm: "w-8 h-8 text-xs",
  md: "w-10 h-10 text-sm",
  lg: "w-12 h-12 text-base",
  xl: "w-16 h-16 text-lg",
};

const shapeStyles = {
  circle: "rounded-full",
  square: "rounded-xl",
};

export const Avatar = forwardRef<HTMLDivElement, AvatarProps>(
  (
    {
      src,
      alt = "",
      fallback,
      size = "md",
      shape = "circle",
      className,
      ...props
    },
    ref
  ) => {
    const [imgError, setImgError] = useState(false);
    const showFallback = !src || imgError;

    const initials = fallback
      ?.split(" ")
      .map((w) => w[0])
      .join("")
      .slice(0, 2)
      .toUpperCase();

    return (
      <div
        ref={ref}
        className={cn(
          "relative inline-flex items-center justify-center overflow-hidden",
          "bg-zen-subtle text-zen-muted font-medium",
          sizeStyles[size],
          shapeStyles[shape],
          className
        )}
        {...props}
      >
        {showFallback ? (
          <span>{initials || "?"}</span>
        ) : (
          <img
            src={src}
            alt={alt}
            onError={() => setImgError(true)}
            className="w-full h-full object-cover"
          />
        )}
      </div>
    );
  }
);

Avatar.displayName = "Avatar";
