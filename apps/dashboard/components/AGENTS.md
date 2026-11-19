# Dashboard Components Instructions (`apps/dashboard/components`)

## Commands
| Task | Command |
|------|---------|
| Test single component | `cd apps/dashboard && bun test components/clone-urls.test.tsx` |

## Code Map
| Type | Location |
|------|----------|
| Dashboard feature components | `./` (`*.tsx` + colocated `*.test.tsx`) |
| Shared shadcn/Radix UI primitives | `../../../packages/ui/src/components/` (`@hasir/ui`) |

## Key Conventions
- Import reusable UI primitives from `@hasir/ui`; place dashboard-specific feature components in this directory.
- Use named exports (`export function ComponentName()`) and kebab-case filenames (`component-name.tsx`).
- Add `"use client"` at the top of interactive components.
- Style exclusively with Tailwind CSS utility classes and tokens defined in `DESIGN.md`; never use inline `style` props.
- Mock Connect clients in tests via `globalThis.__client`, `__connectQueryMocks`, and `__toast` (configured in `../test/bun-setup.ts`).
