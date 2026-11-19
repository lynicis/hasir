# Proto Package Instructions (`packages/proto`)

## Commands
| Task | Command |
|------|---------|
| Generate Go & TS code (`gen/`) | `make proto` (from repo root) |
| Lint schemas | `bunx buf lint` |
| Check breaking changes vs `main` | `bunx buf breaking --against '../../.git#branch=main,subdir=packages/proto'` |

## Code Map
| Path | Purpose |
|------|---------|
| `proto/{user,organization,registry,shared}/v1/*.proto` | Protobuf schema definitions (double-nested path) |
| `buf.yaml`, `buf.gen.yaml` | Buf workspace, lint/breaking rules, and codegen targets |
| `gen/{go,js}/` | Generated Go and TS SDKs (gitignored, ADR-0003) |

## Key Conventions
- Use double-nested import paths (`proto/<domain>/v1/<file>.proto`).
- Publish schema releases to the Buf Schema Registry via `proto/v*` tags (`proto-release.yml`), independent of app and Helm tags.
