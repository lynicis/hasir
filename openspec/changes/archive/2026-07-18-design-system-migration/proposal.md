# Proposal: Design System Migration — Terminal Brutalism

## Why

The dashboard used soft, rounded design tokens inherited from shadcn defaults. To establish a distinctive product identity ("terminal-forward brutalism"), the global color palette, typography, and component geometry needed a full overhaul targeting 0px button radius, 4px card radius, no box shadows, and OKLCH color variables.

## What Changes

- `apps/dashboard/app/globals.css`: OKLCH color palette applied to `:root` and `.dark`; `--radius` and per-component radius variables updated
- Core shadcn UI components updated: `button.tsx` (rounded-none, font-mono uppercase), `card.tsx` (rounded-sm, no shadow), `input.tsx` (rounded-sm, no shadow), `tabs.tsx` (high-contrast terminal style)
- Organization/Repository settings layouts: sidebar shadows and legacy rounded classes removed
- Profile and form components: spacing regularized to baseline grid; `font-mono` applied to technical displays

## Capabilities

### Modified Capabilities
- `dashboard/ui`: visual language standardized to terminal-brutalism design system

## Impact

- `apps/dashboard/app/globals.css`: Complete OKLCH palette
- `apps/dashboard/components/ui/button.tsx`, `card.tsx`, `input.tsx`, `tabs.tsx`: Structural style updates
- `apps/dashboard/app/(authenticated)/organization/[id]/layout.tsx`
- `apps/dashboard/app/(authenticated)/repository/[repositoryId]/layout.tsx`
- `apps/dashboard/app/(authenticated)/profile/page.tsx`
- `apps/dashboard/components/organization-settings-form.tsx`
- `apps/dashboard/components/repository-settings-form.tsx`

## Implementation Branch

`design-system-migration` (archived — completed)
