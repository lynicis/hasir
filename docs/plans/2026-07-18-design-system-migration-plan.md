# Design System Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement this plan task-by-task.

**Goal:** Migrate the application to the new "terminal-forward brutalism" design system by globally updating core color tokens, typography, and standardizing component border-radii/shadows, specifically targeting the Profile, Organization, and Repository Settings pages.

**Architecture:** Top-down approach. We'll first update the foundational layer (`globals.css` colors, layout typography rules), then update the core primitive components (`Button`, `Card`, `Input`, `Tabs` etc.) to reflect the strict geometrical styling (0px for buttons, 4px for cards, removed shadows). Lastly, we will strip legacy utility classes (like `shadow-sm`) directly from the layout wrapper components of the target pages.

**Tech Stack:** Next.js (App Router), Tailwind CSS, shadcn/ui.

---

### Task 1: Update Global Design Tokens & Typography

**Files:**
- Modify: `apps/dashboard/app/globals.css`
- Modify: `apps/dashboard/app/layout.tsx` (if typography imports need adjustments)
- Modify: `apps/dashboard/tailwind.config.ts` (if applicable)

**Step 1:** Review and overwrite OKLCH color variables in `globals.css` for both `:root` and `.dark` blocks, pasting them exactly from `DESIGN.md`. Ensure that sidebar-specific tokens and chart tokens are correctly aligned.

**Step 2:** In `globals.css` or Tailwind config, update the base `--radius` variable. Currently it is likely `0.5rem` or `0.75rem`. Set it to `0.75rem` so the computed `--radius-md` equals 10px and `--radius-sm` equals 8px, but also explicitly add utility values if necessary to force Cards to use `4px` and Buttons to use `0px`.

**Step 3:** Ensure that `Geist Sans` is applied seamlessly to `<body>` and `Geist Mono` is configured properly.

**Step 4:** Commit
```bash
git add apps/dashboard/app/globals.css
git commit -m "style: apply global OKLCH color palette and typography tokens"
```

---

### Task 2: Refactor Core UI Components

**Files:**
- Modify: `apps/dashboard/components/ui/button.tsx`
- Modify: `apps/dashboard/components/ui/card.tsx`
- Modify: `apps/dashboard/components/ui/input.tsx`
- Modify: `apps/dashboard/components/ui/tabs.tsx`

**Step 1:** In `button.tsx`, modify the base CVA styling.
- Change `rounded-md` to `rounded-none`.
- Add `font-mono uppercase tracking-[0.06em]`.

**Step 2:** In `card.tsx`, modify the base styling.
- Ensure the card wrapper uses `rounded-sm` (4px).
- Remove `shadow-sm` and any box-shadow utilities.
- Ensure border color uses `border-border` strictly.

**Step 3:** In `input.tsx`, update styles to use `rounded-sm` and pure hairlines (no shadow). 

**Step 4:** In `tabs.tsx`, update the active tab styling to remove soft background grays, replacing it with the sharp contrast terminal styles defined in the layout section of the design spec.

**Step 5:** Commit
```bash
git add apps/dashboard/components/ui
git commit -m "style: apply brutalist structural updates to core components"
```

---

### Task 3: Refactor Organization & Repository Settings Pages

**Files:**
- Modify: `apps/dashboard/app/(authenticated)/organization/[id]/layout.tsx`
- Modify: `apps/dashboard/app/(authenticated)/repository/[repositoryId]/layout.tsx`

**Step 1:** In both layout files, locate the `<Card>` wrapper that acts as the sidebar container.
- Strip any remaining `shadow-sm`, `rounded-2xl` or overriding padding classes that conflict with the global core components.
- Ensure the background utilizes `bg-card` (or `bg-sidebar` if using the sidebar token).

**Step 2:** Update the sidebar navigation items. 
- Replace hover classes that apply soft backgrounds (`hover:bg-accent`) with the sharper, high-contrast monochrome transitions dictated by the new brutalist design logic.

**Step 3:** Commit
```bash
git add apps/dashboard/app/\(authenticated\)/organization/[id]/layout.tsx
git add apps/dashboard/app/\(authenticated\)/repository/[repositoryId]/layout.tsx
git commit -m "style: update org and repo settings layouts to match new design system"
```

---

### Task 4: Refactor Profile Settings Page & Content Forms

**Files:**
- Modify: `apps/dashboard/app/(authenticated)/profile/page.tsx`
- Modify: `apps/dashboard/components/organization-settings-form.tsx`
- Modify: `apps/dashboard/components/repository-settings-form.tsx`

**Step 1:** In the Profile page layout and the `<OrganizationSettingsForm>`, `<RepositorySettingsForm>` components, identify form container layout gaps. Ensure spacing utilizes standard baseline grid variables (`gap-6` or `gap-4`).

**Step 2:** Check all typographic hierarchies inside settings fields. 
- Ensure field labels use standard Sans typography.
- Ensure any technical ID displays or configuration snippets use `font-mono text-xs`.
- Remove any local `shadow-sm` classes applied to inner containers.

**Step 3:** Commit
```bash
git add apps/dashboard/components/
git add apps/dashboard/app/\(authenticated\)/profile/
git commit -m "style: polish profile and settings forms for typography and spacing"
```