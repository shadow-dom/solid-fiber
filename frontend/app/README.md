# solid-fiber — frontend

The frontend for **solid-fiber**: a Solid + UnoCSS (preset-wind4) single-page app
that is served by a Go [Fiber](https://gofiber.io) backend.

## Package manager

This app uses [bun](https://bun.sh). Install dependencies with:

```bash
bun install
```

## Development

```bash
bun run dev
```

Starts Vite on [http://localhost:5173](http://localhost:5173). Requests to `/api`
are proxied to the Fiber backend on [http://localhost:3000](http://localhost:3000),
so run the backend alongside the dev server.

## Scripts

- `bun run dev` — start the Vite dev server on :5173 (proxies `/api` to the backend on :3000).
- `bun run build` — type-check (`tsc --noEmit`) then build for production.
- `bun run typecheck` — run `tsc --noEmit` with no emit, failing on type errors.
- `bun run test` — run the [Vitest](https://vitest.dev) suite once (`test:watch` for watch mode).
- `bun run lint` — run ESLint (with the Solid plugin); warnings fail.
- `bun run serve` — preview the production build locally.

## Build output

`bun run build` outputs to `../../backend/web/dist` rather than a local `dist/`.
That directory is embedded into the Go binary, so a production build of the Fiber
backend serves the compiled frontend directly.
