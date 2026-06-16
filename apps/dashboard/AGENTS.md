# AGENTS.md — Hasir Dashboard

Next.js 16 (App Router, `output: "standalone"`) + React 19 + TypeScript 6 dashboard for the Hasir protobuf schema registry.
Uses Connect-RPC (gRPC-web), TanStack Query (via `@connectrpc/connect-query`), Zustand, react-hook-form + Zod v4, shadcn/ui (Radix + Tailwind v4, new-york style, lucide icons, Magic UI registry).

## Build / Lint / Test Commands

Package manager is **Bun**.

```sh
bun install              # Install dependencies (uses bun.lock, frozen-lockfile in CI)
bun dev                  # bun --bun run next dev (Turbopack)
bun run build            # bun --bun run next build (standalone output)
bun run start            # bun --bun run next start
bun run lint             # ESLint (flat config, eslint 9, eslint-config-next)
bun run lint:fix         # ESLint with --fix
bun run typeCheck        # TypeScript type checking (noEmit)

bun test                 # Run all tests (Bun's built-in test runner, happy-dom)
bun test path/to/file.test.tsx          # Single test file
bun test -t "test name"                 # Single test by name
bun test path/to/file.test.tsx -t "it"  # Single test in file
bun run test:ci          # CI mode: bun test --reporter=junit, coverage via --coverage
```

Test framework: **Bun test** (`bun:test` globals, `mock.module` for mocking, happy-dom via `test/bun-setup.ts`, configured in `bunfig.toml`).
Test libraries: `@testing-library/react`, `@testing-library/user-event`, `@testing-library/jest-dom`.
Vitest is a devDependency (`vitest`, `@vitest/coverage-v8`, `@vitejs/plugin-react`) but is **not wired into any npm script** — use Bun test only.

## Project Structure

```
app/                     # Next.js App Router pages and API routes
├── globals.css          # Tailwind v4 CSS entry
├── layout.tsx           # Root layout (html, body, SessionProvider, QueryClientProvider)
├── providers.tsx        # Client providers (ThemeProvider, Toaster, SessionProvider)
├── page.tsx             # Landing page
├── (authenticated)/     # Protected route group
│   ├── layout.tsx       # Auth layout with HeaderClient sidebar/nav
│   ├── dashboard/       # Dashboard page
│   ├── organization/
│   │   └── [id]/
│   │       ├── layout.tsx           # Org layout (context provider)
│   │       ├── page.tsx             # Organization overview
│   │       ├── repositories/        # Org repos list
│   │       ├── settings/            # Org settings
│   │       └── users/               # Org member management
│   ├── profile/         # User profile
│   ├── repository/
│   │   └── [repositoryId]/
│   │       ├── layout.tsx           # Repo layout (context provider)
│   │       ├── page.tsx             # Repo overview
│   │       ├── commits/             # Commit history
│   │       ├── documentation/       # Protobuf docs viewer
│   │       ├── files/               # File browser
│   │       ├── sdk-preferences/     # SDK generation config
│   │       └── settings/            # Repo settings
│   └── invite/[token]/  # Invite acceptance
├── api/                 # Route handlers
│   ├── auth/            # login, logout, session (iron-session)
│   │   ├── login/
│   │   ├── logout/
│   │   └── session/
│   └── docs/[organizationId]/[repositoryId]/[commitHash]/
│                       # Protobuf docs proxy
├── docs/
│   └── ssh-configuration/
│                       # SSH configuration guide (public)
├── login/
├── register/
├── forgot-password/
└── reset-password/[token]/

components/              # Feature components (43 source + 40 co-located tests)
├── brand-link.tsx
├── clone-urls.tsx
├── danger-zone.tsx
├── dashboard.tsx
├── delete-member-dialog.tsx
├── delete-organization-dialog.tsx
├── delete-repository-dialog.tsx
├── forgot-password-form.tsx
├── header.tsx
├── header-client.tsx
├── home-page-content.tsx
├── invite-response.tsx
├── invite-user-dialog.tsx
├── login-form.tsx
├── markdown-renderer.tsx
├── member-item.tsx
├── members-list.tsx
├── organization-dialog-form.tsx
├── organization-repositories-content.tsx
├── organization-settings-content.tsx
├── organization-settings-form.tsx
├── organization-users-content.tsx
├── password-confirmation-dialog.tsx
├── profile-form.tsx
├── profile-page-content.tsx
├── register-form.tsx
├── repositories-list.tsx
├── repository-commits-content.tsx
├── repository-dialog-form.tsx
├── repository-documentation-content.tsx
├── repository-files-content.tsx
├── repository-item.tsx
├── repository-sdk-preferences-content.tsx
├── repository-settings-content.tsx
├── repository-settings-form.tsx
├── reset-password-form.tsx
├── sdk-install-guide-dialog.tsx
├── sdk-urls.tsx
├── search-dropdown.tsx
├── ssh-api-key-panel.tsx
├── ssh-configuration-client.tsx
├── ssh-configuration-content.tsx
├── theme-provider.tsx
└── theme-toggle.tsx

components/ui/           # shadcn/ui primitives (do not edit by hand)
├── alert-dialog.tsx
├── alert.tsx
├── avatar.tsx
├── button.tsx
├── card.tsx
├── dialog.tsx
├── dropdown-menu.tsx
├── field.tsx             # Field/FieldGroup/FieldLabel/FieldError pattern
├── file-tree.tsx
├── input.tsx
├── input-group.tsx
├── kbd.tsx
├── label.tsx
├── pagination.tsx        # + pagination.test.tsx
├── popover.tsx
├── radio-group.tsx
├── scroll-area.tsx
├── select.tsx
├── separator.tsx
├── skeleton.tsx
├── sonner.tsx            # Sonner toast wrapper
├── spinner.tsx
├── switch.tsx
├── tabs.tsx
└── textarea.tsx

lib/                     # Utilities, hooks, context providers
├── auth-interceptor.ts  # Connect-RPC interceptor (Unauthenticated → /login redirect)
├── query-retry.ts       # customRetry (skip on NotFound, max 3 retries)
├── repository-context.tsx # Repository context provider
├── session-provider.tsx # Session context (getSession, setSession, clearSession)
├── session.ts           # iron-session helpers
├── use-client.ts        # Generic Connect-RPC client hook (public API)
├── use-client.impl.ts   # Internal implementation (same as use-client.ts)
├── use-debounce.ts      # Generic debounce hook
├── use-documentation.ts # Protobuf docs viewer logic
├── utils.ts             # cn(), isNotFoundError(), isUnauthenticatedError()
└── visibility-mapper.ts # Visibility level display helpers

stores/                  # Zustand stores
└── registry-store.ts   # Global registry state

test/                    # Bun test setup
└── bun-setup.ts         # happy-dom setup, mock.module for packages

proxy.ts                 # Auth-redirect middleware (Next.js middleware pattern)
proxy.test.ts            # Co-located test for proxy
instrumentation.ts       # OpenTelemetry via @vercel/otel
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

- **Strict mode** enabled (`tsconfig.json: strict: true`).
- `noUncheckedIndexedAccess: true`, `verbatimModuleSyntax: true`, `noImplicitOverride: true`.
- Do not add `any` without an eslint-disable comment.
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
- **Client state**: Zustand stores in `stores/` directory (currently `registry-store.ts`).
- **Sessions**: `iron-session` on the server, `SessionProvider` context on the client.

### Testing

- Co-locate tests next to source files (`component.tsx` + `component.test.tsx`).
- Use `describe`/`it`/`expect` from `bun:test` globals (not Vitest globals).
- Mock with `mock.module(...)` (Bun's built-in).
- Component tests: `render()`, `screen`, `userEvent.setup()`.
- Cover rendering, user interactions, error states, and edge cases.
- Avoid `any` in tests when possible; use typed mocks or `as unknown as Type`.

### Environment Variables

Three vars in `.env.example`:
- `NEXT_PUBLIC_API_URL` — Connect-RPC gRPC-web endpoint (default `http://localhost:8080`)
- `NEXT_PUBLIC_BASE_URL` — App base URL (default `http://localhost:3000`)
- `SESSION_SECRET` — iron-session encryption key (min 32 chars)

### Git Conventions

- **Commit style**: Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`).
- Pre-commit hooks via `.pre-commit-config.yaml` — lint + test.

### Docker

Multi-stage `Dockerfile`:
- **Build stage**: `oven/bun:1-alpine` — installs deps, runs `bun run build`.
- **Release stage**: `node:24-alpine` — copies `.next/standalone`, `.next/static`, `public`; runs as non-root `hasir` user.

### Things to Avoid

- Do not manually edit files in `components/ui/` — they are managed by shadcn CLI.
- Do not use Prettier or Biome — ESLint handles all linting.
- Do not add `any` without an explicit eslint-disable comment.
- Do not use `register` pattern in forms — use `Controller`.
- Do not import from `zod` directly — import from `zod/v4`.
- Do not commit `.env` files — only `.env.example` is tracked.

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **dashboard** (1124 symbols, 2689 relationships, 59 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

> Index stale? Run `node .gitnexus/run.cjs analyze` from the project root — it auto-selects an available runner. No `.gitnexus/run.cjs` yet? `npx gitnexus analyze` (npm 11 crash → `npm i -g gitnexus`; #1939).

## Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows. For regression review, compare against the default branch: `detect_changes({scope: "compare", base_ref: "master"})`.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `query({query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `context({name: "symbolName"})`.

## Never Do

- NEVER edit a function, class, or method without first running `impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `rename` which understands the call graph.
- NEVER commit changes without running `detect_changes()` to check affected scope.

## Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/dashboard/context` | Codebase overview, check index freshness |
| `gitnexus://repo/dashboard/clusters` | All functional areas |
| `gitnexus://repo/dashboard/processes` | All execution flows |
| `gitnexus://repo/dashboard/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
|------|---------------------|
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->
