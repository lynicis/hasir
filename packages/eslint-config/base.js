import eslintConfigPrettier from "eslint-config-prettier"
import perfectionist from "eslint-plugin-perfectionist"
import onlyWarn from "eslint-plugin-only-warn"
import turboPlugin from "eslint-plugin-turbo"
import tseslint from "typescript-eslint"
import js from "@eslint/js"

/**
 * A shared ESLint configuration for the repository.
 *
 * @type {import("eslint").Linter.Config}
 * */
export const config = [
  js.configs.recommended,
  eslintConfigPrettier,
  ...tseslint.configs.recommended,
  {
    plugins: {
      perfectionist,
    },
    settings: {
      react: {
        version: "18.3.1",
      },
      next: {
        rootDir: ["apps/dashboard/", "apps/landing/"],
      },
    },
    rules: {
      "@next/next/no-html-link-for-pages": "off",
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
  {
    plugins: {
      turbo: turboPlugin,
    },
    rules: {
      "turbo/no-undeclared-env-vars": "warn",
    },
  },
  {
    plugins: {
      onlyWarn,
    },
  },
  {
    ignores: ["**/node_modules/**",
    "**/.next/**",
    "**/dist/**",
    "**/.turbo/**",
    "**/coverage/**",
    ".worktrees/**",
    "**/.worktrees/**"],
  },
]
