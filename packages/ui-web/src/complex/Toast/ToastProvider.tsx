"use client";

import {
  createContext,
  useCallback,
  useState,
  type ReactNode,
} from "react";
import { createPortal } from "react-dom";
import { ToastItem, type ToastData } from "./Toast";
import { cn } from "../../utils/cn";

type Position =
  | "top-right"
  | "top-left"
  | "bottom-right"
  | "bottom-left"
  | "top-center"
  | "bottom-center";

export interface ToastContextValue {
  toast: (options: Omit<ToastData, "id">) => void;
  dismiss: (id: string) => void;
}

export const ToastContext = createContext<ToastContextValue | null>(null);

const positionStyles: Record<Position, string> = {
  "top-right": "top-4 right-4",
  "top-left": "top-4 left-4",
  "bottom-right": "bottom-4 right-4",
  "bottom-left": "bottom-4 left-4",
  "top-center": "top-4 left-1/2 -translate-x-1/2",
  "bottom-center": "bottom-4 left-1/2 -translate-x-1/2",
};

let toastCounter = 0;

export function ToastProvider({
  children,
  position = "bottom-right",
}: {
  children: ReactNode;
  position?: Position;
}) {
  const [toasts, setToasts] = useState<ToastData[]>([]);

  const addToast = useCallback((options: Omit<ToastData, "id">) => {
    const id = `toast-${++toastCounter}`;
    setToasts((prev) => [...prev, { id, ...options }]);
  }, []);

  const dismiss = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  return (
    <ToastContext.Provider value={{ toast: addToast, dismiss }}>
      {children}
      {typeof document !== "undefined" &&
        createPortal(
          <div
            className={cn(
              "fixed z-50 flex flex-col gap-2 pointer-events-none",
              positionStyles[position]
            )}
          >
            {toasts.map((t) => (
              <div key={t.id} className="pointer-events-auto">
                <ToastItem toast={t} onDismiss={dismiss} />
              </div>
            ))}
          </div>,
          document.body
        )}
    </ToastContext.Provider>
  );
}
