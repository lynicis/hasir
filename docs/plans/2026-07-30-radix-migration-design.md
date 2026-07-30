# Radix Themes Migration Design

## Context & Goal
The project currently uses shadcn-ui (centralized in `packages/ui` and imported via `@hasir/ui/*`). The goal is to migrate from shadcn-ui to Radix Themes across the `apps/dashboard` application. 

## Approach: Full Replacement (Strict Radix)
We will completely replace the internal shadcn-ui library with Radix Themes primitives out of the box.

### 1. Setup & Architecture
- Install `@radix-ui/themes` in `apps/dashboard`.
- Update `apps/dashboard/app/layout.tsx` to wrap the children in Radix's `<Theme>` provider.
- Import `@radix-ui/themes/styles.css` in the global styles.

### 2. Component Migration
- Remove all `@hasir/ui` component imports from dashboard pages and components.
- Directly import and use `@radix-ui/themes` components (e.g., `Button`, `Card`, `TextField.Root`, `Heading`, `Text`, `Flex`, `Grid`).
- Map any existing standard patterns (like standard forms or data displays) to Radix layouts.
- Replicate previous shadcn-ui variants/colors using Radix Themes props (`color`, `variant`, `size`).

### 3. Teardown & Cleanup
- Delete `packages/ui` from the monorepo workspace.
- Remove `@hasir/ui` dependency from `apps/dashboard/package.json`.
- Remove any leftover `components.json` configurations.
- Clear out the now-unused shadcn components and configurations.
