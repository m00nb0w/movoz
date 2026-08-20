"use client";

import { useEffect, useState, type FormEvent } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
import type { Engineer } from "@/lib/types";

export default function EngineersPage() {
  const [engineers, setEngineers] = useState<Engineer[]>([]);
  const [showAll, setShowAll] = useState(false);
  const [name, setName] = useState("");
  const [role, setRole] = useState("");
  const [githubUsername, setGithubUsername] = useState("");
  const [jiraAccountId, setJiraAccountId] = useState("");
  const [startedAt, setStartedAt] = useState("");
  const [error, setError] = useState<string | null>(null);

  async function load() {
    const list = await api.get<Engineer[]>(`/api/engineers?active=${showAll ? "all" : "true"}`);
    setEngineers(list);
  }

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [showAll]);

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await api.post("/api/engineers", {
        name,
        role: role || null,
        github_username: githubUsername || null,
        jira_account_id: jiraAccountId || null,
        started_at: startedAt,
      });
      setName("");
      setRole("");
      setGithubUsername("");
      setJiraAccountId("");
      setStartedAt("");
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to add engineer");
    }
  }

  async function toggleActive(engineer: Engineer) {
    if (engineer.is_active) {
      await api.delete(`/api/engineers/${engineer.id}`);
    } else {
      await api.post(`/api/engineers/${engineer.id}/reactivate`);
    }
    await load();
  }

  return (
    <main className="mx-auto max-w-3xl p-8">
      <h1 className="mb-6 text-2xl font-semibold text-zen-text">Roster</h1>

      <form onSubmit={handleCreate} className="mb-8 grid grid-cols-2 gap-3 rounded-lg border border-zen-border p-4">
        <input className="rounded border border-zen-border bg-transparent p-2" placeholder="Name" value={name} onChange={(e) => setName(e.target.value)} required />
        <input className="rounded border border-zen-border bg-transparent p-2" placeholder="Role" value={role} onChange={(e) => setRole(e.target.value)} />
        <input className="rounded border border-zen-border bg-transparent p-2" placeholder="GitHub username" value={githubUsername} onChange={(e) => setGithubUsername(e.target.value)} />
        <input className="rounded border border-zen-border bg-transparent p-2" placeholder="Jira account ID" value={jiraAccountId} onChange={(e) => setJiraAccountId(e.target.value)} />
        <input className="rounded border border-zen-border bg-transparent p-2" type="date" value={startedAt} onChange={(e) => setStartedAt(e.target.value)} required />
        <button type="submit" className="rounded bg-accent-600 px-3 py-2 text-white">Add engineer</button>
        {error && <p className="col-span-2 text-sm text-red-500">{error}</p>}
      </form>

      <label className="mb-3 flex items-center gap-2 text-sm text-zen-muted">
        <input type="checkbox" checked={showAll} onChange={(e) => setShowAll(e.target.checked)} />
        Show deactivated engineers
      </label>

      <ul className="divide-y divide-zen-border">
        {engineers.map((engineer) => (
          <li key={engineer.id} className="flex items-center justify-between py-3">
            <div>
              <Link href={`/engineers/${engineer.id}`} className="font-medium text-zen-text hover:underline">
                {engineer.name}
              </Link>
              <span className="ml-2 text-sm text-zen-muted">{engineer.role}</span>
              {!engineer.is_active && <span className="ml-2 text-xs text-red-500">deactivated</span>}
            </div>
            <button onClick={() => toggleActive(engineer)} className="text-sm text-zen-muted hover:text-zen-text">
              {engineer.is_active ? "Deactivate" : "Reactivate"}
            </button>
          </li>
        ))}
      </ul>
    </main>
  );
}
