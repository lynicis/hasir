# Dashboard Instructions (`apps/dashboard`)

## Commands
| Task | Command |
|------|---------|
| Test single file | `bun test components/login-form.test.tsx` |
| Lint single file | `bunx eslint components/login-form.tsx` |
| Auto-fix lint & import order | `bun run lint:fix` |
| Typecheck workspace | `bun run typecheck` |

## Code Map
| Area | Location |
|------|----------|
| Public & `(authenticated)` routes | `app/` |
| Auth & docs route handlers | `app/api/auth/{login,session,logout}/`, `app/api/docs/[org]/[repo]/[hash]/` |
| Session gate & security headers | `proxy.ts`, `lib/session.ts` |
| Shared Connect transport & client | `lib/transport.ts`, `lib/use-client.ts`, `lib/use-client.impl.ts` |
| Feature components | `components/` |
| Shared UI primitives (`@hasir/ui`) | `../../packages/ui/src/components/` |
| Query invalidation counters | `stores/registry-store.ts` |
| Test harness & module mocks | `test/bun-setup.ts`, `test/happy-dom-setup.ts` |

## Key Conventions
- Reuse `connectTransport` from `lib/transport.ts` (`[csrfInterceptor, authInterceptor, idempotencyInterceptor]`); do not instantiate separate transports.
- Keep `lib/use-client.ts` as the barrel re-exporting `lib/use-client.impl.ts` so `test/bun-setup.ts` module mocks intercept client calls.
- Store auth tokens exclusively in the `iron-session` HTTP-only cookie (`lib/session.ts`); never use `localStorage`.
- Colocate `*.test.ts` / `*.test.tsx` next to source files using `bun test`, `happy-dom`, and `@testing-library/react`.
