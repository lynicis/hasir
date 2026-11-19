# SDK Generator Instructions (`apps/api/pkg/sdkgenerator`)

## Commands
| Task | Command |
|------|---------|
| Test package | `cd apps/api && go test -count=1 ./pkg/sdkgenerator/...` |

## Code Map
| Component | File |
|-----------|------|
| `Generator` & `CommandRunner` interfaces | `generator.go`, `runner.go` |
| Registry & constructors | `registry.go` |
| Buf generator & BSR import resolver | `buf_generator.go`, `buf_resolver.go` |
| Language generators (Go, JS, Rust, Java, C#) | `{go,js,rust,java,csharp}_*.go`, `buf_{go,js}_*.go` |
| Documentation generator & template | `documentation.go`, `template/proto-doc-tempate.mustache` |

## Key Conventions
- Call `Validate` on all input paths and proto files before invoking a generator.
- Mock external CLI calls (`buf`, `protoc`, `git`, `tar`, `go`, `npm`, `npx`) with `MockCommandRunner` in tests.
- Register new BSR module mappings in `defaultBsrModuleMap` (`buf_resolver.go`); never hardcode mappings in generators.
- Generate `buf.yaml` and `buf.gen.yaml` dynamically in the target workspace and clean them up after generation.
- Keep the embedded template filename `template/proto-doc-tempate.mustache` (load-bearing spelling).
