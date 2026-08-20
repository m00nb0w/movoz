"use client";

import { useMemo, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { api } from "@/lib/api";

interface ChatTurn {
  role: "user" | "assistant";
  content: string;
}

interface ProposedRankingEntry {
  engineer_id: number;
  rank: number;
}

function extractJSONBlock(text: string): { rationale?: string; ranking: ProposedRankingEntry[] } | null {
  const matches = [...text.matchAll(/```json\s*([\s\S]*?)\s*```/g)];
  if (matches.length === 0) return null;
  try {
    return JSON.parse(matches[matches.length - 1][1]);
  } catch {
    return null;
  }
}

export default function AIRankingChatPage() {
  const params = useParams<{ id: string; subId: string }>();
  const cycleId = Number(params.id);
  const subAttributeId = Number(params.subId);
  const router = useRouter();

  const [sessionId, setSessionId] = useState<number | null>(null);
  const [turns, setTurns] = useState<ChatTurn[]>([]);
  const [input, setInput] = useState("");
  const [streaming, setStreaming] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [proposedRanking, setProposedRanking] = useState<ProposedRankingEntry[] | null>(null);
  const [accepting, setAccepting] = useState(false);
  const assistantBufferRef = useRef("");

  async function sendMessage() {
    if (!input.trim() || streaming) return;
    const userMessage = input;
    setInput("");
    setError(null);
    setTurns((prev) => [...prev, { role: "user", content: userMessage }, { role: "assistant", content: "" }]);
    setStreaming(true);
    assistantBufferRef.current = "";

    try {
      const res = await fetch(`/scout/api/cycles/${cycleId}/ai-sessions`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ session_id: sessionId, sub_attribute_id: subAttributeId, message: userMessage }),
      });
      if (!res.ok) {
        // Several backend validation paths (invalid/missing cycle or
        // sub-attribute, missing message, session lookup failures) return a
        // plain JSON error body with a non-2xx status *before* any SSE
        // framing is written, so there is no "\n\n"-delimited frame to parse
        // here — this must be handled the same way src/lib/api.ts handles
        // every other non-2xx response, or the error silently vanishes.
        const body = await res.json().catch(() => ({}));
        throw new Error(body.error || `Request failed with status ${res.status}`);
      }
      if (!res.body) throw new Error("no response body");

      const reader = res.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });

        const frames = buffer.split("\n\n");
        buffer = frames.pop() ?? "";

        for (const frame of frames) {
          const lines = frame.split("\n");
          const eventLine = lines.find((l) => l.startsWith("event: "));
          const dataLine = lines.find((l) => l.startsWith("data: "));
          if (!dataLine) continue;
          const data = dataLine.slice("data: ".length);

          if (eventLine?.includes("session")) {
            setSessionId(JSON.parse(data).session_id);
          } else if (eventLine?.includes("error")) {
            setError(JSON.parse(data).error);
          } else {
            assistantBufferRef.current += data.replace(/\\n/g, "\n");
            const snapshot = assistantBufferRef.current;
            setTurns((prev) => {
              const next = [...prev];
              next[next.length - 1] = { role: "assistant", content: snapshot };
              return next;
            });
          }
        }
      }

      const block = extractJSONBlock(assistantBufferRef.current);
      if (block) {
        setProposedRanking(block.ranking);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Chat failed");
    } finally {
      setStreaming(false);
    }
  }

  function updateProposedRank(engineerId: number, rank: number) {
    setProposedRanking((prev) => (prev ?? []).map((entry) => (entry.engineer_id === engineerId ? { ...entry, rank } : entry)));
  }

  // Mirrors the validation in sub-attributes/[subId]/page.tsx: ranks must be
  // a strict 1..N permutation with no ties. The backend enforces this too
  // (SubmitRanking rejects an invalid permutation), but disabling Accept
  // client-side gives the manager immediate feedback instead of a round-trip
  // error, matching the sibling ranking page's UX.
  const proposedRanks = useMemo(() => (proposedRanking ?? []).map((entry) => entry.rank), [proposedRanking]);
  const hasDuplicateRank = new Set(proposedRanks).size !== proposedRanks.length;
  const hasOutOfRangeRank = proposedRanks.some(
    (r) => !Number.isInteger(r) || r < 1 || r > (proposedRanking?.length ?? 0)
  );
  const proposedRankingInvalid = hasDuplicateRank || hasOutOfRangeRank;

  async function acceptRanking() {
    if (!sessionId || !proposedRanking || proposedRankingInvalid) return;
    setAccepting(true);
    setError(null);
    try {
      await api.post(`/api/cycles/${cycleId}/ai-sessions/${sessionId}/accept`, { rankings: proposedRanking });
      router.push(`/cycles/${cycleId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to accept ranking");
    } finally {
      setAccepting(false);
    }
  }

  return (
    <main className="mx-auto max-w-2xl p-8">
      <h1 className="mb-6 text-2xl font-semibold text-zen-text">AI Ranking Assistant</h1>

      <div className="mb-4 max-h-96 space-y-3 overflow-y-auto rounded-lg border border-zen-border p-4">
        {turns.length === 0 && (
          <p className="text-sm text-zen-muted">
            Describe what you observed this cycle — who stood out, who struggled — and the assistant will propose a
            ranking with rationale.
          </p>
        )}
        {turns.map((turn, i) => (
          <p key={i} className={turn.role === "user" ? "text-zen-text" : "text-zen-muted"}>
            <strong>{turn.role === "user" ? "You" : "Assistant"}:</strong> {turn.content}
          </p>
        ))}
      </div>

      {proposedRanking && (
        <div className="mb-4 rounded-lg border border-accent-600 p-4">
          <h2 className="mb-2 font-medium text-zen-text">Proposed ranking (edit before accepting)</h2>
          <ul className="space-y-2">
            {proposedRanking.map((entry) => (
              <li key={entry.engineer_id} className="flex items-center justify-between">
                <span className="text-zen-text">Engineer #{entry.engineer_id}</span>
                <input
                  type="number"
                  min={1}
                  max={proposedRanking.length}
                  value={entry.rank}
                  onChange={(e) => updateProposedRank(entry.engineer_id, Number(e.target.value))}
                  className="w-16 rounded border border-zen-border bg-transparent p-1 text-center"
                />
              </li>
            ))}
          </ul>
          {hasDuplicateRank && (
            <p className="mt-3 text-sm text-red-500">
              Two engineers share the same rank — ranks must be unique 1..{proposedRanking.length}.
            </p>
          )}
          {hasOutOfRangeRank && !hasDuplicateRank && (
            <p className="mt-3 text-sm text-red-500">Ranks must be between 1 and {proposedRanking.length}.</p>
          )}
          {/* NF3: the ranking is only persisted when the admin explicitly clicks
              this button — nothing here is auto-applied from the chat. */}
          <button
            onClick={acceptRanking}
            disabled={accepting || proposedRankingInvalid}
            className="mt-3 rounded bg-accent-600 px-4 py-2 text-white disabled:opacity-50"
          >
            {accepting ? "Saving..." : "Accept & save ranking"}
          </button>
        </div>
      )}

      {error && <p className="mb-4 text-sm text-red-500">{error}</p>}

      <div className="flex gap-2">
        <input
          className="flex-1 rounded border border-zen-border bg-transparent p-2"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && sendMessage()}
          placeholder="Describe this cycle's observations..."
          disabled={streaming}
        />
        <button
          onClick={sendMessage}
          disabled={streaming}
          className="rounded bg-accent-600 px-4 py-2 text-white disabled:opacity-50"
        >
          {streaming ? "..." : "Send"}
        </button>
      </div>
    </main>
  );
}
