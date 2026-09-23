# Design: Design System Migration — Terminal Brutalism

## Context

The dashboard uses shadcn/ui with Tailwind CSS. The existing design tokens used soft OKLCH values and default rounded radii. The target aesthetic is "terminal-forward brutalism": geometric precision, monospace typography for technical elements, no shadows, and high-contrast borders.

## Goals / Non-Goals

**Goals:**
- Establish a canonical OKLCH color palette in `globals.css`
- Enforce strict geometry rules: 0px buttons, 4px cards, no box shadows
- Apply `font-mono uppercase tracking-wide` to all button labels
- Strip legacy `shadow-sm`, `rounded-2xl` from layout wrappers

**Non-Goals:**
- Changing the component API (props, exports) — style-only changes
- Migrating from shadcn/ui to Radix Themes (that is a separate change: `2026-07-30-radix-migration`)

## Decisions

### Decision 1: Top-down migration order

Apply changes in dependency order: global tokens → primitive components → page layouts → form components. This avoids visual inconsistencies mid-migration.

### Decision 2: OKLCH variables in :root and .dark

Define all color tokens as `--color-*` OKLCH variables in `:root` and override in `.dark`. Tailwind CSS reads these via `@theme inline` in the config.

### Decision 3: CVA-based button geometry

Modify the `cva` base class in `button.tsx` to use `rounded-none`, `font-mono`, `uppercase`, and `tracking-[0.06em]`. All variant classes inherit the base geometry.

### Decision 4: UI compliance check

All visual/interaction decisions must trace to named sections in `DESIGN.md` in `apps/dashboard/`. Any undocumented decision added to `DESIGN.md` before task closure.

## PostgreSQL Migration

Not required.
