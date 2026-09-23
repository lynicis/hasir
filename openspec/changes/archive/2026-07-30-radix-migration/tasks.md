# Tasks: Radix Themes Migration

## 1. Install and Configure Radix Themes

- [x] 1.1 Run `bun add @radix-ui/themes --cwd apps/dashboard` — verify: package in `apps/dashboard/package.json`
- [x] 1.2 Update `apps/dashboard/app/layout.tsx`: import `@radix-ui/themes/styles.css` and wrap children in `<Theme>` — verify: `make typecheck` passes

## 2. Migrate Dashboard Components

- [x] 2.1 Run `grep -r "@hasir/ui" apps/dashboard/` to list all files requiring update
- [x] 2.2 Replace all `@hasir/ui/*` imports with `@radix-ui/themes` equivalents in each file — verify: `make typecheck` passes for each file after update
- [x] 2.3 Run `bun test --coverage` — verify: ≥85% coverage, no regressions

## 3. Teardown @hasir/ui

- [x] 3.1 Run `bun remove @hasir/ui --cwd apps/dashboard` — verify: not in `package.json`
- [x] 3.2 Delete `apps/dashboard/components.json` — verify: file absent
- [x] 3.3 Delete `packages/ui` — verify: directory absent; `turbo run build` succeeds

## 4. Integration Verification

- [x] 4.1 Run `make build` — verify: 0 errors, 0 warnings
- [x] 4.2 Run `make lint` and `make typecheck` — verify: both pass
- [x] 4.3 Run `bun test --coverage` for `apps/dashboard` — verify: ≥85% coverage
- [x] 4.4 Run `make vuln` — verify: exits 0
- [x] 4.5 Run `bun changeset` — verify: `.changeset/` entry for `hasir-dashboard` minor bump
