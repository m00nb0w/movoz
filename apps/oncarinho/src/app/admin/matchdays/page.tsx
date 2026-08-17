"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useLocale, useTranslations } from "next-intl";
import { Button, Input, Card } from "@movoz/ui-web";
import { api, ApiError } from "@/lib/api/client";
import type { Matchday } from "@/lib/api/types";

export default function AdminMatchdaysPage() {
  const t = useTranslations("admin.matchdays");
  const locale = useLocale();
  const dateFormatter = new Intl.DateTimeFormat(locale, { dateStyle: "medium", timeZone: "UTC" });
  const [matchdays, setMatchdays] = useState<Matchday[]>([]);
  const [newDate, setNewDate] = useState("");
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .getMatchdays()
      .then(setMatchdays)
      .catch((err) => {
        console.error("admin matchdays: failed to load", err);
        setError(t("loadError"));
      });
  }, [t]);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setCreating(true);
    setError(null);
    try {
      const matchday = await api.createMatchday(newDate);
      setMatchdays((prev) => [matchday, ...prev]);
      setNewDate("");
    } catch (err) {
      console.error("admin matchdays: failed to create", err);
      setError(err instanceof ApiError && err.status === 400 ? t("invalidInput") : t("loadError"));
    } finally {
      setCreating(false);
    }
  }

  return (
    <main className="mx-auto max-w-3xl px-4 py-12">
      <h1 className="mb-6 font-serif text-3xl font-bold text-zen-text">{t("title")}</h1>

      <Card padding="md" className="mb-8">
        <form onSubmit={handleCreate} className="flex flex-wrap items-end gap-4">
          <Input
            type="date"
            label={t("newMatchdayLabel")}
            value={newDate}
            onChange={(e) => setNewDate(e.target.value)}
            required
          />
          <Button type="submit" loading={creating}>
            {t("create")}
          </Button>
        </form>
        {error && <p className="mt-2 text-sm text-red-500">{error}</p>}
      </Card>

      <ul>
        {matchdays.map((matchday) => (
          <li key={matchday.id} className="border-b border-zen-border py-3">
            <Link href={`/admin/matchdays/${matchday.id}`} className="hover:underline">
              {dateFormatter.format(new Date(matchday.played_on))}
            </Link>
          </li>
        ))}
      </ul>
    </main>
  );
}
