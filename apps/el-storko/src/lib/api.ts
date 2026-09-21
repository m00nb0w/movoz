export type WorkItemType = "epic" | "task";
export type WorkItemStatus = "todo" | "in_progress" | "blocked" | "done";
export type WorkItemSource = "personal" | "jira" | "agent";

export interface WorkItem {
  id: number;
  type: WorkItemType;
  parent_id: number | null;
  title: string;
  description: string;
  status: WorkItemStatus;
  source: WorkItemSource;
  jira_key: string | null;
  jira_url: string | null;
  created_at: string;
  updated_at: string;
}

export interface ListFilters {
  source?: WorkItemSource;
  parent_id?: number;
  type?: WorkItemType;
  status?: WorkItemStatus;
}

export interface CreateWorkItemInput {
  type: WorkItemType;
  title: string;
  description?: string;
  parent_id?: number;
  status?: WorkItemStatus;
}

export interface UpdateWorkItemInput {
  title?: string;
  description?: string;
  status?: WorkItemStatus;
  parent_id?: number | null;
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`el-storko API error ${res.status}: ${body}`);
  }
  return res.json() as Promise<T>;
}

export async function listWorkItems(filters: ListFilters = {}): Promise<WorkItem[]> {
  const params = new URLSearchParams();
  if (filters.source) params.set("source", filters.source);
  if (filters.parent_id !== undefined) params.set("parent_id", String(filters.parent_id));
  if (filters.type) params.set("type", filters.type);
  if (filters.status) params.set("status", filters.status);

  const query = params.toString();
  const res = await fetch(`/api/work-items${query ? `?${query}` : ""}`, { cache: "no-store" });
  const data = await handleResponse<{ items: WorkItem[] }>(res);
  return data.items;
}

export async function createWorkItem(input: CreateWorkItemInput): Promise<WorkItem> {
  const res = await fetch("/api/work-items", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return handleResponse<WorkItem>(res);
}

export async function getWorkItem(id: number): Promise<WorkItem> {
  const res = await fetch(`/api/work-items/${id}`, { cache: "no-store" });
  return handleResponse<WorkItem>(res);
}

export async function updateWorkItem(id: number, input: UpdateWorkItemInput): Promise<WorkItem> {
  const res = await fetch(`/api/work-items/${id}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return handleResponse<WorkItem>(res);
}

export async function deleteWorkItem(id: number): Promise<void> {
  const res = await fetch(`/api/work-items/${id}`, { method: "DELETE" });
  if (!res.ok && res.status !== 204) {
    const body = await res.text();
    throw new Error(`el-storko API error ${res.status}: ${body}`);
  }
}
