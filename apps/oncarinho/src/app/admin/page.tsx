"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { Button, Input, Card } from "@movoz/ui-web";
import { api, ApiError } from "@/lib/api/client";

export default function AdminLoginPage() {
  const t = useTranslations("admin.login");
  const router = useRouter();
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await api.login(password);
      router.push("/admin/matchdays");
    } catch (err) {
      if (!(err instanceof ApiError && err.status === 401)) console.error(err);
      setError(
        err instanceof ApiError && err.status === 401 ? t("incorrectPassword") : t("genericError")
      );
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="mx-auto flex max-w-sm flex-col px-4 py-24">
      <Card padding="lg">
        <h1 className="mb-4 font-serif text-2xl font-bold text-zen-text">{t("title")}</h1>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <Input
            type="password"
            label={t("passwordLabel")}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            error={error ?? undefined}
            required
          />
          <Button type="submit" loading={loading}>
            {t("submit")}
          </Button>
        </form>
      </Card>
    </main>
  );
}
