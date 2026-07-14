# 0009. Git History Preservation and Import Strategy

## Status
Accepted

## Date
2026-07-14

## Decision
We will use git-filter-repo to rewrite the histories of the five standalone repositories and import them into the monorepo. We will rewrite paths to target subdirectories (for example, api to apps/api, dashboard to apps/dashboard) and merge them using git merge --allow-unrelated-histories.

## Problem being solved
We need to migrate five separate repositories into a single monorepo while preserving their git histories, tags, and commit metadata. We need a reliable, fast, and reproducible tool to perform this migration without losing valuable historical context.

## Alternatives considered
* **git subtree**: Using git subtree to import histories. This was rejected because it is slower, less flexible, and does not handle tag rewriting or complex path restructuring as cleanly as git-filter-repo.
* **git filter-branch**: Using the legacy git filter-branch tool. This was rejected because it is officially deprecated, extremely slow, and prone to corrupting history.
* **Merge Without Rewriting**: Merging histories directly using git merge --allow-unrelated-histories without path rewriting. This was rejected because it would merge all files into the root directory, causing conflicts and a messy repository structure.
* **Start Fresh**: Discarding git history and copying the latest files. This was rejected because preserving git history is a core requirement for tracking changes and maintaining blame context.

## Trade-offs
Using git-filter-repo requires installing an external Python-based tool and writing a custom migration script. However, it is extremely fast, safe, and provides powerful path and tag rewriting capabilities.

## Benefits
* Preserves full git history, commit metadata, and tags.
* Rewrites paths cleanly to avoid conflicts and match the target monorepo layout.
* Rewrites tags to be path-scoped (for example, v1.0.0 to api/v1.0.0) to avoid tag collisions.
* Extremely fast execution (seconds per repository).
* Reproducible migration process via scripting.

## Drawbacks
* Requires installing git-filter-repo (Python dependency).
* Requires writing and testing a migration script.
* Can be complex to handle edge cases like branch renaming and tag filtering.

## Why chosen
git-filter-repo is the modern, recommended tool for git history rewriting. It is faster and safer than git filter-branch and subtree. It provides the exact path and tag rewriting capabilities we need to merge five repositories into a clean monorepo structure while preserving historical context.
