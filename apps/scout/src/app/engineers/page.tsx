"use client";

import { useEffect, useState, type FormEvent } from "react";
import Link from "next/link";
import { Badge, Button, Card, Container, Input, Text } from "@movoz/ui-web";
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
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editName, setEditName] = useState("");
  const [editRole, setEditRole] = useState("");
  const [editGithubUsername, setEditGithubUsername] = useState("");
  const [editJiraAccountId, setEditJiraAccountId] = useState("");
  const [editStartedAt, setEditStartedAt] = useState("");
  const [editError, setEditError] = useState<string | null>(null);

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

  function startEdit(engineer: Engineer) {
    setEditingId(engineer.id);
    setEditName(engineer.name);
    setEditRole(engineer.role ?? "");
    setEditGithubUsername(engineer.github_username ?? "");
    setEditJiraAccountId(engineer.jira_account_id ?? "");
    setEditStartedAt(engineer.started_at.slice(0, 10));
    setEditError(null);
  }

  function cancelEdit() {
    setEditingId(null);
    setEditError(null);
  }

  async function saveEdit(e: FormEvent, id: number) {
    e.preventDefault();
    setEditError(null);
    try {
      await api.put(`/api/engineers/${id}`, {
        name: editName,
        role: editRole || null,
        github_username: editGithubUsername || null,
        jira_account_id: editJiraAccountId || null,
        started_at: editStartedAt,
      });
      setEditingId(null);
      await load();
    } catch (err) {
      setEditError(err instanceof Error ? err.message : "Failed to save changes");
    }
  }

  return (
    <Container maxWidth="md" className="py-12">
      <Text as="h1" font="marker" size="2xl" weight="bold" className="mb-6">
        Roster
      </Text>

      <Card variant="outlined" className="mb-8">
        <form onSubmit={handleCreate} className="grid grid-cols-2 gap-3">
          <Input placeholder="Name" value={name} onChange={(e) => setName(e.target.value)} required />
          <Input placeholder="Role" value={role} onChange={(e) => setRole(e.target.value)} />
          <Input placeholder="GitHub username" value={githubUsername} onChange={(e) => setGithubUsername(e.target.value)} />
          <Input placeholder="Jira account ID" value={jiraAccountId} onChange={(e) => setJiraAccountId(e.target.value)} />
          <Input type="date" value={startedAt} onChange={(e) => setStartedAt(e.target.value)} required />
          <Button type="submit">Add engineer</Button>
          {error && (
            <Text size="sm" color="accent" className="col-span-2">
              {error}
            </Text>
          )}
        </form>
      </Card>

      <label className="mb-3 flex items-center gap-2 text-sm text-ink-soft">
        <input type="checkbox" checked={showAll} onChange={(e) => setShowAll(e.target.checked)} />
        Show deactivated engineers
      </label>

      <Card variant="outlined" padding="none">
        {engineers.map((engineer, i) =>
          editingId === engineer.id ? (
            <div key={engineer.id} className={`p-4 ${i > 0 ? "border-t border-line" : ""}`}>
              <form onSubmit={(e) => saveEdit(e, engineer.id)} className="grid grid-cols-2 gap-3">
                <Input placeholder="Name" value={editName} onChange={(e) => setEditName(e.target.value)} required />
                <Input placeholder="Role" value={editRole} onChange={(e) => setEditRole(e.target.value)} />
                <Input
                  placeholder="GitHub username"
                  value={editGithubUsername}
                  onChange={(e) => setEditGithubUsername(e.target.value)}
                />
                <Input
                  placeholder="Jira account ID"
                  value={editJiraAccountId}
                  onChange={(e) => setEditJiraAccountId(e.target.value)}
                />
                <Input type="date" value={editStartedAt} onChange={(e) => setEditStartedAt(e.target.value)} required />
                <div className="flex gap-2">
                  <Button type="submit">Save</Button>
                  <Button type="button" variant="secondary" onClick={cancelEdit}>
                    Cancel
                  </Button>
                </div>
                {editError && (
                  <Text size="sm" color="accent" className="col-span-2">
                    {editError}
                  </Text>
                )}
              </form>
            </div>
          ) : (
            <div key={engineer.id} className={`flex items-center justify-between p-4 ${i > 0 ? "border-t border-line" : ""}`}>
              <div className="flex items-center gap-2">
                <Link href={`/engineers/${engineer.id}`} className="font-medium text-ink hover:underline">
                  {engineer.name}
                </Link>
                {engineer.role && (
                  <Text size="sm" color="muted">
                    {engineer.role}
                  </Text>
                )}
                {!engineer.is_active && (
                  <Badge variant="subtle" color="danger" size="sm">
                    deactivated
                  </Badge>
                )}
              </div>
              <div className="flex gap-3">
                <Button variant="ghost" size="sm" onClick={() => startEdit(engineer)}>
                  Edit
                </Button>
                <Button variant="ghost" size="sm" onClick={() => toggleActive(engineer)}>
                  {engineer.is_active ? "Deactivate" : "Reactivate"}
                </Button>
              </div>
            </div>
          )
        )}
      </Card>
    </Container>
  );
}
