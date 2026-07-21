# DASHBOARD KNOWLEDGE BASE

## OVERVIEW
Next.js 16 web UI for managing organizations, repositories, users, and SSH public keys.

## STRUCTURE
```
apps/dashboard/
├── app/            # Next.js App Router pages, layouts, and providers
├── components/     # UI components, shadcn/ui primitives, and page layouts
├── lib/            # Connect-RPC client, session management, and utilities
├── public/         # Static assets like logos and icons
├── stores/         # Zustand state stores for global client state
└── test/           # Test setup and global type definitions
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Pages & Routing | `app/` | App Router with public and `(authenticated)` route groups |
| UI Components | `components/` | Custom components and shadcn/ui primitives |
| API Client & Hooks | `lib/` | Connect-RPC client setup and custom hooks |
| Global State | `stores/` | Zustand stores for user session and registry state |
| Test Setup | `test/` | Bun test configuration and global types |

## CONVENTIONS
- **State Management Split**: Use Zustand for global client state, such as auth and session, and TanStack Query for server state, such as fetching and caching.
- **API Transport**: Use Connect-RPC, which is gRPC-web, via the custom `useClient` hook in `lib/use-client.ts` for all backend communication.
- **Colocated Tests**: Place `.test.tsx` and `.test.ts` files directly next to the components or utilities they test.
- **Import Ordering**: Follow the Perfectionist plugin rules for consistent import and export ordering.

## ANTI-PATTERNS
- **No Custom Fetch**: Avoid using standard fetch or axios for backend communication. Always use Connect-RPC.
- **No Separate Test Folder**: Don't write tests in a separate tests directory. Colocate them with the source files.
- **No LocalStorage for Tokens**: Never store sensitive tokens in localStorage. Use HTTP-only cookies instead.
- **No Inline Styles**: Avoid inline styles. Use Tailwind CSS utility classes.
