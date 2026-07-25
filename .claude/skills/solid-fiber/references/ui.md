# UI: the shadcn/ui design language (for Solid)

shadcn/ui is React, but it isn't a dependency you install — it's a *design system*
you own: Tailwind semantic tokens + accessible Radix primitives + a small set of
composable, variant-driven components you copy into your app. This project follows
that philosophy, translated to Solid.

## What's already here

`frontend/app/src/styles/theme.css` defines shadcn/ui's exact token set — `--background`,
`--foreground`, `--card`, `--popover`, `--primary`, `--secondary`, `--muted`,
`--accent`, `--destructive`, `--border`, `--input`, `--ring`, `--radius` — wired for
light/dark plus named palettes, and mapped to UnoCSS utilities in `unocss.config.ts`
(`bg-background`, `text-muted-foreground`, `border-border`, `ring-ring`, …).

**Always style through these tokens. Never hard-code a color** (`bg-blue-500`, `#0a0a0b`).
That's what keeps every surface theme-aware and consistent — the whole point of the token system.

## The rule: compose primitives, don't inline

shadcn/ui's core practice is a `src/components/ui/` library of small, reusable,
**variant-driven** components with a consistent API (`variant`, `size`), built on
accessible primitives. Feature code composes those — it does not repeat long class
strings inline.

Build these primitives as you need them (start with what the feature uses), mirroring
shadcn/ui's component names and prop shapes so they're familiar: `Button`, `Input`,
`Textarea`, `Label`, `Card` (+ `CardHeader`/`CardContent`/`CardFooter`), `Badge`,
`Select`, `Dialog`, `DropdownMenu`. A feature then reads:

```tsx
<Button variant="destructive" size="sm" onClick={() => remove(item.id)}>Delete</Button>
```

### The `cn()` helper

One utility, `src/lib/cn.ts`, merges class names (shadcn/ui uses `clsx` + `tailwind-merge`;
under UnoCSS, `clsx` alone is enough):

```ts
import clsx, { type ClassValue } from 'clsx';
export const cn = (...classes: ClassValue[]) => clsx(classes);
```

### A primitive, shadcn/ui-style

Variants live in one place; the component stays tiny and self-explanatory:

```tsx
// src/components/ui/Button.tsx
import { splitProps, type ComponentProps } from 'solid-js';
import { cn } from '../../lib/cn';

const variants = {
  default: 'bg-primary text-primary-foreground hover:opacity-90',
  secondary: 'bg-secondary text-secondary-foreground hover:bg-accent',
  outline: 'border border-border bg-background hover:bg-accent hover:text-accent-foreground',
  ghost: 'hover:bg-accent hover:text-accent-foreground',
  destructive: 'bg-destructive text-destructive-foreground hover:opacity-90',
} as const;
const sizes = { sm: 'h-8 px-3 text-sm', default: 'h-9 px-4', lg: 'h-10 px-6', icon: 'h-9 w-9' } as const;

export function Button(props: ComponentProps<'button'> & { variant?: keyof typeof variants; size?: keyof typeof sizes }) {
  const [local, rest] = splitProps(props, ['variant', 'size', 'class']);
  return (
    <button
      class={cn(
        'inline-flex items-center justify-center gap-2 rounded-md font-medium transition-colors',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background',
        'disabled:opacity-50 disabled:pointer-events-none',
        variants[local.variant ?? 'default'],
        sizes[local.size ?? 'default'],
        local.class,
      )}
      {...rest}
    />
  );
}
```

## Accessibility is part of the design, not an add-on

shadcn/ui is accessible by default because it builds on Radix. The Solid equivalent is
**Kobalte (`@kobalte/core`)** — use it for anything interactive/stateful (dropdown menu,
dialog, select, popover, tooltip, checkbox, tabs). It gives you focus management, keyboard
navigation, and ARIA for free; you style its parts with the token utilities. `ThemeToggle.tsx`
currently hand-rolls a menu — new interactive components should prefer Kobalte rather than
re-implementing focus/keyboard handling.

Baseline expectations for every component: keyboard operable, visible `focus-visible` ring
(the `ring` token is already set up), labelled inputs (`<Label for>` / `aria-label`), and
correct roles. These aren't extra — a component that isn't accessible isn't done.

## Migrating existing inline UI

`WorkItems.tsx` / `WorkItemCard.tsx` currently inline their classes (they predate the
`components/ui` layer). When you extend them, extract the repeated buttons/inputs/cards into
`components/ui` primitives and compose from those — leave the code cleaner than you found it,
consistent with "let the code explain itself."

## Verify

Interactive components deserve tests too (Vitest + `@solidjs/testing-library` + Kobalte render
fine in jsdom): assert the accessible name/role and key interactions, not just that it renders.
Then run the frontend gates (SKILL.md → Verification gates).
