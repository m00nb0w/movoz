import { getTranslations } from "next-intl/server";
import Link from "next/link";
import { api } from "@/lib/api/client";
import { availableYears } from "@/lib/years";
import { SeasonSelector } from "@/components/SeasonSelector";
import { StatTypeTabs } from "@/components/StatTypeTabs";
import { Badge } from "@movoz/ui-web";

type Stat = "goals" | "assists" | "cards";
const VALID_STATS: Stat[] = ["goals", "assists", "cards"];

export default async function LeaderboardPage({
  searchParams,
}: {
  searchParams: Promise<{ year?: string; stat?: string }>;
}) {
  const t = await getTranslations("leaderboard");
  const tp = await getTranslations("positions");
  const params = await searchParams;

  const parsedYear =
    params.year && params.year !== "all" ? parseInt(params.year, 10) : "all";
  const year: number | "all" =
    parsedYear === "all" || (Number.isInteger(parsedYear) && parsedYear > 0) ? parsedYear : "all";
  const stat: Stat = VALID_STATS.includes(params.stat as Stat) ? (params.stat as Stat) : "goals";

  let entries, matchdays;
  try {
    [entries, matchdays] = await Promise.all([
      api.getLeaderboard(year, stat),
      api.getMatchdays(),
    ]);
  } catch (err) {
    console.error("leaderboard: failed to load data", err);
    return (
      <main className="mx-auto max-w-5xl px-4 py-12 text-center">
        <p className="text-zen-muted">{t("loadError")}</p>
      </main>
    );
  }
  const years = availableYears(matchdays);

  return (
    <main className="mx-auto max-w-5xl px-4 py-12">
      <h1 className="font-serif text-3xl font-bold text-zen-text">{t("title")}</h1>

      <div className="my-6 flex flex-wrap gap-4">
        <SeasonSelector
          years={years}
          selected={year}
          basePath="/leaderboard"
          includeAllTime
          extraParams={{ stat }}
        />
        <StatTypeTabs selected={stat} basePath="/leaderboard" extraParams={{ year: String(year) }} />
      </div>

      {entries.length === 0 ? (
        <p className="text-zen-muted">{t("empty")}</p>
      ) : (
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b-2 border-zen-border text-left text-zen-muted">
              <th className="py-2">#</th>
              <th className="py-2">{t("player")}</th>
              <th className="py-2">{t("position")}</th>
              <th className="py-2">{t("status")}</th>
              <th className="py-2 text-right">{t(stat)}</th>
            </tr>
          </thead>
          <tbody>
            {entries.map((entry, i) => (
              <tr key={entry.player_id} className="border-b border-zen-border">
                <td className="py-2">{i + 1}</td>
                <td className="py-2">
                  <Link href={`/players/${entry.player_id}`} className="hover:underline">
                    {entry.player_name}
                  </Link>
                </td>
                <td className="py-2">{entry.position ? tp(entry.position) : "—"}</td>
                <td className="py-2">
                  <Badge variant="subtle" color={entry.is_active ? "accent" : "default"} size="sm">
                    {entry.is_active ? t("active") : t("inactive")}
                  </Badge>
                </td>
                <td className="py-2 text-right font-medium">{entry.value}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </main>
  );
}
