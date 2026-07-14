<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This monorepo is indexed by GitNexus as **hasir-monorepo**. Use the GitNexus MCP tools to understand code across all apps, assess impact, and navigate safely.

> Index stale? Run `npx gitnexus analyze` from the repo root.

## Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows. For regression review, compare against the default branch: `detect_changes({scope: "compare", base_ref: "main"})`.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `query({search_query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `context({name: "symbolName"})`.
- For security review, `explain({target: "fileOrSymbol"})` lists taint findings (source→sink flows; needs `analyze --pdg`).

## Never Do

- NEVER edit a function, class, or method without first running `impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `rename` which understands the call graph.
- NEVER commit changes without running `detect_changes()` to check affected scope.

## Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/hasir-monorepo/context` | Codebase overview, check index freshness |
| `gitnexus://repo/hasir-monorepo/processes` | All execution flows |
| `gitnexus://repo/hasir-monorepo/process/{name}` | Step-by-step execution trace |

<!-- gitnexus:end -->
