# 0003. Local Codegen Never Committed

## Status
Accepted

## Date
2026-07-14

## Decision
We will generate protobuf code locally during development and CI/CD builds. We will never commit the generated code to the repository.

## Problem being solved
Committing generated code to the repository leads to merge conflicts, repository bloat, and potential drift between the protobuf definitions and the generated code. We need a clean, reproducible way to handle generated code that ensures consistency across all environments.

## Alternatives considered
* **Commit Generated Code**: Committing generated code to the repository. This was rejected because it causes frequent merge conflicts in generated files and bloats the repository history.
* **Generate During Release Only**: Generating code during the release phase only. This was rejected because it makes local development and testing difficult, as developers would not have access to the generated code locally without running a manual step.

## Trade-offs
Not committing generated code means developers must run a generation step locally, and CI/CD pipelines must run the generation step before building or testing. However, this ensures that the generated code is always in sync with the protobuf definitions.

## Benefits
* Eliminates merge conflicts in generated files.
* Keeps the repository clean and small.
* Ensures that the generated code is always in sync with the protobuf definitions.
* Guarantees that the build is reproducible from the source definitions.

## Drawbacks
* Developers must run a generation step locally.
* CI/CD pipelines must run the generation step, which slightly increases build times.
* Requires developers to have the buf toolchain installed locally.

## Why chosen
The benefits of avoiding merge conflicts and ensuring code consistency outweigh the minor overhead of running a generation step. Local generation ensures that the source of truth remains the protobuf definitions, not the generated code.
