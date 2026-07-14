// path: packages/config/eslint.config.mjs
// Shared ESLint flat-config preset used by every JS/TS workspace in the
// monorepo (the root config at /eslint.config.mjs re-exports this).
//
// The preset matches the conventions the dashboard repo already used
// (ESLint 9 + eslint-config-next + eslint-plugin-perfectionist with
// line-length descending sort). This file is the *single source of truth*
// — apps extend `@hasir/config/eslint` rather than each rolling their own.
//
// See ADR-0001 for why one preset beats per-app customization.

import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import perfectionist from "eslint-plugin-perfectionist";
import nextTs from "eslint-config-next/typescript";

export const hasirEslintPreset = defineConfig([
  ...nextVitals,
  ...nextTs,
  {
    plugins: {
      perfectionist,
    },
    rules: {
      "perfectionist/sort-imports": [
        "warn",
        {
          type: "line-length",
          order: "desc",
        },
      ],
      "perfectionist/sort-exports": [
        "warn",
        {
          type: "line-length",
          order: "desc",
        },
      ],
    },
  },
  globalIgnores([
    "**/.next/**",
    "**/.gitnexus/**",
    "**/.turbo/**",
    "**/out/**",
    "**/build/**",
    "**/coverage/**",
    "**/dist/**",
    "**/gen/**",
    "next-env.d.ts",
    "**/*.tsbuildinfo",
  ]),
]);

export default hasirEslintPreset;
