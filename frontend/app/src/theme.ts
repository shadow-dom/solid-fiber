import { createEffect, createSignal, onCleanup, onMount } from 'solid-js';

export type Mode = 'light' | 'dark' | 'system';
export type Palette = 'default' | 'rose' | 'emerald' | 'amber' | 'violet';

export const PALETTES: Palette[] = ['default', 'rose', 'emerald', 'amber', 'violet'];
export const MODES: Mode[] = ['light', 'dark', 'system'];

const KEY_MODE = 'theme-mode';
const KEY_PALETTE = 'theme-palette';

const media = () => window.matchMedia('(prefers-color-scheme: dark)');

const readMode = (): Mode => {
  const v = localStorage.getItem(KEY_MODE);
  return MODES.includes(v as Mode) ? (v as Mode) : 'system';
};
const readPalette = (): Palette => {
  const v = localStorage.getItem(KEY_PALETTE);
  return PALETTES.includes(v as Palette) ? (v as Palette) : 'default';
};

const resolveMode = (m: Mode): 'light' | 'dark' =>
  m === 'system' ? (media().matches ? 'dark' : 'light') : m;

const apply = (resolved: 'light' | 'dark', palette: Palette) => {
  const root = document.documentElement;
  root.classList.toggle('dark', resolved === 'dark');
  if (palette === 'default') root.removeAttribute('data-theme');
  else root.setAttribute('data-theme', palette);
};

const [mode, setMode] = createSignal<Mode>('system');
const [palette, setPalette] = createSignal<Palette>('default');
const [resolved, setResolved] = createSignal<'light' | 'dark'>('light');

export const useTheme = () => {
  onMount(() => {
    setMode(readMode());
    setPalette(readPalette());

    const mq = media();
    const onChange = () => {
      if (mode() === 'system') {
        const r = resolveMode('system');
        setResolved(r);
        apply(r, palette());
      }
    };
    mq.addEventListener('change', onChange);
    onCleanup(() => mq.removeEventListener('change', onChange));
  });

  createEffect(() => {
    const m = mode();
    const p = palette();
    localStorage.setItem(KEY_MODE, m);
    localStorage.setItem(KEY_PALETTE, p);
    const r = resolveMode(m);
    setResolved(r);
    apply(r, p);
  });

  return { mode, palette, resolved, setMode, setPalette };
};
