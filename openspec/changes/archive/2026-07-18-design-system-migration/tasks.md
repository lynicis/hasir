# Tasks: Design System Migration — Terminal Brutalism

## 1. Global Tokens & Typography

- [x] 1.1 Overwrite OKLCH color variables in `globals.css` `:root` and `.dark` blocks from design spec — verify: `bun run build` passes, no CSS errors
- [x] 1.2 Set `--radius` to `0.75rem`; add explicit `--radius-button: 0` and `--radius-card: 4px` — verify: Tailwind resolves correctly
- [x] 1.3 Confirm `Geist Sans` applied to `body` and `Geist Mono` configured — verify: `make typecheck` passes

## 2. Core UI Components

- [x] 2.1 `button.tsx`: Change base CVA class to `rounded-none font-mono uppercase tracking-[0.06em]` — verify: `bun test components/ui/button`
- [x] 2.2 `card.tsx`: Apply `rounded-sm`; remove `shadow-sm`; enforce `border-border` — verify: visual regression absent
- [x] 2.3 `input.tsx`: Apply `rounded-sm`; remove shadow utilities — verify: `bun test`
- [x] 2.4 `tabs.tsx`: Replace soft hover backgrounds with high-contrast monochrome terminal style — verify: `bun test`

## 3. Settings Page Layouts

- [x] 3.1 `organization/[id]/layout.tsx` and `repository/[repositoryId]/layout.tsx`: Strip `shadow-sm`, `rounded-2xl`; apply `bg-card`; update nav hover states — verify: `make typecheck`

## 4. Profile & Form Components

- [x] 4.1 `profile/page.tsx` and settings forms: Regularize spacing to `gap-4`/`gap-6`; apply `font-mono text-xs` to technical displays; remove local `shadow-sm` — verify: `make typecheck`

## 5. Integration Verification

- [x] 5.1 Run `make lint` and `make typecheck` — verify: both pass
- [x] 5.2 Run `bun test --coverage` for `apps/dashboard` — verify: ≥85% coverage for changed files
- [x] 5.3 Run `bun changeset` — verify: `.changeset/` entry present
