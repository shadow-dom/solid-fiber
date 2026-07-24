// Pure helpers shared by the work-item create form and edit card.

export const PRIORITY_LABELS = ['None', 'Low', 'Medium', 'High'];

export const priorityLabel = (p?: number): string => PRIORITY_LABELS[p ?? 0] ?? 'None';

export const priorityClass = (p?: number): string => {
  switch (p) {
    case 3:
      return 'bg-destructive text-destructive-foreground';
    case 2:
      return 'bg-primary text-primary-foreground';
    case 1:
      return 'bg-secondary text-secondary-foreground';
    default:
      return 'bg-muted text-muted-foreground';
  }
};

/** Parse a comma-separated string into trimmed, non-empty labels. */
export const toLabels = (text: string): string[] =>
  text
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean);

/** Render a labels array back to a comma-separated string. */
export const fromLabels = (labels?: string[]): string => (labels ?? []).join(', ');

/** Convert a `YYYY-MM-DD` string to a UTC unix timestamp (seconds); 0 when empty. */
export const dateToUnix = (s: string): number =>
  s ? Math.floor(Date.parse(`${s}T00:00:00Z`) / 1000) : 0;

/** Convert a unix timestamp (seconds) to a `YYYY-MM-DD` string; '' when unset. */
export const unixToDate = (n?: number): string =>
  n ? new Date(n * 1000).toISOString().slice(0, 10) : '';
