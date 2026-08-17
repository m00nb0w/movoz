import { getTranslations } from "next-intl/server";
import Link from "next/link";
import { api } from "@/lib/api/client";
import { availableYears } from "@/lib/years";
import { StatTile } from "@/components/StatTile";
import { SeasonSelector } from "@/components/SeasonSelector";
import { Badge, Button } from "@movoz/ui-web";

export default async function DashboardPage({
  searchParams,
}: {
  searchParams: Promise<{ year?: string }>;
}) {
  const t = await getTranslations("dashboard");
  const tp = await getTranslations("positions");
  const params = await searchParams;
  const currentYear = new Date().getUTCFullYear();
  const parsedYear = params.year ? parseInt(params.year, 10) : NaN;
  const year = Number.isInteger(parsedYear) && parsedYear > 0 ? parsedYear : currentYear;

  let summary, players, leaderboard, matchdays;
  try {
    [summary, players, leaderboard, matchdays] = await Promise.all([
      api.getSummary(year),
      api.getPlayers(),
      api.getLeaderboard(year, "goals"),
      api.getMatchdays(),
    ]);
  } catch (err) {
    console.error("dashboard: failed to load data", err);
    return (
      <main className="mx-auto max-w-5xl px-4 py-12 text-center">
        <p className="text-zen-muted">{t("loadError")}</p>
      </main>
    );
  }

  const years = availableYears(matchdays);
  const topScorers = leaderboard.slice(0, 5);

  return (
    <main className="mx-auto max-w-5xl px-4 py-12">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <h1 className="font-serif text-3xl font-bold text-zen-text">{t("title")}</h1>
        <SeasonSelector years={years} selected={year} basePath="/" />
      </div>

      <div className="my-8 grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatTile label={t("matchesPlayed")} value={summary.matches_played} />
        <StatTile label={t("goalsScored")} value={summary.goals_scored} />
        <StatTile label={t("rosterSize")} value={summary.roster_size} />
      </div>

      <hr className="border-t-2 border-zen-border" />

      <div className="mt-8 grid grid-cols-1 gap-12 md:grid-cols-[1.2fr_1fr]">
        <section>
          <h2 className="mb-4 text-xl font-semibold text-zen-text">
            {t("topScorers", { year })}
          </h2>
          {topScorers.length === 0 ? (
            <p className="text-zen-muted">{t("noGoals", { year })}</p>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-zen-border text-left text-zen-muted">
                  <th className="py-2">#</th>
                  <th className="py-2">{t("player")}</th>
                  <th className="py-2 text-right">{t("goals")}</th>
                </tr>
              </thead>
              <tbody>
                {topScorers.map((entry, i) => (
                  <tr key={entry.player_id} className="border-b border-zen-border">
                    <td className="py-2">{i + 1}</td>
                    <td className="py-2">
                      <Link href={`/players/${entry.player_id}`} className="hover:underline">
                        {entry.player_name}
                      </Link>
                    </td>
                    <td className="py-2 text-right">{entry.value}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
          <Link href="/leaderboard">
            <Button variant="ghost" size="sm" className="mt-4">
              {t("viewFullLeaderboard")}
            </Button>
          </Link>
        </section>

        <section>
          <h2 className="mb-4 text-xl font-semibold text-zen-text">{t("roster")}</h2>
          <ul>
            {players.map((player) => (
              <li
                key={player.id}
                className="flex items-center justify-between border-b border-zen-border py-2"
              >
                <Link href={`/players/${player.id}`} className="hover:underline">
                  {player.name}
                </Link>
                {player.position && (
                  <Badge variant="subtle" size="sm">
                    {tp(player.position)}
                  </Badge>
                )}
              </li>
            ))}
          </ul>
        </section>
      </div>
    </main>
  );
}
