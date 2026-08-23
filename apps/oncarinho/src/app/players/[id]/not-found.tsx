import { getTranslations } from "next-intl/server";
import Link from "next/link";

export default async function PlayerNotFound() {
  const t = await getTranslations("playerProfile");

  return (
    <main className="mx-auto max-w-5xl px-4 py-24 text-center">
      <p className="text-ink-soft">{t("notFound")}</p>
      <Link href="/leaderboard" className="mt-4 inline-block text-ink hover:underline">
        {t("back")}
      </Link>
    </main>
  );
}
