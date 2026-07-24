import { For, Show, createSignal, type Component } from 'solid-js';
import type { WorkItem } from './api';
import {
  PRIORITY_LABELS,
  priorityLabel,
  priorityClass,
  toLabels,
  fromLabels,
  dateToUnix,
  unixToDate,
} from './format';

const fieldClass =
  'px-3 py-2 rounded-md border border-input bg-background text-sm ' +
  'focus:outline-none focus-visible:ring-2 focus-visible:ring-ring';

export const WorkItemCard: Component<{
  item: WorkItem;
  onUpdate: (item: WorkItem) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
}> = (props) => {
  const [editing, setEditing] = createSignal(false);

  const [title, setTitle] = createSignal('');
  const [description, setDescription] = createSignal('');
  const [priority, setPriority] = createSignal(0);
  const [labelsText, setLabelsText] = createSignal('');
  const [dueDate, setDueDate] = createSignal('');
  const [storyPoints, setStoryPoints] = createSignal('');
  const [milestone, setMilestone] = createSignal(false);

  const startEdit = () => {
    const it = props.item;
    setTitle(it.title);
    setDescription(it.description_markdown ?? '');
    setPriority(it.priority ?? 0);
    setLabelsText(fromLabels(it.labels));
    setDueDate(unixToDate(it.due_date));
    setStoryPoints(it.story_points ? String(it.story_points) : '');
    setMilestone(it.is_milestone ?? false);
    setEditing(true);
  };

  const save = (e: Event) => {
    e.preventDefault();
    const t = title().trim();
    if (!t) return;
    const updated: WorkItem = {
      ...props.item,
      title: t,
      description_markdown: description(),
      priority: priority(),
      labels: toLabels(labelsText()),
      due_date: dateToUnix(dueDate()),
      story_points: storyPoints() ? Number(storyPoints()) : 0,
      is_milestone: milestone(),
    };
    void props.onUpdate(updated).then(() => setEditing(false));
  };

  return (
    <li class="rounded-lg border border-border bg-card text-card-foreground px-4 py-3">
      <Show
        when={editing()}
        fallback={
          <div class="flex items-start gap-3">
            <div class="flex-1 min-w-0 space-y-1">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="font-medium">{props.item.title}</span>
                <span class={`text-xs px-2 py-0.5 rounded-full ${priorityClass(props.item.priority)}`}>
                  {priorityLabel(props.item.priority)}
                </span>
                <Show when={props.item.is_milestone}>
                  <span class="text-xs text-primary" title="Milestone">
                    ★ Milestone
                  </span>
                </Show>
              </div>
              <Show when={props.item.description_markdown}>
                <p class="text-sm text-muted-foreground whitespace-pre-wrap">
                  {props.item.description_markdown}
                </p>
              </Show>
              <div class="flex items-center gap-2 flex-wrap text-xs text-muted-foreground">
                <For each={props.item.labels}>
                  {(label) => (
                    <span class="px-2 py-0.5 rounded-full bg-muted ring-1 ring-border">{label}</span>
                  )}
                </For>
                <Show when={props.item.due_date}>
                  <span>Due {unixToDate(props.item.due_date)}</span>
                </Show>
                <Show when={props.item.story_points}>
                  <span>{props.item.story_points} pts</span>
                </Show>
              </div>
            </div>
            <div class="flex items-center gap-1 shrink-0">
              <button
                type="button"
                onClick={startEdit}
                class="text-xs px-2 py-1 rounded-md border border-border hover:bg-accent hover:text-accent-foreground transition-colors"
              >
                Edit
              </button>
              <button
                type="button"
                onClick={() => void props.onDelete(props.item.id)}
                class="text-xs px-2 py-1 rounded-md text-destructive hover:bg-destructive hover:text-destructive-foreground transition-colors"
              >
                Delete
              </button>
            </div>
          </div>
        }
      >
        <form onSubmit={save} class="space-y-3">
          <input
            aria-label="Title"
            class={`w-full ${fieldClass}`}
            value={title()}
            onInput={(e) => setTitle(e.currentTarget.value)}
          />
          <textarea
            aria-label="Description"
            placeholder="Description (markdown)"
            rows={3}
            class={`w-full ${fieldClass}`}
            value={description()}
            onInput={(e) => setDescription(e.currentTarget.value)}
          />
          <div class="flex flex-wrap gap-3">
            <label class="flex flex-col gap-1 text-xs text-muted-foreground">
              Priority
              <select
                class={fieldClass}
                value={priority()}
                onChange={(e) => setPriority(Number(e.currentTarget.value))}
              >
                <For each={PRIORITY_LABELS}>{(label, i) => <option value={i()}>{label}</option>}</For>
              </select>
            </label>
            <label class="flex flex-col gap-1 text-xs text-muted-foreground">
              Due date
              <input
                type="date"
                class={fieldClass}
                value={dueDate()}
                onInput={(e) => setDueDate(e.currentTarget.value)}
              />
            </label>
            <label class="flex flex-col gap-1 text-xs text-muted-foreground">
              Story points
              <input
                type="number"
                min="0"
                step="0.5"
                class={`w-28 ${fieldClass}`}
                value={storyPoints()}
                onInput={(e) => setStoryPoints(e.currentTarget.value)}
              />
            </label>
          </div>
          <input
            aria-label="Labels"
            placeholder="Labels (comma separated)"
            class={`w-full ${fieldClass}`}
            value={labelsText()}
            onInput={(e) => setLabelsText(e.currentTarget.value)}
          />
          <label class="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={milestone()}
              onChange={(e) => setMilestone(e.currentTarget.checked)}
            />
            Milestone
          </label>
          <div class="flex items-center gap-2">
            <button
              type="submit"
              disabled={title().trim() === ''}
              class="px-3 py-1.5 rounded-md bg-primary text-primary-foreground text-sm font-medium hover:opacity-90 transition-opacity disabled:opacity-50"
            >
              Save
            </button>
            <button
              type="button"
              onClick={() => setEditing(false)}
              class="px-3 py-1.5 rounded-md border border-border text-sm hover:bg-accent hover:text-accent-foreground transition-colors"
            >
              Cancel
            </button>
          </div>
        </form>
      </Show>
    </li>
  );
};
