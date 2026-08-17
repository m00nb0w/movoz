"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { Button, IconButton } from "@movoz/ui-web";
import { Trash2 } from "lucide-react";
import { api } from "@/lib/api/client";
import type { Player, MatchStat } from "@/lib/api/types";

interface Row {
  playerId: number;
  goals: number;
  assists: number;
  yellowCards: number;
  redCards: number;
}

const FIELDS = ["goals", "assists", "yellowCards", "redCards"] as const;

export default function AdminMatchdayStatsPage() {
  const t = useTranslations("admin.matchdays");
  const params = useParams<{ id: string }>();
  const matchdayId = Number(params.id);

  const [players, setPlayers] = useState<Player[]>([]);
  const [rows, setRows] = useState<Row[]>([]);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([api.getPlayers(), api.getMatchdayStats(matchdayId)])
      .then(([activePlayers, stats]) => {
        setPlayers(activePlayers);
        const byPlayer = new Map(stats.map((s: MatchStat) => [s.player_id, s]));
        setRows(
          activePlayers.map((p) => {
            const existing = byPlayer.get(p.id);
            return {
              playerId: p.id,
              goals: existing?.goals ?? 0,
              assists: existing?.assists ?? 0,
              yellowCards: existing?.yellow_cards ?? 0,
              redCards: existing?.red_cards ?? 0,
            };
          })
        );
      })
      .catch(() => setError(t("loadError")));
  }, [matchdayId, t]);

  function updateRow(playerId: number, field: (typeof FIELDS)[number], value: number) {
    setRows((prev) =>
      prev.map((row) => (row.playerId === playerId ? { ...row, [field]: value } : row))
    );
  }

  async function handleRemove(playerId: number) {
    try {
      await api.deleteMatchdayStat(matchdayId, playerId);
      setRows((prev) => prev.filter((row) => row.playerId !== playerId));
    } catch {
      setError(t("loadError"));
    }
  }

  async function handleSave() {
    setSaving(true);
    setError(null);
    try {
      await api.upsertMatchdayStats(
        matchdayId,
        rows.map((row) => ({
          player_id: row.playerId,
          goals: row.goals,
          assists: row.assists,
          yellow_cards: row.yellowCards,
          red_cards: row.redCards,
        }))
      );
    } catch {
      setError(t("loadError"));
    } finally {
      setSaving(false);
    }
  }

  return (
    <main className="mx-auto max-w-4xl px-4 py-12">
      <h1 className="mb-6 font-serif text-3xl font-bold text-zen-text">{t("statsTitle")}</h1>
      {error && <p className="mb-4 text-sm text-red-500">{error}</p>}

      <table className="w-full text-sm">
        <thead>
          <tr className="border-b-2 border-zen-border text-left text-zen-muted">
            <th className="py-2">{t("player")}</th>
            <th className="py-2 text-right">{t("goals")}</th>
            <th className="py-2 text-right">{t("assists")}</th>
            <th className="py-2 text-right">{t("yellow")}</th>
            <th className="py-2 text-right">{t("red")}</th>
            <th className="py-2" />
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => {
            const player = players.find((p) => p.id === row.playerId);
            return (
              <tr key={row.playerId} className="border-b border-zen-border">
                <td className="py-2">{player?.name}</td>
                {FIELDS.map((field) => (
                  <td key={field} className="py-2 text-right">
                    <input
                      type="number"
                      min={0}
                      value={row[field]}
                      onChange={(e) => updateRow(row.playerId, field, Number(e.target.value))}
                      className="w-16 border border-zen-border bg-zen-bg px-2 py-1 text-right text-zen-text"
                    />
                  </td>
                ))}
                <td className="py-2">
                  <IconButton
                    icon={<Trash2 size={16} />}
                    label={t("remove")}
                    onClick={() => handleRemove(row.playerId)}
                  />
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>

      <Button className="mt-6" onClick={handleSave} loading={saving}>
        {t("save")}
      </Button>
    </main>
  );
}
