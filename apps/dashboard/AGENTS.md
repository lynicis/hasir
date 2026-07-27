# DASHBOARD KNOWLEDGE BASE

## OVERVIEW
Next.js 16 (App Router, React 19, Turbopack) management UI: organizations, repositories, users, SSH keys, SDK docs. Bun test + happy-dom.

## STRUCTURE
```
apps/dashboard/
├── app/            # Routes: public (/login,/register,/invite/[token]) + (authenticated) group
│   └── api/        # Route handlers: auth/{login,session,logout}, docs/[org]/[repo]/[hash]
├── components/     # Feature components + ui/ shadcn primitives (see components/AGENTS.md)
├── lib/            # Connect-RPC client, session, hooks (19 files)
├── public/         # Static assets
├── stores/         # registry-store.ts only (Zustand query-invalidation counters)
├── test/           # bun-setup, happy-dom-setup, types.d.ts
├── proxy.ts        # Next 16 middleware: session gate + CSP/security headers
└── instrumentation.ts
```

## ENTRY / DATA FLOW
`app/layout.tsx` (Geist fonts) → `app/providers.tsx`: TransportProvider(`createConnectTransport(NEXT_PUBLIC_API_URL, [authInterceptor])`) → QueryClientProvider → SessionProvider → ThemeProvider → Toaster.
`proxy.ts` gates non-public paths on the `hasir-session` cookie; public: `/`, `/login`, `/register`, `/forgot-password`, `/reset-password`, `/docs/ssh-configuration`.
Auth route handlers: `api/auth/login` (Connect login → decode JWT → iron-session cookie, 7d httpOnly), `api/auth/session` (GET auto-refresh via renewTokens; POST zod-validated swap), `api/auth/logout`.
`lib/use-client.ts`: `useClient(service)` = memoized createClient + authInterceptor (skips public methods, fetches `/api/auth/session`, 401 → logout + redirect `/login`). connect-query `useQuery` + customRetry (no retry on NotFound, max 3).

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Pages & Routing | `app/` | Public + `(authenticated)` route groups |
| Auth/session | `app/api/auth/`, `lib/session.ts` | iron-session, server-only |
| API Client & Hooks | `lib/use-client.ts` | Never fetch directly |
| UI Components | `components/` | See `components/AGENTS.md` |
| Global State | `stores/registry-store.ts` | Invalidation counters only |
| Tests | colocated `*.test.tsx` | bun test, 55+ tests |

## CONVENTIONS
- **State Management Split**: TanStack Query for server state; Zustand (`stores/registry-store.ts`) only as version counters for query invalidation.
- **API Transport**: Connect-RPC via the `useClient` hook in `lib/use-client.ts` for all backend communication.
- **Colocated Tests**: `.test.tsx`/`.test.ts` next to the source; module-level `mock.module` mocks exposed via `globalThis.__client` / `__connectQueryMocks` / `__toast`.
- **Test Stack**: bun test + happy-dom + @testing-library/react. NO vitest/jest/MSW/playwright (README is stale).
- **Import Ordering**: Perfectionist plugin, line-length desc. `@/*` → `./*` alias. Next `output: standalone`; `turbopack.root` = monorepo root.

## ANTI-PATTERNS
- **No Custom Fetch**: No fetch/axios to the backend — use Connect-RPC. Sole exception: `use-documentation.ts` via the `/api/docs` proxy.
- **No Separate Test Folder**: Colocate tests with source files.
- **No LocalStorage for Tokens**: iron-session HTTP-only cookie only.
- **No Inline Styles**: Tailwind CSS utility classes.

## NOTES
- `lib/use-client.impl.ts` is a byte-identical duplicate of `use-client.ts` — edit both or delete one.
- The Connect transport is instantiated twice (`providers.tsx` + `use-client.ts`).
- Hotspots: `components/ssh-api-key-panel.tsx` 597 LOC; `components/ui/file-tree.tsx` 507 LOC.
- README references a `user-store` — stale; only `registry-store.ts` exists.
