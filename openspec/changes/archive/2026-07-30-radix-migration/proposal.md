# Proposal: Radix Themes Migration

## Why

The dashboard used an internal `packages/ui` package built on shadcn/ui components. Maintaining a custom component library adds overhead and diverges from upstream improvements. Migrating to `@radix-ui/themes` gives us first-class primitives with built-in accessibility, theming, and a smaller maintenance surface. The `packages/ui` package can be deleted entirely.

## What Changes

- `apps/dashboard`: `@radix-ui/themes` installed; `<Theme>` provider wraps the app layout; `@radix-ui/themes/styles.css` imported
- All `@hasir/ui/*` imports in `apps/dashboard/app/` and `apps/dashboard/components/` replaced with `@radix-ui/themes` equivalents
- `packages/ui` workspace deleted; `@hasir/ui` devDep removed from `apps/dashboard`
- `apps/dashboard/components.json` removed

## Capabilities

### Modified Capabilities
- `dashboard/ui`: Component layer migrated from custom shadcn wrapper to `@radix-ui/themes` primitives

## Impact

- `apps/dashboard/package.json`: `@radix-ui/themes` added; `@hasir/ui` removed
- `apps/dashboard/app/layout.tsx`: `<Theme>` wrapper + CSS import
- All `apps/dashboard/app/**/*.tsx` and `components/**/*.tsx` files that import `@hasir/ui`: updated imports
- `apps/dashboard/components.json`: Deleted
- `packages/ui/`: Directory deleted from monorepo

## Implementation Branch

`radix-migration` (archived — completed)
