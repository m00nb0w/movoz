"use client";

import { type ElementType, type ReactNode, createElement } from "react";
import { cn } from "../../utils/cn";

export interface TextProps<E extends ElementType = "p"> {
  as?: E;
  size?: "xs" | "sm" | "base" | "lg" | "xl" | "2xl" | "3xl" | "4xl" | "5xl";
  weight?: "light" | "normal" | "medium" | "semibold" | "bold";
  color?: "default" | "muted" | "accent";
  font?: "sans" | "serif" | "ui";
  truncate?: boolean;
  align?: "left" | "center" | "right";
  className?: string;
  children: ReactNode;
}

const sizeStyles = {
  xs: "text-xs",
  sm: "text-sm",
  base: "text-base",
  lg: "text-lg",
  xl: "text-xl",
  "2xl": "text-2xl",
  "3xl": "text-3xl",
  "4xl": "text-4xl",
  "5xl": "text-5xl",
};

const weightStyles = {
  light: "font-light",
  normal: "font-normal",
  medium: "font-medium",
  semibold: "font-semibold",
  bold: "font-bold",
};

const colorStyles = {
  default: "text-zen-text",
  muted: "text-zen-muted",
  accent: "text-accent",
};

const fontStyles = {
  sans: "font-sans",
  serif: "font-serif",
  ui: "font-ui",
};

const alignStyles = {
  left: "text-left",
  center: "text-center",
  right: "text-right",
};

export function Text<E extends ElementType = "p">({
  as,
  size = "base",
  weight = "normal",
  color = "default",
  font,
  truncate = false,
  align,
  className,
  children,
  ...props
}: TextProps<E> & Omit<React.ComponentPropsWithoutRef<E>, keyof TextProps<E>>) {
  const Component = as || "p";
  return createElement(
    Component,
    {
      className: cn(
        sizeStyles[size],
        weightStyles[weight],
        colorStyles[color],
        font && fontStyles[font],
        align && alignStyles[align],
        truncate && "truncate",
        className
      ),
      ...props,
    },
    children
  );
}
