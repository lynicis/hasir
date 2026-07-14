# 0002. Flatten Proto Directory Structure

## Status
Accepted

## Date
2026-07-14

## Decision
We will flatten the protobuf directory structure. Instead of nesting the definitions inside proto/hasir-proto/, we will place them directly under proto/ at the root of the monorepo.

## Problem being solved
The original repository structure nested the protobuf definitions inside proto/hasir-proto/ (for example, proto/hasir-proto/proto/organization/v1/organization.proto). In a monorepo, this nesting is redundant and creates unnecessarily deep directory paths. We need a clean, flat structure that makes the protobuf definitions easy to find and work with.

## Alternatives considered
* **Keep Nested Structure**: Keeping the nested structure proto/hasir-proto/ to preserve the exact original repository layout and simplify the git history import. This was rejected because it introduces permanent, unnecessary complexity to the repository layout.
* **Move to Packages**: Moving the protobuf definitions into a shared package under packages/proto/. This was rejected because protobuf definitions are API contracts that define the interface between services, so they deserve a top-level directory rather than being treated as a shared library package.

## Trade-offs
Flattening the directory requires path rewriting during the git history import, which adds complexity to the migration script. However, it results in a much cleaner and more idiomatic repository layout.

## Benefits
* Eliminates redundant directory nesting.
* Simplifies import paths and configuration for the buf toolchain.
* Makes the repository structure more intuitive for developers.

## Drawbacks
* Requires path rewriting during the git history migration.
* May break existing references in git history if not handled carefully during the import.

## Why chosen
The long-term benefits of a clean, flat directory structure outweigh the one-time migration complexity of path rewriting. A flat proto/ directory is the standard pattern for monorepos and provides a better developer experience.
