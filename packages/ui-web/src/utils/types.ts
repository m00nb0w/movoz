import type { ElementType, ComponentPropsWithoutRef, ReactNode } from "react";

export type PolymorphicProps<E extends ElementType, P = {}> = P &
  Omit<ComponentPropsWithoutRef<E>, keyof P> & {
    as?: E;
  };

export type Size = "sm" | "md" | "lg";

export type ButtonVariant = "primary" | "secondary" | "ghost" | "danger";
export type BadgeVariant = "solid" | "subtle" | "outline";
export type CardVariant = "elevated" | "outlined" | "filled";
export type ToastVariant = "success" | "error" | "warning" | "info";
