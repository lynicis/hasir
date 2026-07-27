# PROTO PACKAGE KNOWLEDGE BASE

## OVERVIEW
Protobuf schema source of truth + buf workspace. Publishes to the Buf Schema Registry (BSR) via its own release channel. npm package `@hasir/proto` wraps generated code for TS consumers.

## STRUCTURE
```
packages/proto/
├── proto/                 # Schemas — DOUBLE-NESTED path
│   └── {user,organization,registry,shared}/v1/*.proto
├── buf.yaml               # Workspace + lint/breaking config
├── buf.gen.yaml           # Codegen targets (go + js)
├── buf.lock
├── gen/{go,js}/           # Generated output — GITIGNORED, never commit
├── go.mod / Makefile      # Go module for the generated Go SDK
└── package.json           # @hasir/proto npm wrapper
```

## CONVENTIONS
- All proto work via buf CLI: root `make proto` = turbo `proto` task = `buf generate` here.
- Generated code (`gen/`) is gitignored (ADR-0003) — regenerated on demand, cached by turbo.
- CI (`proto.yml`): `buf lint` + `buf breaking` against main on every PR touching `packages/proto/**`.

## RELEASE
- Tag `proto/v*` → `proto-release.yml` → buf lint/breaking/push to BSR (with label).
- Independent of app releases (`hasir-*@`) and helm (`hasir-helm@`).

## ANTI-PATTERNS
- NEVER commit `gen/` output or generated buf configs.
- NEVER hand-edit generated files.
- NEVER merge breaking schema changes casually — CI enforces `buf breaking` vs main.

## NOTES
- Import paths use the double nesting: `proto/<domain>/v1/file.proto`.
- Older docs (README/ARCHITECTURE) claiming a root-level `proto/` dir are stale — schemas live here.
