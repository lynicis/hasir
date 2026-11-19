# Proposal: Repo-Wide Structural Refactoring

## Implementation Branch
- Worktree: `../hasir-repo-wide-refactor`
- Branch: `refactor/repo-wide-refactor`

## Summary
Perform a behavior-preserving, repository-wide structural refactoring across `apps/api`, `apps/dashboard`, and `packages/ui` to decompose monolithic files, eliminate high-frequency code duplication, enforce `AGENTS.md` architecture boundaries (such as routing external processes through `sdkgenerator.CommandRunner`), and verify quality with CodeRabbit's `code-review` skill.

## Affected Apps & Packages
- `apps/api` (internal domain services, HTTP/SSH transport handlers, PostgreSQL repositories & job queues, `main.go`)
- `apps/dashboard` (API transport/client hooks, error handling, form validation schemas, key-management panel)
- `packages/ui` (`file-tree` component utilities, CSS theme entrypoint)

## Schema, Protobuf & Secrets Impact
- **Database migrations**: None required.
- **Protobuf schemas**: No `.proto` files modified (zero breaking or additive proto changes).
- **Environment variables / secrets**: No new environment variables or secrets introduced.

## Non-goals
- Changing observable API behavior, RPC signatures, HTTP/SSH routes, or UI visual appearance.
- Adding new product features or modifying PostgreSQL schemas.
- Modifying `apps/landing` or `deploy/` Helm/Docker manifests.
