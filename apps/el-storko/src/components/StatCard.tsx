import { Card, Text } from "@movoz/ui-web";
import type { ReactNode } from "react";

export function StatCard({
  title,
  summary,
  children,
}: {
  title: string;
  summary?: string;
  children: ReactNode;
}) {
  return (
    <Card variant="outlined" padding="md">
      <div className="mb-2 flex items-baseline justify-between gap-2">
        <Text as="h3" font="marker" weight="semibold" size="base">
          {title}
        </Text>
        {summary && (
          <Text size="sm" color="muted">
            {summary}
          </Text>
        )}
      </div>
      {children}
    </Card>
  );
}
