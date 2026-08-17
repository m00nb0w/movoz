import { getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";
import Link from "next/link";
import { api, ApiError } from "@/lib/api/client";
import { StatTile } from "@/components/StatTile";
import { Badge } from "@movoz/ui-web";

export default async function PlayerProfilePage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const t = await getTranslations("playerProfile");
  const tp = await getTranslations("positions");

  let profile;
  try {
    profile = await api.getPlayerProfile(Number(id));
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) notFound();
    return (
      <main className="mx-auto max-w-5xl px-4 py-24 text-center">
        <p className="text-zen-muted">{t("loadError")}</p>
      </main>
    );
  }

  return (
    <main className="mx-auto max-w-5xl px-4 py-12">
      <Link href="/leaderboard" className="text-sm text-zen-muted hover:text-zen-text">
        {t("back")}
      </Link>

      <div className="mb-6 mt-4 flex flex-wrap items-center gap-3">
        <h1 className="font-serif text-3xl font-bold text-zen-text">{profile.player.name}</h1>
        {profile.player.position && <Badge variant="subtle">{tp(profile.player.position)}</Badge>}
        {!profile.player.is_active && <Badge variant="outline">{t("inactive")}</Badge>}
      </div>

      <hr className="border-t-2 border-zen-border" />

      <div className="my-8 grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
        <StatTile label={t("matchesPlayed")} value={profile.all_time.matches_played} />
        <StatTile label={t("goals")} value={profile.all_time.goals} />
        <StatTile label={t("assists")} value={profile.all_time.assists} />
        <StatTile label={t("yellowCards")} value={profile.all_time.yellow_cards} />
        <StatTile label={t("redCards")} value={profile.all_time.red_cards} />
      </div>

      <h2 className="mb-4 text-xl font-semibold text-zen-text">{t("seasonBySeason")}</h2>
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b-2 border-zen-border text-left text-zen-muted">
            <th className="py-2">{t("year")}</th>
            <th className="py-2 text-right">{t("matches")}</th>
            <th className="py-2 text-right">{t("goals")}</th>
            <th className="py-2 text-right">{t("assists")}</th>
            <th className="py-2 text-right">{t("yellow")}</th>
            <th className="py-2 text-right">{t("red")}</th>
          </tr>
        </thead>
        <tbody>
          {profile.by_year.map((row) => (
            <tr key={row.year} className="border-b border-zen-border">
              <td className="py-2">{row.year}</td>
              <td className="py-2 text-right">{row.matches_played}</td>
              <td className="py-2 text-right">{row.goals}</td>
              <td className="py-2 text-right">{row.assists}</td>
              <td className="py-2 text-right">{row.yellow_cards}</td>
              <td className="py-2 text-right">{row.red_cards}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </main>
  );
}
