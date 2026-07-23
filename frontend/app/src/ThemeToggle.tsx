import { For, Show, createSignal, onCleanup, onMount, type Component } from 'solid-js';
import { Dynamic } from 'solid-js/web';
import { MODES, PALETTES, useTheme, type Mode, type Palette } from './theme';

const SunIcon: Component = () => (
  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-4 h-4">
    <circle cx="12" cy="12" r="4" />
    <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41" />
  </svg>
);
const MoonIcon: Component = () => (
  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-4 h-4">
    <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
  </svg>
);
const SystemIcon: Component = () => (
  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-4 h-4">
    <rect x="2" y="3" width="20" height="14" rx="2" />
    <path d="M8 21h8M12 17v4" />
  </svg>
);
const CheckIcon: Component = () => (
  <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
    <polyline points="20 6 9 17 4 12" />
  </svg>
);

const MODE_META: Record<Mode, { label: string; Icon: Component }> = {
  light: { label: 'Light', Icon: SunIcon },
  dark: { label: 'Dark', Icon: MoonIcon },
  system: { label: 'System', Icon: SystemIcon },
};

const PALETTE_LABEL: Record<Palette, string> = {
  default: 'Default',
  rose: 'Rose',
  emerald: 'Emerald',
  amber: 'Amber',
  violet: 'Violet',
};

// Preview swatch for each palette — uses HSL values that match theme.css.
const PALETTE_SWATCH: Record<Palette, string> = {
  default: 'hsl(221.2 83.2% 53.3%)',
  rose: 'hsl(346.8 77.2% 49.8%)',
  emerald: 'hsl(142.1 76.2% 36.3%)',
  amber: 'hsl(35.5 91.7% 54%)',
  violet: 'hsl(262.1 83.3% 57.8%)',
};

export const ThemeToggle: Component = () => {
  const { mode, palette, resolved, setMode, setPalette } = useTheme();
  const [open, setOpen] = createSignal(false);
  let rootEl: HTMLDivElement | undefined;

  onMount(() => {
    const onClick = (e: MouseEvent) => {
      if (rootEl && !rootEl.contains(e.target as Node)) setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpen(false);
    };
    document.addEventListener('mousedown', onClick);
    document.addEventListener('keydown', onKey);
    onCleanup(() => {
      document.removeEventListener('mousedown', onClick);
      document.removeEventListener('keydown', onKey);
    });
  });

  return (
    <div ref={rootEl} class="relative inline-block">
      <button
        type="button"
        aria-label="Theme settings"
        aria-haspopup="menu"
        aria-expanded={open()}
        onClick={() => setOpen(!open())}
        class="inline-flex items-center justify-center w-9 h-9 rounded-md
               bg-card text-card-foreground border border-border
               hover:bg-accent hover:text-accent-foreground
               shadow-sm transition-colors
               focus:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
      >
        <span class="relative w-4 h-4 block">
          <span
            class="absolute inset-0 transition-all duration-300"
            classList={{
              'opacity-100 rotate-0 scale-100': resolved() === 'light',
              'opacity-0 -rotate-90 scale-50': resolved() !== 'light',
            }}
          >
            <SunIcon />
          </span>
          <span
            class="absolute inset-0 transition-all duration-300"
            classList={{
              'opacity-100 rotate-0 scale-100': resolved() === 'dark',
              'opacity-0 rotate-90 scale-50': resolved() !== 'dark',
            }}
          >
            <MoonIcon />
          </span>
        </span>
      </button>

      <Show when={open()}>
        <div
          role="menu"
          class="absolute right-0 mt-2 w-56 origin-top-right z-50
                 rounded-lg border border-border bg-popover text-popover-foreground
                 shadow-lg shadow-black/10
                 p-2"
        >
          <div class="px-2 pt-1 pb-2 text-xs font-medium text-muted-foreground uppercase tracking-wide">
            Mode
          </div>
          <div class="grid grid-cols-3 gap-1 mb-2">
            <For each={MODES}>
              {(m) => {
                const meta = MODE_META[m];
                const selected = () => mode() === m;
                return (
                  <button
                    type="button"
                    aria-label={meta.label}
                    aria-pressed={selected()}
                    onClick={() => setMode(m)}
                    class="flex flex-col items-center gap-1 py-2 rounded-md text-xs
                           border border-transparent
                           hover:bg-accent hover:text-accent-foreground
                           transition-colors"
                    classList={{
                      'bg-accent text-accent-foreground border-border': selected(),
                    }}
                  >
                    <Dynamic component={meta.Icon} />
                    <span>{meta.label}</span>
                  </button>
                );
              }}
            </For>
          </div>

          <div class="h-px bg-border my-1" />

          <div class="px-2 pt-1 pb-2 text-xs font-medium text-muted-foreground uppercase tracking-wide">
            Palette
          </div>
          <div class="flex flex-col gap-0.5">
            <For each={PALETTES}>
              {(p) => {
                const selected = () => palette() === p;
                return (
                  <button
                    type="button"
                    role="menuitemradio"
                    aria-checked={selected()}
                    onClick={() => setPalette(p)}
                    class="w-full flex items-center gap-2 px-2 py-1.5 rounded-md text-sm
                           hover:bg-accent hover:text-accent-foreground transition-colors"
                  >
                    <span
                      class="w-4 h-4 rounded-full ring-1 ring-border"
                      style={{ background: PALETTE_SWATCH[p] }}
                      aria-hidden="true"
                    />
                    <span>{PALETTE_LABEL[p]}</span>
                    <Show when={selected()}>
                      <span class="ml-auto text-primary">
                        <CheckIcon />
                      </span>
                    </Show>
                  </button>
                );
              }}
            </For>
          </div>
        </div>
      </Show>
    </div>
  );
};
