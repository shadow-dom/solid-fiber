import type { Component } from 'solid-js';
import { ThemeToggle } from './ThemeToggle';
import { WorkItems } from './WorkItems';

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

      <main class="max-w-3xl mx-auto px-6 py-12 space-y-8">
        <section class="space-y-2">
          <h1 class="text-3xl font-bold tracking-tight">
            Work <span class="text-primary">items</span>
          </h1>
          <p class="text-muted-foreground">
            Backed by the Fiber API and a SQLite store. Create, rename, reprioritize, and delete —
            changes persist across restarts.
          </p>
        </section>

        <WorkItems />
      </main>
    </div>
  );
};

export default App;
