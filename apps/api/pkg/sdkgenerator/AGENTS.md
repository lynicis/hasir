# SDK Generator Package

## OVERVIEW
This package manages automatic SDK generation from protobuf schemas using the buf CLI and protoc plugins.

## WHERE TO LOOK
| Component | File | Description |
|---|---|---|
| Core Interface | `generator.go` | Defines the Generator and CommandRunner interfaces. |
| Buf Generator | `buf_generator.go` | Implements SDK generation using the buf CLI. |
| Registry | `registry.go` | Registers and retrieves generators for different SDK types. |
| Import Resolver | `buf_resolver.go` | Parses imports and maps them to BSR modules. |
| Command Runner | `runner.go` | Executes external CLI commands like buf and protoc. |
| Go Generators | `go_*.go`, `buf_go_*.go` | Generators for Go SDKs (Protobuf, gRPC, ConnectRPC). |
| JS/TS Generators | `js_*.go`, `buf_js_*.go` | Generators for JavaScript/TypeScript SDKs. |
| Other Languages | `rust_*.go`, `java_*.go`, `csharp_*.go` | Generators for Rust, Java, and C# SDKs. |
| Documentation | `documentation.go` | Generates HTML/Markdown documentation from proto files. |
| Templates | `template/` | Mustache templates for generating documentation. |
| Tests | `*_test.go` | Unit and integration tests for generators and resolvers. |

## CONVENTIONS
- **Buf-First Generation**: Prefer using `buf` plugins over raw `protoc` commands for new SDK targets.
- **Mocking Command Execution**: Always use `MockCommandRunner` in tests to avoid executing real CLI commands.
- **Dynamic Configuration**: Generate `buf.yaml` and `buf.gen.yaml` dynamically in the repository path before running generation.
- **Strict Validation**: Validate all input paths and proto files before starting generation to prevent path traversal.
- **Test Coverage**: Write unit tests for any new generator implementation using mock runners to verify the generated arguments.

## ANTI-PATTERNS
- Hardcoding BSR module mappings is forbidden. Add them to `defaultBsrModuleMap` in `buf_resolver.go` or pass them as custom modules.
- Avoid executing commands directly using exec.Command. Always use the `CommandRunner` interface to allow mocking.
- Never commit generated buf.yaml or buf.gen.yaml files to the user's schema repository. These files must be generated dynamically and cleaned up if necessary.
- Don't bypass input validation. Every generator must call `Validate` before running.
