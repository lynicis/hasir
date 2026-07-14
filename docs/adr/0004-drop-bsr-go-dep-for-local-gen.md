# 0004. Drop BSR Go Dependency for Local Gen

## Status
Accepted

## Date
2026-07-14

## Decision
We will drop the Buf Schema Registry (BSR) Go dependency (buf.build/gen/go/hasir/hasir) in hasir-api and switch to importing locally generated Go packages from the proto/gen/go directory.

## Problem being solved
The original hasir-api repository imported generated Go code from the BSR. This creates an external dependency on the BSR, requires a BSR token for builds, and introduces latency and potential failure points. In a monorepo, we have the protobuf definitions in the same repository, so we can generate the Go code locally and import it directly.

## Alternatives considered
* **Continue Using BSR Go Dependency**: Continuing to use the BSR Go dependency. This was rejected because it defeats one of the main benefits of a monorepo (local, self-contained builds) and introduces external dependencies.
* **Generate Directly Into API Directory**: Generating the Go code directly into the hasir-api directory. This was rejected because it couples the protobuf generation to a specific service, whereas the protobuf definitions are shared across multiple services.

## Trade-offs
Switching to local generation requires configuring the buf toolchain to generate code into a shared directory (proto/gen/go) and updating the Go module configuration to import from that directory. However, it eliminates the external dependency on the BSR.

## Benefits
* Eliminates external dependency on the BSR for Go builds.
* Simplifies local development by allowing developers to test protobuf changes immediately without publishing to the BSR.
* Reduces build times and failure points in CI/CD pipelines.
* Ensures that the Go service always uses the exact version of the protobuf definitions present in the repository.

## Drawbacks
* Requires configuring Go workspaces or module replacements to import the locally generated code.
* Requires developers to run the generation step locally before building the Go service.

## Why chosen
Local generation is the standard pattern for monorepos. It ensures that the Go service always uses the exact version of the protobuf definitions present in the repository, without relying on an external registry. This improves build reliability and developer velocity.
