const MS_PER_DAY = 24 * 60 * 60 * 1000;

function startOfDay(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate());
}

/**
 * Red if overdue, yellow if due within the next 3 days inclusive, otherwise
 * undefined (FR-020). Compares calendar days, not timestamps, so "today" and
 * "3 days from today" are never off-by-one due to time-of-day.
 */
export function getDueDateColor(dueDate: string | null): "red" | "yellow" | undefined {
  if (!dueDate) return undefined;

  const due = startOfDay(new Date(`${dueDate}T00:00:00`));
  const today = startOfDay(new Date());
  const diffDays = Math.round((due.getTime() - today.getTime()) / MS_PER_DAY);

  if (diffDays < 0) return "red";
  if (diffDays <= 3) return "yellow";
  return undefined;
}
