# The `@hasir/config` package owns the shared ESLint preset, the shared
# `tsconfig.base.json`, and any other tooling config that every workspace
# inherits. Defined in ADR-0001.

This package is intentionally *not* published. It is consumed only by
workspaces in this monorepo via the `workspace:` protocol.

## Consuming the ESLint preset

In any JS/TS workspace (`apps/dashboard/eslint.config.mjs`):

```mjs
import hasirPreset from "@hasir/config/eslint";

const eslintConfig = [
  ...hasirPreset,
  // workspace-local overrides go here
];

export default eslintConfig;
```

## Consuming the base tsconfig

In any JS/TS workspace's `tsconfig.json`:

```json
{
  "extends": "@hasir/config/tsconfig.base.json",
  "compilerOptions": {
    "paths": { "@/*": ["./*"] }
  }
}
```

## Why a separate package?

A root-only config file forces every workspace to re-declare the same rule
set. Splitting it into `@hasir/config` makes the preset addressable from
ESLint's `import` resolution (which the flat-config format requires) and
keeps the root `eslint.config.mjs` thin. See ADR-0001 for the full
discussion.
