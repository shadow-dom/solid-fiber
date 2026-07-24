import { describe, it, expect } from 'vitest';
import { toLabels, fromLabels, dateToUnix, unixToDate, priorityLabel, priorityClass } from './format';

describe('format helpers', () => {
  it('toLabels trims and drops empty entries', () => {
    expect(toLabels(' a, b ,, c ')).toEqual(['a', 'b', 'c']);
    expect(toLabels('')).toEqual([]);
  });

  it('fromLabels joins with commas', () => {
    expect(fromLabels(['a', 'b'])).toBe('a, b');
    expect(fromLabels(undefined)).toBe('');
  });

  it('date <-> unix round-trips at UTC midnight', () => {
    const u = dateToUnix('2026-07-24');
    expect(u).toBe(Date.parse('2026-07-24T00:00:00Z') / 1000);
    expect(unixToDate(u)).toBe('2026-07-24');
    expect(dateToUnix('')).toBe(0);
    expect(unixToDate(0)).toBe('');
    expect(unixToDate(undefined)).toBe('');
  });

  it('priorityLabel maps levels with a default', () => {
    expect(priorityLabel(3)).toBe('High');
    expect(priorityLabel(0)).toBe('None');
    expect(priorityLabel(undefined)).toBe('None');
  });

  it('priorityClass returns a class string per level', () => {
    expect(priorityClass(3)).toContain('destructive');
    expect(priorityClass(undefined)).toContain('muted');
  });
});
