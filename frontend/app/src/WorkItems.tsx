import { For, Show, Switch, Match, createResource, createSignal, type Component } from 'solid-js';
import {
  listWorkItems,
  createWorkItem,
  updateWorkItem,
  deleteWorkItem,
  type WorkItem,
} from './api';

const PRIORITY_LABELS = ['None', 'Low', 'Medium', 'High'];

const priorityLabel = (p?: number) => PRIORITY_LABELS[p ?? 0] ?? 'None';

export const WorkItems: Component = () => {
  const [projectId, setProjectId] = createSignal('demo');
  const [items, { refetch }] = createResource(projectId, listWorkItems);

  const [title, setTitle] = createSignal('');
  const [priority, setPriority] = createSignal(0);
  const [busy, setBusy] = createSignal(false);
  const [error, setError] = createSignal<string | null>(null);

  const run = async (fn: () => Promise<unknown>) => {
    setBusy(true);
    setError(null);
    try {
      await fn();
      await refetch();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  };

  const add = (e: Event) => {
    e.preventDefault();
    const t = title().trim();
    if (!t) return;
    // Capture reactive values in the event handler; the async closure uses them.
    const project = projectId();
    const p = priority();
    void run(async () => {
      await createWorkItem({ title: t, project_id: project, priority: p });
      setTitle('');
      setPriority(0);
    });
  };

  const cyclePriority = (item: WorkItem) =>
    run(() => updateWorkItem(item.id, { ...item, priority: ((item.priority ?? 0) + 1) % 4 }));

  const rename = (item: WorkItem, next: string) => {
    const t = next.trim();
    if (!t || t === item.title) return;
    void run(() => updateWorkItem(item.id, { ...item, title: t }));
  };

  const remove = (item: WorkItem) => run(() => deleteWorkItem(item.id));

  return (
    <div class="space-y-6">
      <div class="flex items-center gap-2">
        <label class="text-sm text-muted-foreground" for="project">Project</label>
        <input
          id="project"
          value={projectId()}
          onInput={(e) => setProjectId(e.currentTarget.value)}
          class="px-2 py-1 rounded-md border border-input bg-background text-sm w-40
                 focus:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        />
      </div>

      <form onSubmit={add} class="flex flex-wrap items-center gap-2">
        <input
          placeholder="New work item title…"
          value={title()}
          onInput={(e) => setTitle(e.currentTarget.value)}
          class="flex-1 min-w-60 px-3 py-2 rounded-md border border-input bg-background
                 focus:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        />
        <select
          value={priority()}
          onChange={(e) => setPriority(Number(e.currentTarget.value))}
          class="px-2 py-2 rounded-md border border-input bg-background text-sm"
        >
          <For each={PRIORITY_LABELS}>
            {(label, i) => <option value={i()}>{label}</option>}
          </For>
        </select>
        <button
          type="submit"
          disabled={busy() || title().trim() === ''}
          class="px-4 py-2 rounded-md bg-primary text-primary-foreground font-medium
                 hover:opacity-90 transition-opacity disabled:opacity-50"
        >
          Add
        </button>
      </form>

      <Show when={error()}>
        <p class="text-sm text-destructive">{error()}</p>
      </Show>

      <Switch>
        <Match when={items.loading}>
          <p class="text-sm text-muted-foreground">Loading…</p>
        </Match>
        <Match when={items.error}>
          <p class="text-sm text-destructive">Failed to load: {String(items.error)}</p>
        </Match>
        <Match when={items()}>
          <Show
            when={(items() ?? []).length > 0}
            fallback={<p class="text-sm text-muted-foreground">No work items yet. Add one above.</p>}
          >
            <ul class="space-y-2">
              <For each={items()}>
                {(item) => (
                  <li class="flex items-center gap-3 rounded-lg border border-border bg-card text-card-foreground px-4 py-3">
                    <input
                      class="flex-1 bg-transparent focus:outline-none focus-visible:ring-1 focus-visible:ring-ring rounded px-1"
                      value={item.title}
                      onChange={(e) => rename(item, e.currentTarget.value)}
                    />
                    <button
                      type="button"
                      onClick={() => void cyclePriority(item)}
                      title="Click to change priority"
                      class="text-xs px-2 py-1 rounded-full border border-border bg-muted text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors"
                    >
                      {priorityLabel(item.priority)}
                    </button>
                    <button
                      type="button"
                      onClick={() => void remove(item)}
                      class="text-xs px-2 py-1 rounded-md text-destructive hover:bg-destructive hover:text-destructive-foreground transition-colors"
                    >
                      Delete
                    </button>
                  </li>
                )}
              </For>
            </ul>
          </Show>
        </Match>
      </Switch>
    </div>
  );
};
