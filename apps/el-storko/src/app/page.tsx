import { KanbanBoard } from "@/components/KanbanBoard";

export default function BoardPage() {
  return (
    <main className="mx-auto max-w-6xl px-4 py-8">
      <h1 className="mb-6 text-2xl font-bold text-ink">el-storko</h1>
      <KanbanBoard />
    </main>
  );
}
