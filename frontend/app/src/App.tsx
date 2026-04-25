import type { Component } from 'solid-js';
import { ThemeToggle } from './ThemeToggle';

const App: Component = () => {
  return (
    <div class="min-h-screen bg-background text-foreground">
      <header class="flex items-center justify-between px-6 py-4 border-b border-border">
        <div class="flex items-center gap-2">
          <div class="w-7 h-7 rounded-md bg-primary" />
          <span class="font-semibold tracking-tight">solid-fiber</span>
        </div>
        <ThemeToggle />
      </header>

      <main class="max-w-3xl mx-auto px-6 py-20 space-y-12">
        <section class="text-center space-y-4">
          <h1 class="text-5xl font-bold tracking-tight">
            A modern <span class="text-primary">themed</span> stack
          </h1>
          <p class="text-muted-foreground text-lg">
            Solid + UnoCSS preset-wind4 + Fiber. Fully tokenized via CSS variables.
          </p>
        </section>

        <section class="grid sm:grid-cols-2 gap-4">
          <div class="rounded-lg border border-border bg-card text-card-foreground p-6">
            <h3 class="font-semibold mb-2">Card</h3>
            <p class="text-sm text-muted-foreground">
              Surfaces use <code>bg-card</code> + <code>text-card-foreground</code>.
            </p>
          </div>
          <div class="rounded-lg border border-border bg-muted text-muted-foreground p-6">
            <h3 class="font-semibold mb-2 text-foreground">Muted</h3>
            <p class="text-sm">
              Subtle blocks use <code>bg-muted</code>.
            </p>
          </div>
        </section>

        <section class="flex flex-wrap gap-3 justify-center">
          <button class="px-4 py-2 rounded-md bg-primary text-primary-foreground hover:opacity-90 transition-opacity font-medium">
            Primary
          </button>
          <button class="px-4 py-2 rounded-md bg-secondary text-secondary-foreground hover:bg-accent transition-colors font-medium">
            Secondary
          </button>
          <button class="px-4 py-2 rounded-md border border-border bg-background text-foreground hover:bg-accent hover:text-accent-foreground transition-colors font-medium">
            Outline
          </button>
          <button class="px-4 py-2 rounded-md bg-destructive text-destructive-foreground hover:opacity-90 transition-opacity font-medium">
            Destructive
          </button>
        </section>
      </main>
    </div>
  );
};

export default App;
