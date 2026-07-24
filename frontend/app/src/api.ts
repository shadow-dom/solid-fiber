// Typed client for the work-items API. In dev, Vite proxies /api to the Fiber
// backend on :3000; in production the SPA is served from the same origin.

export interface WorkItem {
  id: string;
  title: string;
  description_markdown: string;
  priority?: number;
  labels?: string[];
  is_milestone?: boolean;
  project_id: string;
}

export type NewWorkItem = Pick<WorkItem, 'title' | 'project_id'> & Partial<WorkItem>;

interface Envelope<T> {
  status: boolean;
  data: T;
  error: string | null;
}

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(url, options);
  // DELETE returns 204 with no body.
  if (res.status === 204) return undefined as T;

  const body = (await res.json()) as Envelope<T>;
  if (!res.ok || !body.status) {
    throw new Error(body?.error || `Request failed (${res.status})`);
  }
  return body.data;
}

const json = (method: string, payload: unknown): RequestInit => ({
  method,
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(payload),
});

export const listWorkItems = (projectId: string): Promise<WorkItem[]> =>
  request<WorkItem[]>(`/api/work-items?project_id=${encodeURIComponent(projectId)}`);

export const createWorkItem = (input: NewWorkItem): Promise<WorkItem> =>
  request<WorkItem>('/api/work-items', json('POST', input));

export const updateWorkItem = (id: string, input: WorkItem): Promise<WorkItem> =>
  request<WorkItem>(`/api/work-items/${encodeURIComponent(id)}`, json('PUT', input));

export const deleteWorkItem = (id: string): Promise<void> =>
  request<void>(`/api/work-items/${encodeURIComponent(id)}`, { method: 'DELETE' });
