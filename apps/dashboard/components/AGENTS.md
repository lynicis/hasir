# DASHBOARD COMPONENTS KNOWLEDGE BASE

## OVERVIEW
This directory contains shadcn/ui primitives and custom feature components for the Hasir dashboard.

## WHERE TO LOOK
| Component Type | Location | Description |
|----------------|----------|-------------|
| UI Primitives | `ui/` | Reusable shadcn/ui components (e.g., Button, Dialog, Input, Spinner) |
| Feature Components | `./` | Custom dashboard components (e.g., `clone-urls.tsx`, `danger-zone.tsx`, forms, lists) |
| Component Tests | `./` & `ui/` | Colocated `*.test.tsx` files next to their respective components |

## CONVENTIONS
- **Named Exports**: Export all components using named exports (e.g., `export function ModeToggle()`) rather than default exports.
- **Client Directives**: Explicitly mark interactive components with `"use client"` at the top of the file.
- **Tailwind Styling**: Use Tailwind CSS utility classes exclusively. Avoid custom CSS or inline styles.
- **Radix UI Primitives**: Extend Radix UI primitives via shadcn/ui patterns in the `ui/` directory.
- **Component Naming**: Use kebab-case for filenames (e.g., `clone-urls.tsx`) and PascalCase for component names (e.g., `CloneUrls`).
- **Props Typing**: Always define explicit TypeScript interfaces for component props (e.g., `interface CloneUrlsProps`).

## ANTI-PATTERNS
- **No Direct API Calls**: Do not perform direct fetch or axios calls inside components. Use Connect-RPC client hooks.
- **No Inline Styles**: Do not use the `style` prop for styling. Use Tailwind classes.
- **No Global State in Components**: Do not manage global state locally. Use Zustand stores from `@/stores`.
- **No Unused shadcn/ui Imports**: Do not import unused primitives or duplicate shadcn/ui components.
- **No Direct DOM Manipulation**: Do not use `document.getElementById` or direct ref manipulation for styling or visibility. Use React state.
- **No Hardcoded API Endpoints**: Do not hardcode API URLs. Use environment variables or configuration helpers.

## TESTING
- **Testing Framework**: Use Bun test runner for component unit tests.
- **Mocking**: Mock Connect-RPC clients and Zustand stores where necessary to isolate component behavior.
