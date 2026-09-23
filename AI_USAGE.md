# AI Usage Guide

This project embraces AI-assisted development while maintaining code quality
standards. This guide explains how AI tooling is configured and how to use it
effectively.

## Tooling

Hasir uses two AI coding assistants side-by-side:

| Tool | Config | Primary use |
|------|--------|-------------|
| **Antigravity (AGY)** | `.agents/` directory | Agentic tasks, OpenSpec workflow, skills |
| **OpenCode** | `opencode.json` | Interactive chat, inline edits, LSP-aware completions |

### MCP Servers

Both tools have access to MCP servers for live context:

| Server | Purpose |
|--------|---------|
| `postgresql` | Live schema inspection and query execution against the dev DB |
| `nextjs` | Next.js dev-tools (routing, RSC, component tree) |
| `shadcn` | Component registry lookups |

### AI Instruction Files

- **`AGENTS.md`** — Orchestrator-level instructions for AI agents: monorepo
  structure, skill routing, conventions, and anti-patterns.
- **`CONTEXT.md`** — Domain model and ubiquitous language for the platform.
- **`opencode.json`** — OpenCode configuration: instruction files, LSP, and
  MCP server definitions.
- **`docs/ARCHITECTURE.md`** and **`docs/RELEASE.md`** — Supplementary context
  loaded by OpenCode.

### Skills System

Agent skills live in `.agents/skills/` and are pinned via `skills-lock.json`.
They extend what AI agents know about project-specific patterns:

| Skill | Source |
|-------|--------|
| `building-components` | vercel/components.build |
| `bun-runtime` | oakoss/agent-skills |
| `changesets` | oakoss/agent-skills |
| `docker-patterns` | affaan-m/everything-claude-code |
| `github-actions` | oakoss/agent-skills |
| `golang-concurrency / database / testing / testify` | samber/cc-skills-golang |
| `helm-chart-scaffolding` | wshobson/agents |
| `nextjs-app-router-patterns / next-dev-loop / next-cache-* / next-partial-*` | project-local |
| `openspec-*` | project-local (OpenSpec workflow) |
| `postgresql-table-design / protobuf / shadcn / tailwind-design-system` | project-local |
| `turborepo / vercel-react-best-practices` | project-local |

### OpenSpec Workflow

Structured change management lives in `.agents/workflows/` and `openspec/`.
Use the `opsx` commands for any non-trivial feature:

```
opsx new      → create a change with design + spec + tasks
opsx ff       → fast-forward through artifact creation
opsx apply    → implement tasks one by one
opsx verify   → validate implementation against spec
opsx archive  → finalize and archive the change
```

## Guidelines for AI-Assisted Contributions

### Do

- Use AI for boilerplate generation, refactoring, test writing, and
  documentation.
- Let AI assist with cross-app refactors where consistent changes are needed.
- Review and understand AI-generated code before committing — you own it.
- Use conventional commit messages, whether written by you or AI.
- Use the OpenSpec workflow (`opsx`) for features that span multiple files.

### Don't

- Don't commit AI-generated code you haven't reviewed and understood.
- Don't use AI to generate large blocks of untested code.
- Don't let AI suppress type errors (`as any`, `@ts-ignore`, `@ts-expect-error`).
- Don't use AI to blindly migrate code between languages or frameworks
  without understanding the semantics.
- Don't bypass project anti-patterns (ORM migrations, plaintext secrets, etc.)
  even if the AI suggests it — see `AGENTS.md` for the full list.

## AI-Generated Code Quality

All AI-generated code must meet the same bar as hand-written code:

1. **Passes existing lint and type checks** — run `make lint` before pushing.
2. **Has tests** — the testing requirement does not change.
3. **Follows project conventions** — match the style and patterns of the
   surrounding code.
4. **No speculative abstraction** — don't add flexibility you don't need today.

## Reporting AI Issues

If the AI tooling produces incorrect, insecure, or non-functional code,
open an issue or reach out to **me@lynicis.dev**.

## Transparency

We believe AI is a force multiplier for human developers, not a replacement.
All contributors — human or AI-assisted — are expected to follow the same
[Code of Conduct](CODE_OF_CONDUCT.md) and
[Contributing Guidelines](CONTRIBUTING.md).
