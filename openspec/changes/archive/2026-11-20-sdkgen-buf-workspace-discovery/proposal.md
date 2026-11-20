# Proposal: SDK generation honors the repository's buf workspace

## Summary
SDK and documentation generation currently writes a `buf.yaml` with `modules: [{path: "."}]` at the checkout root and ignores any `buf.yaml` in a subdirectory. Pushing a monorepo whose buf workspace lives in `packages/proto/` breaks every local import (`imported file does not exist`). Generation will now discover the repository's own buf workspace and run buf there.

## Affected Apps & Packages
- `apps/api/pkg/sdkgenerator` (buf generator, documentation generator, proto discovery)

## Schema, Protobuf & Secrets Impact
- **Database migrations**: None.
- **Protobuf schemas**: None.
- **Environment variables / secrets**: None.

## Non-goals
- A BSR proxy. Remote deps already resolve directly from `buf.build`; the failure was local import resolution.
- Merging multiple sibling buf workspaces automatically.
