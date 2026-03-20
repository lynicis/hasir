# AGENTS.md — Hasir Dashboard

Next.js 16 (App Router) + React 19 + TypeScript 5 dashboard for the Hasir protobuf schema registry.
Uses Connect-RPC (gRPC-web), TanStack Query, Zustand, react-hook-form + Zod v4, shadcn/ui (Radix + Tailwind v4).

## Build / Lint / Test Commands

Package manager is **Bun**.

```sh
bun install              # Install dependencies
bun dev                  # Start dev server (next dev)
bun run build            # Production build (next build, standalone output)
bun run lint             # ESLint (flat config, eslint 9)
bun run lint:fix         # ESLint with --fix

bun vitest               # Run tests in watch mode
bun vitest run           # Run all tests once
bun vitest run path/to/file.test.tsx          # Single test file
bun vitest run -t "test name"                 # Single test by name
bun vitest run path/to/file.test.tsx -t "it"  # Single test in file
bun run test:ci          # CI mode: no watch, coverage, junit output
```

Test framework: **Vitest** (jsdom env, `@vitejs/plugin-react`, globals enabled, 15 s timeout).
Tests use `@testing-library/react`, `@testing-library/user-event`, `@testing-library/jest-dom`.

## Project Structure

```
app/                     # Next.js App Router pages and API routes
  (authenticated)/       # Protected route group (dashboard, org, repo, profile, invite)
  api/                   # Route handlers (auth, docs proxy)
components/              # Feature components with co-located *.test.tsx files
  ui/                    # shadcn/ui primitives (do not edit by hand)
lib/                     # Utilities, hooks, context providers
stores/                  # Zustand stores
proxy.ts                 # Auth-redirect middleware
instrumentation.ts       # OpenTelemetry setup
```

## Code Style

### Imports

- Path alias: `@/*` maps to project root (`@/components/ui/button`, `@/lib/utils`).
- Relative imports only for siblings in the same directory (`./utils`, `./search-dropdown`).
- Sorted by **line-length descending** (enforced by `eslint-plugin-perfectionist`).
  Type imports are separated from value imports.
- External packages first, then `@/` local imports, each group longest-line-first.

### Formatting

- No Prettier or Biome — ESLint is the sole formatter/linter.
- Indentation: 2 spaces (tabs are not used).
- Trailing commas in multi-line structures.
- Use `cn()` from `@/lib/utils` (clsx + tailwind-merge) to compose class names.

### TypeScript

- **Strict mode** enabled. Do not add `any` without an eslint-disable comment.
- Use `interface` for component props and object shapes.
- Use `type` for Zod-inferred types (`type ISchema = z.infer<typeof schema>`),
  union types, and simple aliases.
- Prefix Zod-inferred form types with `I` (e.g., `ISchema`, `ILoginSchema`).
- Generics where appropriate (e.g., `useDebounce<T>`, `useClient<T extends DescService>`).

### Naming Conventions

| Element            | Convention              | Example                            |
| ------------------ | ----------------------- | ---------------------------------- |
| Files              | `kebab-case.tsx`/`.ts`  | `repository-dialog-form.tsx`       |
| Test files         | `*.test.tsx`/`.test.ts` | `repository-dialog-form.test.tsx`  |
| Components         | PascalCase              | `RepositoryDialogForm`             |
| Functions/vars     | camelCase               | `handleFormSubmit`, `isLoading`    |
| Constants          | SCREAMING_SNAKE_CASE    | `DEFAULT_PAGE_LIMIT`               |
| Types/Interfaces   | PascalCase              | `SessionData`, `DeleteMemberDialogProps` |
| Custom hooks       | `use` prefix            | `useClient`, `useDebounce`         |
| Zustand stores     | `use` + domain + Store  | `useRegistryStore`                 |

### Component Patterns

- **`"use client"`** on all interactive components. Server components are thin page
  wrappers that set metadata and delegate to client components.
- **Named exports** for components (`export function Header() {}`).
  Default exports only for page-level content components.
- Props defined as `interface` directly above the component.
- Skeleton variants co-exported from the same file (e.g., `LoginFormSkeleton`).
- Dynamic imports with `ssr: false` for client-only components.

### Forms

- **react-hook-form** + `zodResolver` with Zod v4 (`import { z } from "zod/v4"`).
- Schema: `const schema = z.object({...})`, type: `type ISchema = z.infer<typeof schema>`.
- Use `Controller` component pattern (not `register`).
- Field components: `Field`, `FieldGroup`, `FieldLabel`, `FieldError` from `@/components/ui/field`.
- Show API errors via `toast.error()`. Show form-level errors via `setError("root", {...})`.

### Error Handling

- Wrap API calls in `try/catch`. Check error type with `ConnectError` + error code:
  `error instanceof ConnectError && error.code === Code.NotFound`.
- Use type guards from `@/lib/utils`: `isNotFoundError()`, `isUnauthenticatedError()`.
- User-facing errors: `toast.error("Failed to ...")` from Sonner.
- Simple catch blocks may omit the error binding: `catch { toast.error(...) }`.
- Auth errors: the Connect-RPC interceptor catches `Code.Unauthenticated` and redirects to `/login`.
- Custom retry: `customRetry` from `@/lib/query-retry` skips retry on NotFound, max 3 retries.

### State Management

- **Server state**: TanStack Query via `@connectrpc/connect-query` generated hooks.
- **Client state**: Zustand stores in `stores/` directory.
- **Sessions**: `iron-session` on the server, `SessionProvider` context on the client.

### Testing

- Co-locate tests next to source files (`component.tsx` + `component.test.tsx`).
- Use `describe`/`it`/`expect` from Vitest globals.
- Mock with `vi.fn()`, `vi.mock()`.
- Component tests: `render()`, `screen`, `userEvent.setup()`.
- Cover rendering, user interactions, error states, and edge cases.
- Avoid `any` in tests when possible; use typed mocks or `as unknown as Type`.

### Git Conventions

- **Commit style**: Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`).
- Pre-commit hooks run ESLint and Vitest via `.pre-commit-config.yaml`.
- CI runs lint + test on push to `master`; coverage uploaded to Codecov.

### Things to Avoid

- Do not manually edit files in `components/ui/` — they are managed by shadcn CLI.
- Do not use Prettier or Biome — ESLint handles all linting.
- Do not add `any` without an explicit eslint-disable comment.
- Do not use `register` pattern in forms — use `Controller`.
- Do not import from `zod` directly — import from `zod/v4`.
- Do not commit `.env` files — only `.env.example` is tracked.
