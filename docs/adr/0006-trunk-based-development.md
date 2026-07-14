# 0006. Trunk-Based Development

## Status
Accepted

## Date
2026-07-14

## Decision
We will adopt Trunk-Based Development as our branching model. Developers will merge short-lived feature branches directly into the main branch. We will use squash merges to keep the commit history clean.

## Problem being solved
Long-lived branches (like in GitFlow) lead to merge conflicts, integration delays, and drift between environments. In a monorepo, these problems are amplified because multiple teams work in the same repository. We need a branching model that encourages frequent integration, fast feedback, and continuous delivery.

## Alternatives considered
* **GitFlow**: Using long-lived develop, release, and hotfix branches. This was rejected because it creates significant merge overhead, delays integration, and complicates the release process in a monorepo.
* **GitHub Flow**: A simpler model using feature branches merged directly into main, but without the strict emphasis on short-lived branches and automated testing before merge.

## Trade-offs
Trunk-Based Development requires high test coverage and reliable CI/CD pipelines to ensure that the main branch is always deployable. It also requires developers to merge small, incremental changes rather than large features.

## Benefits
* Minimizes merge conflicts and integration pain.
* Accelerates feedback loops for developers.
* Simplifies CI/CD pipelines by focusing on a single branch.
* Encourages small, incremental changes and continuous integration.

## Drawbacks
* Requires high discipline from developers to keep the main branch green.
* Requires comprehensive automated testing to catch regressions early.
* Can be difficult to manage large feature releases without feature flags.

## Why chosen
Trunk-Based Development is the most effective branching model for monorepos. It aligns with our goals of continuous integration, fast feedback, and independent service releases. It reduces the overhead of branch management and keeps the team focused on delivering value incrementally.
