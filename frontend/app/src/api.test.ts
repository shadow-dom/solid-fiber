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
  it('unwraps data on success', async () => {
    stubFetch(200, {
      status: true,
      data: [{ id: '1', title: 'a', project_id: 'p', description_markdown: '' }],
      error: null,
    });
    const items = await listWorkItems('p');
    expect(items).toHaveLength(1);
    expect(items[0].title).toBe('a');
  });

  it('encodes the project id into the query', async () => {
    const fn = stubFetch(200, { status: true, data: [], error: null });
    await listWorkItems('a/b');
    expect(fn).toHaveBeenCalledWith('/api/work-items?project_id=a%2Fb', undefined);
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
