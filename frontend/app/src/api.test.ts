import { describe, it, expect, vi, afterEach } from 'vitest';
import { listWorkItems, createWorkItem, deleteWorkItem } from './api';

function stubFetch(status: number, body?: unknown) {
  const fn = vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  });
  vi.stubGlobal('fetch', fn);
  return fn;
}

afterEach(() => vi.unstubAllGlobals());

describe('api client', () => {
  it('returns a page with items and total from meta', async () => {
    stubFetch(200, {
      status: true,
      data: [{ id: '1', title: 'a', project_id: 'p', description_markdown: '' }],
      meta: { total: 5, limit: 20, offset: 0 },
      error: null,
    });
    const page = await listWorkItems('p');
    expect(page.items).toHaveLength(1);
    expect(page.items[0].title).toBe('a');
    expect(page.total).toBe(5);
  });

  it('encodes the project id and passes limit/offset', async () => {
    const fn = stubFetch(200, { status: true, data: [], meta: { total: 0, limit: 20, offset: 40 }, error: null });
    await listWorkItems('a/b', { limit: 20, offset: 40 });
    expect(fn).toHaveBeenCalledWith('/api/work-items?project_id=a%2Fb&limit=20&offset=40');
  });

  it('throws with the server error message on an error envelope', async () => {
    stubFetch(400, { status: false, data: null, error: 'title is required' });
    await expect(createWorkItem({ title: '', project_id: 'p' })).rejects.toThrow('title is required');
  });

  it('resolves without parsing a body on 204', async () => {
    const fn = stubFetch(204);
    await expect(deleteWorkItem('1')).resolves.toBeUndefined();
    expect(fn).toHaveBeenCalledWith('/api/work-items/1', { method: 'DELETE' });
  });
});
