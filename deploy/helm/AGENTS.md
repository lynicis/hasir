# AGENTS.md — helm-charts

Single Helm chart for deploying the **Hasir stack** (API + Dashboard + PostgreSQL) on Kubernetes.

## Quick commands

```bash
# Prerequisites
helm repo add bitnami https://charts.bitnami.com/bitnami

# Before any helm command
helm dependency update

# Install
helm install hasir . --namespace hasir --create-namespace

# Override values
helm install hasir . -f my-values.yaml
```

No Makefile, no test/lint runners, no typecheck — all `helm` CLI.

## Chart structure

- `Chart.yaml` — v2 Helm chart, depends on Bitnami PostgreSQL 15.5.3
- `values.yaml` — all config surface (domain, images, persistence, SMTP, secrets, TLS)
- `templates/` — standard resources: Deployments (api, dashboard), Services (api-http, api-ssh, dashboard), PVCs (ssh-keys, repos, sdk), ConfigMap, Secrets, Ingress
- ConfigMap drives all env vars (API connects to DB/dashboard via config)

## Architecture notes

- **API** container exposes HTTP (8080) and SSH (2222) ports. SSH gets a separate `LoadBalancer` service for git push.
- **Dashboard** (port 3000) is a static/Next.js app reading `NEXT_PUBLIC_API_URL` from ConfigMap.
- **Ingress** routes `/api` → API service, `/` → Dashboard service. TLS via cert-manager + letsencrypt.
- **Persistence**: 3 PVCs (ssh-keys 1Gi, repos 5Gi, sdk 5Gi) gated by `api.persistence.enabled`.

## CI

GitHub Actions on push to `main`: `helm/chart-releaser-action` publishes to GitHub Pages. No lint/test step.

## Secrets

`secrets.yaml` template reads from `values.yaml` and base64-encodes. In production, override via `--set` or a separate values file — never commit real secrets to `values.yaml`.

## GitNexus

This repo is indexed as **helm-charts** but consists only of YAML templates (no code). GitNexus impact/context tools are not useful here.

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **helm-charts** (54 symbols, 48 relationships, 0 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

> Index stale? Run `node .gitnexus/run.cjs analyze` from the project root — it auto-selects an available runner. No `.gitnexus/run.cjs` yet? `npx gitnexus analyze` (npm 11 crash → `npm i -g gitnexus`; #1939).

## Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows. For regression review, compare against the default branch: `detect_changes({scope: "compare", base_ref: "main"})`.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `query({query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `context({name: "symbolName"})`.

## Never Do

- NEVER edit a function, class, or method without first running `impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `rename` which understands the call graph.
- NEVER commit changes without running `detect_changes()` to check affected scope.

## Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/helm-charts/context` | Codebase overview, check index freshness |
| `gitnexus://repo/helm-charts/clusters` | All functional areas |
| `gitnexus://repo/helm-charts/processes` | All execution flows |
| `gitnexus://repo/helm-charts/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
|------|---------------------|
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->
