import { For, Show, Switch, Match, createResource, createSignal, type Component } from 'solid-js';
import {
  listWorkItems,
  createWorkItem,
  updateWorkItem,
  deleteWorkItem,
  type WorkItem,
} from './api';
import { PRIORITY_LABELS, toLabels, dateToUnix } from './format';
import { WorkItemCard } from './WorkItemCard';

const fieldClass =
  'px-3 py-2 rounded-md border border-input bg-background text-sm ' +
  'focus:outline-none focus-visible:ring-2 focus-visible:ring-ring';

export const WorkItems: Component = () => {
  const [projectId, setProjectId] = createSignal('demo');
  const [items, { refetch }] = createResource(projectId, listWorkItems);

  const [title, setTitle] = createSignal('');
  const [priority, setPriority] = createSignal(0);
  const [showMore, setShowMore] = createSignal(false);
  const [description, setDescription] = createSignal('');
  const [labelsText, setLabelsText] = createSignal('');
  const [dueDate, setDueDate] = createSignal('');
  const [storyPoints, setStoryPoints] = createSignal('');
  const [milestone, setMilestone] = createSignal(false);

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
    // Capture reactive values before the async closure.
    const input: WorkItem = {
      id: '',
      title: t,
      project_id: projectId(),
      description_markdown: description(),
      priority: priority(),
      labels: toLabels(labelsText()),
      due_date: dateToUnix(dueDate()),
      story_points: storyPoints() ? Number(storyPoints()) : 0,
      is_milestone: milestone(),
    };
    void run(async () => {
      await createWorkItem(input);
      setTitle('');
      setPriority(0);
      setDescription('');
      setLabelsText('');
      setDueDate('');
      setStoryPoints('');
      setMilestone(false);
      setShowMore(false);
    });
  };

  const onUpdate = (item: WorkItem) => run(() => updateWorkItem(item.id, item));
  const onDelete = (id: string) => run(() => deleteWorkItem(id));

  return (
    <div class="space-y-6">
      <div class="flex items-center gap-2">
        <label class="text-sm text-muted-foreground" for="project">Project</label>
        <input
          id="project"
          value={projectId()}
          onInput={(e) => setProjectId(e.currentTarget.value)}
          class={`w-40 ${fieldClass}`}
        />
      </div>

      <form onSubmit={add} class="space-y-3">
        <div class="flex flex-wrap items-center gap-2">
          <input
            placeholder="New work item title…"
            value={title()}
            onInput={(e) => setTitle(e.currentTarget.value)}
            class={`flex-1 min-w-60 ${fieldClass}`}
          />
          <select
            aria-label="Priority"
            value={priority()}
            onChange={(e) => setPriority(Number(e.currentTarget.value))}
            class={fieldClass}
          >
            <For each={PRIORITY_LABELS}>{(label, i) => <option value={i()}>{label}</option>}</For>
          </select>
          <button
            type="button"
            onClick={() => setShowMore(!showMore())}
            aria-expanded={showMore()}
            class="px-3 py-2 rounded-md border border-border text-sm hover:bg-accent hover:text-accent-foreground transition-colors"
          >
            {showMore() ? 'Fewer' : 'More'}
          </button>
          <button
            type="submit"
            disabled={busy() || title().trim() === ''}
            class="px-4 py-2 rounded-md bg-primary text-primary-foreground font-medium hover:opacity-90 transition-opacity disabled:opacity-50"
          >
            Add
          </button>
        </div>

        <Show when={showMore()}>
          <div class="space-y-3 rounded-lg border border-border bg-muted/40 p-4">
            <textarea
              aria-label="Description"
              placeholder="Description (markdown)"
              rows={2}
              value={description()}
              onInput={(e) => setDescription(e.currentTarget.value)}
              class={`w-full ${fieldClass}`}
            />
            <input
              aria-label="Labels"
              placeholder="Labels (comma separated)"
              value={labelsText()}
              onInput={(e) => setLabelsText(e.currentTarget.value)}
              class={`w-full ${fieldClass}`}
            />
            <div class="flex flex-wrap items-end gap-3">
              <label class="flex flex-col gap-1 text-xs text-muted-foreground">
                Due date
                <input
                  type="date"
                  value={dueDate()}
                  onInput={(e) => setDueDate(e.currentTarget.value)}
                  class={fieldClass}
                />
              </label>
              <label class="flex flex-col gap-1 text-xs text-muted-foreground">
                Story points
                <input
                  type="number"
                  min="0"
                  step="0.5"
                  value={storyPoints()}
                  onInput={(e) => setStoryPoints(e.currentTarget.value)}
                  class={`w-28 ${fieldClass}`}
                />
              </label>
              <label class="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={milestone()}
                  onChange={(e) => setMilestone(e.currentTarget.checked)}
                />
                Milestone
              </label>
            </div>
          </div>
        </Show>
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
                {(item) => <WorkItemCard item={item} onUpdate={onUpdate} onDelete={onDelete} />}
              </For>
            </ul>
          </Show>
        </Match>
      </Switch>
    </div>
  );
};
