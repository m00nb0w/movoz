import { WorkItemList } from "@/components/WorkItemList";

export default function ListPage() {
  return (
    <main className="mx-auto max-w-6xl px-4 py-8">
      <h1 className="mb-6 text-2xl font-bold text-ink">All Work Items</h1>
      <WorkItemList />
    </main>
  );
}
