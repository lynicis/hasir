# Design: Radix Themes Migration

## Context

`packages/ui` contained shadcn/ui components vendored into the monorepo. The dashboard imported them as `@hasir/ui`. Radix Themes (`@radix-ui/themes`) provides styled, accessible primitives out-of-the-box, eliminating the need for the internal package. The migration is a full replacement — no shadcn components remain after completion.

## Goals / Non-Goals

**Goals:**
- Complete removal of `packages/ui` and all `@hasir/ui` imports
- All UI built from `@radix-ui/themes` primitives directly in the dashboard
- `<Theme>` provider wrapping the app for consistent token propagation

**Non-Goals:**
- Custom Radix Themes overrides or a new internal UI package
- Mixing shadcn and Radix components during or after migration
- Changes to `apps/landing`

## Decisions

### Decision 1: Full replacement, no gradual migration

Migrate all `@hasir/ui` usages in one change. Gradual migration would leave `packages/ui` in place temporarily, increasing maintenance burden. The scope of `@hasir/ui` usage is bounded to `apps/dashboard`.

### Decision 2: Prop mapping

Map shadcn variants to Radix props using the Radix Themes `color`, `variant`, and `size` system. No custom CSS overrides needed for standard patterns.

### Decision 3: <Theme> wraps <body> in layout.tsx

`@radix-ui/themes/styles.css` is imported first in `layout.tsx`; `<Theme>` wraps the body children. This is an RSC-compatible pattern.

## PostgreSQL Migration

Not required.
