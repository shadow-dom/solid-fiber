# Adding a frontend feature

The reference is `src/WorkItems.tsx` + `src/WorkItemCard.tsx` + `src/api.ts`.
SolidJS is fine-grained reactive — read signals as function calls (`count()`),
never destructure props, and do side-effecting work in event handlers or
`createEffect`. Styling uses UnoCSS semantic tokens (`bg-card`, `text-muted-foreground`,
`border-border`, `bg-primary`, …) that are theme-aware — reuse them, don't hard-code colors.

## 1. Typed API client — extend `src/api.ts`

Every response is the envelope `{status, data, error}` (+ `meta` for lists). The
existing `request<T>()` helper unwraps `data`, throws on `!status`, and handles the
`204` on delete. Reuse it; only lists need custom handling for `meta`.

```ts
export interface <Resource> {
  id: string;
  name: string;
  project_id: string;
}
export type New<Resource> = Pick<<Resource>, 'name' | 'project_id'> & Partial<<Resource>>;

export const list<Resource>s = async (
  projectId: string,
  opts: { limit?: number; offset?: number } = {},
): Promise<Page<<Resource>>> => {
  const params = new URLSearchParams({ project_id: projectId });
  if (opts.limit != null) params.set('limit', String(opts.limit));
  if (opts.offset != null) params.set('offset', String(opts.offset));
  const res = await fetch(`/api/<resource>s?${params}`);
  const body = (await res.json()) as { status: boolean; data: <Resource>[]; meta?: { total: number; limit: number; offset: number }; error: string | null };
  if (!res.ok || !body.status) throw new Error(body?.error || `Request failed (${res.status})`);
  const items = body.data ?? [];
  return { items, total: body.meta?.total ?? items.length, limit: body.meta?.limit ?? items.length, offset: body.meta?.offset ?? 0 };
};

export const create<Resource> = (input: New<Resource>): Promise<<Resource>> =>
  request<<Resource>>('/api/<resource>s', json('POST', input));
export const update<Resource> = (id: string, input: <Resource>): Promise<<Resource>> =>
  request<<Resource>>(`/api/<resource>s/${encodeURIComponent(id)}`, json('PUT', input));
export const delete<Resource> = (id: string): Promise<void> =>
  request<void>(`/api/<resource>s/${encodeURIComponent(id)}`, { method: 'DELETE' });
```

## 2. Component — `src/<Resource>s.tsx`

Fetch with `createResource` keyed on a source function so it refetches when inputs
change; do mutations through a small `run()` wrapper that toggles busy/error and
refetches. Copy the shape of `WorkItems.tsx`.

```tsx
export const <Resource>s: Component = () => {
  const [projectId, setProjectId] = createSignal('demo');
  const [page, { refetch }] = createResource(
    () => ({ projectId: projectId() }),
    (src) => list<Resource>s(src.projectId),
  );
  const items = () => page()?.items ?? [];

  const [error, setError] = createSignal<string | null>(null);
  const run = async (fn: () => Promise<unknown>) => {
    setError(null);
    try { await fn(); await refetch(); }
    catch (e) { setError(e instanceof Error ? e.message : String(e)); }
  };

  // Loading/error/empty handled with <Switch>/<Match>/<Show> as in WorkItems.tsx.
};
```

### ESLint rules that will fail the build if violated

`bun run lint` runs at `--max-warnings 0`, so these are hard failures:

- **`solid/reactivity`** — do not read signals inside an async closure that runs
  later. Capture the values first, in the synchronous part of the event handler:
  ```tsx
  const add = (e: Event) => {
    e.preventDefault();
    const name = title().trim();
    const project = projectId();            // capture BEFORE the async closure
    void run(() => create<Resource>({ name, project_id: project }));
  };
  ```
- **`no-unassigned-vars`** — the plugin doesn't understand Solid's `ref={el}`
  compiler assignment. Use a callback ref: `ref={(el) => (rootEl = el)}`.

## 3. Render it

Add `<<Resource>s />` to `App.tsx` (keep the header + `<ThemeToggle />`). If the
app grows past one view, that's the moment to add a router.

## 4. Tests — `src/<Resource>s.test.tsx`

Vitest + `@solidjs/testing-library`, stubbing `fetch`. Cover: renders items,
create flow (assert the `POST` body), and the error state. Pure helpers go in
`format.ts` with their own `*.test.ts`. Pattern from `WorkItems.test.tsx`:

```tsx
const envelope = (data: unknown) => ({ ok: true, status: 200, json: async () => ({ status: true, data, error: null }) });
afterEach(() => vi.unstubAllGlobals());

it('renders fetched items', async () => {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(envelope([{ id: '1', name: 'First', project_id: 'demo' }])));
  render(() => <<Resource>s />);
  expect(await screen.findByText('First')).toBeInTheDocument();
});
```

Assertion tips learned the hard way: `getByText('Low')`-style queries can match a
`<select>` `<option>` *and* a badge — prefer `getAllByText(...).length` or a more
specific query. When a value renders inside an `<input>`, use `findByDisplayValue`.

## Then verify

Run the frontend gates (SKILL.md → Verification gates) — and remember to
`touch backend/web/dist/.gitkeep` after `bun run build`, or the embed placeholder
is gone and the Go build breaks in a fresh clone.
