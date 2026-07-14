// path: eslint.config.mjs
// Root ESLint config. Re-exports the shared preset from packages/config so
// every workspace (apps/*, packages/*) inherits the same rules. The preset
// itself lives in packages/config/eslint.config.mjs — see ADR-0001.
import sharedConfig from "./packages/config/eslint.config.mjs";

const eslintConfig = [
  ...sharedConfig,
  {
    ignores: [
      "**/node_modules/**",
      "**/.next/**",
      "**/.turbo/**",
      "**/dist/**",
      "**/gen/**",
      "**/coverage/**",
      "**/*.tsbuildinfo",
      "**/next-env.d.ts",
      "apps/api/**"
    ]
  }
];

export default eslintConfig;
