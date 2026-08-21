"use client";

import { useEffect } from "react";
import Link from "next/link";

/**
 * Shared App Router error boundary for every route segment under /scout.
 *
 * Several pages fetch on mount (roster dashboard, cycle view, engineer card,
 * ranking pages) and previously had no error handling at all: a 401/404/network
 * failure left the page stuck on a blank/loading state with nothing on screen.
 * Next.js automatically wraps each route segment in the nearest `error.tsx`, so
 * this one file covers render-time throws and async errors that propagate up
 * during rendering, without touching each page's fetch logic.
 *
 * Known limitation: a rejected promise inside a `useEffect` that never calls
 * `setState` (i.e. never re-renders and never re-throws during render) is *not*
 * caught by this boundary — React has no way to route it here. Those call sites
 * still need their own try/catch (or a suspense-based data layer) to surface the
 * failure; this boundary is the standard safety net, not a substitute for that.
 */
export default function ScoutError({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
  useEffect(() => {
    // Surfacing the real error in the console matters here: the message shown
    // to the admin is deliberately generic, and Next.js redacts server-side
    // error messages in production builds.
    console.error("scout: unhandled error", error);
  }, [error]);

  return (
    <main className="mx-auto max-w-2xl p-8">
      <h1 className="mb-2 text-2xl font-semibold text-zen-text">Something went wrong</h1>
      <p className="mb-6 text-sm text-zen-muted">
        This page failed to load. If you have been idle for a while your session may have expired — try again, or sign
        in again.
      </p>
      {error.message && (
        <p className="mb-6 break-words rounded border border-zen-border p-3 text-sm text-zen-muted">{error.message}</p>
      )}
      <div className="flex gap-3">
        <button onClick={() => reset()} className="rounded bg-accent-600 px-4 py-2 text-white">
          Try again
        </button>
        <Link href="/login" className="rounded border border-zen-border px-4 py-2 text-sm text-zen-text">
          Sign in
        </Link>
      </div>
    </main>
  );
}
