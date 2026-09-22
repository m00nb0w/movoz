"use client";

import { useEffect } from "react";
import Link from "next/link";
import { Button, Card, Container, Text } from "@movoz/ui-web";

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
    <Container maxWidth="sm" className="py-12">
      <Text as="h1" font="marker" size="2xl" weight="bold" className="mb-2">
        Something went wrong
      </Text>
      <Text size="sm" color="muted" className="mb-6">
        This page failed to load. If you have been idle for a while your session may have expired — try again, or
        sign in again.
      </Text>
      {error.message && (
        <Card variant="outlined" padding="sm" className="mb-6">
          <Text size="sm" color="muted" className="break-words">
            {error.message}
          </Text>
        </Card>
      )}
      <div className="flex items-center gap-4">
        <Button onClick={() => reset()}>Try again</Button>
        <Link href="/login" className="text-sm text-accent hover:underline">
          Sign in
        </Link>
      </div>
    </Container>
  );
}
