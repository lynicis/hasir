# 0005. Path-Scoped Per-App SemVer Tags

## Status
Accepted

## Date
2026-07-14

## Decision
We will use independent per-app semantic versioning with path-scoped tags (for example, api/v1.4.0, dashboard/v2.1.0, proto/v0.5.0, helm/hasir/v0.2.0) to version and release services independently.

## Problem being solved
We need to release services independently. A single unified repository version would force us to release all services together or bump the version of services that have not changed. This violates our primary objective of independent service releases and creates unnecessary deployment risk.

## Alternatives considered
* **Single Unified Repository Version**: Bumping a single version for the entire repository on every release. This was rejected because it couples the release cycles of all services and makes it difficult to track which service changed in a given release.
* **Commit SHA Versioning**: Relying solely on git commit SHAs for versioning. This was rejected because it makes it difficult to track releases, roll back to specific versions, and manage Helm chart dependencies.

## Trade-offs
Path-scoped tags require more tooling to manage and automate releases, as we must detect which paths changed and bump the corresponding tags. However, it allows complete independence of service releases.

## Benefits
* Supports independent release cycles for each service.
* Clear mapping between git tags and deployed service versions.
* Integrates well with Helm chart versioning and Docker image tagging.
* Reduces deployment risk by isolating releases to changed services.

## Drawbacks
* More complex release automation and tooling.
* Developers must be aware of which paths they are modifying to trigger the correct release pipelines.
* Git history becomes populated with many tags.

## Why chosen
Independent release cycles are a core requirement for our platform. Path-scoped tags are the industry standard for achieving this in a monorepo while maintaining a single repository. This approach provides the necessary flexibility for our teams.
