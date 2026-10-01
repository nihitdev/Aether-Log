# Aether-Log website

A responsive React + TypeScript project website built with Vite and pnpm.
Warm paper colors, locally bundled fonts, a record-flow illustration, interactive
pipeline stages, accessible Go/C/Rust client tabs, and copyable examples.
A header toggle switches between light and dark mode. It follows the system
preference until a visitor makes a choice, then remembers that choice locally.

## Run locally

Requires Node.js 22.12+ and pnpm. From this directory:

```sh
pnpm install --frozen-lockfile
pnpm dev
```

Vite prints the local address (normally `http://127.0.0.1:5173`). For a chosen
port use `pnpm dev --port 5174 --strictPort`.

## Build and check

```sh
pnpm check
pnpm format:check
pnpm build
pnpm preview
```

`pnpm build` checks TypeScript and generates the static production bundle in
`dist/`. Upload the contents of `dist/` to a static host. No Hub, database,
credentials, environment variables or runtime server is needed by the site.
`base: './'` supports serving under a subdirectory. Documentation uses query
parameters instead of path routes, so it needs no SPA rewrite rule.

Dependencies and fonts are bundled locally; the site does not load fonts, images
or analytics from third-party services. GitHub links navigate to the repository.
Markdown is rendered without raw HTML execution.

## Content

- `src/components/`: landing sections and shared code/documentation components.
- `src/content.ts`: pipeline explanations, quick-start commands and SDK examples.
- `src/styles.css`: responsive layout, reduced-motion support and visual design.
- `public/favicon.svg`: project mark.

The documentation page imports `../docs/architecture.md`,
`../docs/protocol.md`, and `../sdk/rust/README.md` through Vite's raw imports.
Build from the full repository so those source documents are available. The
getting-started page and landing text are maintained in the website source.
Documentation is lazy loaded separately from the landing page.

The animated illustration is explanatory, not a live connection to a Hub.
The site accurately describes delivery limitations and supports exactly Go,
C and Rust. It introduces no SDK or server changes.

`node_modules/`, `dist/` and TypeScript build metadata are ignored. Keep
`pnpm-lock.yaml` for reproducible installs. The website is independent of normal
Go/Rust/C checks.
