import { config as baseConfig } from "@hasir/eslint-config/base";
import { nextJsConfig } from "@hasir/eslint-config/next-js";

export default [
  ...baseConfig,
  {
    files: ["apps/dashboard/**/*.ts", "apps/dashboard/**/*.tsx", "apps/landing/**/*.ts", "apps/landing/**/*.tsx"],
    ...nextJsConfig[nextJsConfig.length - 2], // The next plugin config
  },
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
      "apps/api/**",
      ".worktrees/**"
    ]
  }
];
