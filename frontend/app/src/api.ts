// Typed client for the work-items API. In dev, Vite proxies /api to the Fiber
// backend on :3000; in production the SPA is served from the same origin.

export interface WorkItem {
  id: string;
  title: string;
  description_markdown: string;
  parent_id?: string;
  column_id?: string;
  assignee_id?: string;
  reporter_id?: string;
  sprint_id?: string;
  priority?: number;
  estimate_hours?: number;
  story_points?: number;
  due_date?: number;
  is_milestone?: boolean;
  epic_color?: string;
  labels?: string[];
  project_id: string;
}

export type NewWorkItem = Pick<WorkItem, 'title' | 'project_id'> & Partial<WorkItem>;

interface Envelope<T> {
  status: boolean;
  data: T;
  error: string | null;
}

/** A page of results plus pagination metadata. */
export interface Page<T> {
  items: T[];
  total: number;
  limit: number;
  offset: number;
}

interface ListEnvelope {
  status: boolean;
  data: WorkItem[];
  meta?: { total: number; limit: number; offset: number };
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

export const listWorkItems = async (
  projectId: string,
  opts: { limit?: number; offset?: number } = {},
): Promise<Page<WorkItem>> => {
  const params = new URLSearchParams({ project_id: projectId });
  if (opts.limit != null) params.set('limit', String(opts.limit));
  if (opts.offset != null) params.set('offset', String(opts.offset));

  const res = await fetch(`/api/work-items?${params.toString()}`);
  const body = (await res.json()) as ListEnvelope;
  if (!res.ok || !body.status) {
    throw new Error(body?.error || `Request failed (${res.status})`);
  }
  const items = body.data ?? [];
  return {
    items,
    total: body.meta?.total ?? items.length,
    limit: body.meta?.limit ?? opts.limit ?? items.length,
    offset: body.meta?.offset ?? opts.offset ?? 0,
  };
};

export const createWorkItem = (input: NewWorkItem): Promise<WorkItem> =>
  request<WorkItem>('/api/work-items', json('POST', input));

export const updateWorkItem = (id: string, input: WorkItem): Promise<WorkItem> =>
  request<WorkItem>(`/api/work-items/${encodeURIComponent(id)}`, json('PUT', input));

export const deleteWorkItem = (id: string): Promise<void> =>
  request<void>(`/api/work-items/${encodeURIComponent(id)}`, { method: 'DELETE' });
