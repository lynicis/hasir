# Radix Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Migrate `apps/dashboard` from `@hasir/ui` (shadcn-ui) to `@radix-ui/themes` and remove `packages/ui`.

**Architecture:** We will install `@radix-ui/themes`, configure the `<Theme>` provider, rewrite all dashboard pages/components to use Radix primitives directly, and then tear down the unused `packages/ui` workspace.

**Tech Stack:** Next.js, React, Tailwind CSS, Radix Themes, Bun.

---

### Task 1: Install and Configure Radix Themes

**Files:**
- Modify: `apps/dashboard/package.json`
- Modify: `apps/dashboard/app/layout.tsx`

**Step 1: Install `@radix-ui/themes`**

```bash
bun add @radix-ui/themes --cwd apps/dashboard
```
Expected: Installs successfully in `apps/dashboard`.

**Step 2: Update Layout Provider and Import Styles**

In `apps/dashboard/app/layout.tsx`, wrap the `children` in `<Theme>` and import styles:

```tsx
import "@radix-ui/themes/styles.css";
snip import { Theme } from "@radix-ui/themes";
snip // ... other imports

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <Theme>
          {children}
        </Theme>
      </body>
    </html>
  );
snip }
```

**Step 3: Commit**

```bash
git add apps/dashboard/package.json apps/dashboard/bun.lockb apps/dashboard/app/layout.tsx
git commit -m "chore: setup radix themes in dashboard"
```

---

### Task 2: Migrate Dashboard Components

**Files:**
- Modify: All files in `apps/dashboard/app/` and `apps/dashboard/components/` that import `@hasir/ui/*`.

**Step 1: Identify all shadcn usages**

```bash
grep -r "@hasir/ui" apps/dashboard/
```

**Step 2: Rewrite components**

For each file found, replace `@hasir/ui` components with `@radix-ui/themes` equivalents.
Example:
```tsx
// Before:
import { Button } from "@hasir/ui/components/ui/button";
snip import { Card, CardHeader, CardTitle, CardContent } from "@hasir/ui/components/ui/card";

snip // After:
import { Button, Card, Heading, Box } from "@radix-ui/themes";
snip ```

**Step 3: Run Typecheck**

```bash
make typecheck
```
Expected: PASS

**Step 4: Commit**

```bash
git add apps/dashboard/
git commit -m "refactor: migrate dashboard to radix themes"
```

---

### Task 3: Teardown `@hasir/ui`

**Files:**
- Modify: `apps/dashboard/package.json`
- Modify: `apps/dashboard/components.json`
- Delete: `packages/ui`

**Step 1: Remove `@hasir/ui` from dashboard**

```bash
bun remove @hasir/ui --cwd apps/dashboard
rm apps/dashboard/components.json
```

**Step 2: Delete `packages/ui`**

```bash
rm -rf packages/ui
```

**Step 3: Test and Lint**

```bash
make test
make lint
make build
```
Expected: PASS

**Step 4: Commit**

```bash
git add apps/dashboard/package.json apps/dashboard/bun.lockb apps/dashboard/components.json packages/ui
git commit -m "chore: remove hasir ui package"
```
