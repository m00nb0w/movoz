"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Button, Input, Badge, Dropdown } from "@movoz/ui-web";
import { api } from "@/lib/api/client";
import type { Player, Position } from "@/lib/api/types";

const POSITIONS: Position[] = ["goalkeeper", "defender", "midfielder", "forward"];

export default function AdminPlayersPage() {
  const t = useTranslations("admin.players");
  const tp = useTranslations("positions");
  const [players, setPlayers] = useState<Player[]>([]);
  const [showInactive, setShowInactive] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [name, setName] = useState("");
  const [position, setPosition] = useState<Position | null>(null);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.getPlayers(showInactive).then(setPlayers).catch(() => setError(t("loadError")));
  }, [showInactive, t]);

  function startEdit(player: Player) {
    setEditingId(player.id);
    setName(player.name);
    setPosition(player.position);
  }

  function cancelEdit() {
    setEditingId(null);
    setName("");
    setPosition(null);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);
    try {
      if (editingId !== null) {
        const updated = await api.updatePlayer(editingId, name, position);
        setPlayers((prev) => prev.map((p) => (p.id === editingId ? updated : p)));
        cancelEdit();
      } else {
        const player = await api.createPlayer(name, position);
        setPlayers((prev) => [...prev, player]);
        setName("");
        setPosition(null);
      }
    } catch {
      setError(t("loadError"));
    } finally {
      setSaving(false);
    }
  }

  async function toggleActive(player: Player) {
    try {
      if (player.is_active) {
        await api.deactivatePlayer(player.id);
      } else {
        await api.reactivatePlayer(player.id);
      }
      setPlayers((prev) =>
        prev.map((p) => (p.id === player.id ? { ...p, is_active: !p.is_active } : p))
      );
    } catch {
      setError(t("loadError"));
    }
  }

  return (
    <main className="mx-auto max-w-3xl px-4 py-12">
      <h1 className="mb-6 font-serif text-3xl font-bold text-zen-text">{t("title")}</h1>
      {error && <p className="mb-4 text-sm text-red-500">{error}</p>}

      <form onSubmit={handleSubmit} className="mb-8 flex flex-wrap items-end gap-4">
        <Input label={t("nameLabel")} value={name} onChange={(e) => setName(e.target.value)} required />
        <Dropdown
          trigger={<Button variant="secondary">{position ? tp(position) : t("positionLabel")}</Button>}
          items={POSITIONS.map((p) => ({ label: tp(p), value: p }))}
          onSelect={(value) => setPosition(value as Position)}
        />
        <Button type="submit" loading={saving}>
          {editingId !== null ? t("save") : t("add")}
        </Button>
        {editingId !== null && (
          <Button type="button" variant="ghost" onClick={cancelEdit}>
            {t("cancel")}
          </Button>
        )}
      </form>

      <label className="mb-4 flex items-center gap-2 text-sm text-zen-muted">
        <input
          type="checkbox"
          checked={showInactive}
          onChange={(e) => setShowInactive(e.target.checked)}
        />
        {t("showInactive")}
      </label>

      <table className="w-full text-sm">
        <thead>
          <tr className="border-b-2 border-zen-border text-left text-zen-muted">
            <th className="py-2">{t("name")}</th>
            <th className="py-2">{t("position")}</th>
            <th className="py-2">{t("status")}</th>
            <th className="py-2" />
          </tr>
        </thead>
        <tbody>
          {players.map((player) => (
            <tr key={player.id} className="border-b border-zen-border">
              <td className="py-2">{player.name}</td>
              <td className="py-2">{player.position ? tp(player.position) : "—"}</td>
              <td className="py-2">
                <Badge variant="subtle" color={player.is_active ? "accent" : "default"} size="sm">
                  {player.is_active ? t("active") : t("inactive")}
                </Badge>
              </td>
              <td className="py-2 flex gap-2">
                <Button variant="ghost" size="sm" onClick={() => startEdit(player)}>
                  {t("edit")}
                </Button>
                <Button variant="ghost" size="sm" onClick={() => toggleActive(player)}>
                  {player.is_active ? t("deactivate") : t("reactivate")}
                </Button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </main>
  );
}
