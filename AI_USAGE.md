# AI Usage Guide

This project embraces AI-assisted development while maintaining code quality
standards. This guide explains how AI tooling is configured and how to use it
effectively.

## Tooling

Hasir uses **OpenCode** as its primary AI coding assistant, configured via
`opencode.json` at the root. It is backed by GitNexus (code intelligence),
PostgreSQL (database tooling), and Next.js (dev server) MCP servers.

AI instructions live in:

- **`AGENTS.md`** — Orchestrator-level instructions for AI agents working on
  the monorepo. Defines the project structure, skill routing, and behavioral
  guidelines.
- **`CLAUDE.md`** — Unified GitNexus skill configuration, providing each AI
  agent with codebase-awareness across the entire monorepo.

## Guidelines for AI-Assisted Contributions

### Do

- Use AI for boilerplate generation, refactoring, test writing, and
  documentation.
- Let AI assist with cross-app refactors where consistent changes are needed.
- Review and understand AI-generated code before committing — you own it.
- Use conventional commit messages, whether written by you or AI.

### Don't

- Don't commit AI-generated code you haven't reviewed and understood.
- Don't use AI to generate large blocks of untested code.
- Don't let AI suppress type errors (`as any`, `@ts-ignore`, etc.).
- Don't use AI to blindly migrate code between languages or frameworks
  without understanding the semantics.

## AI-Generated Code Quality

All AI-generated code must meet the same bar as hand-written code:

1. **Passes existing lint and type checks** — run `make lint` before pushing.
2. **Has tests** — the testing requirement does not change.
3. **Follows project conventions** — match the style and patterns of the
   surrounding code.
4. **No speculative abstraction** — don't add flexibility you don't need
   today.

## Reporting AI Issues

If the AI tooling produces incorrect, insecure, or non-functional code,
open an issue or reach out to **me@lynicis.dev**.

## Transparency

We believe AI is a force multiplier for human developers, not a replacement.
All contributors — human or AI-assisted — are expected to follow the same
[Code of Conduct](CODE_OF_CONDUCT.md) and
[Contributing Guidelines](CONTRIBUTING.md).
