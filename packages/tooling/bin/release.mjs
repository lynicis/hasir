// path: packages/tooling/bin/release.mjs
// Node/Bun entry-point invoked by `bun run scripts/release.sh`. The shell
// script does the actual git work (tag pushing, release creation); this
// file exists so the monorepo has a single discoverable `hasir-release`
// CLI from `bunx @hasir/tooling` if someone wants to skip the shell path.

import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const scriptPath = fileURLToPath(new URL("../../../scripts/release.sh", import.meta.url));

const [, , app, bump = "patch"] = process.argv;

if (!app) {
  console.error("usage: hasir-release <app> [patch|minor|major]");
  process.exit(2);
}

const result = spawnSync("bash", [scriptPath, app, bump], { stdio: "inherit" });
process.exit(result.status ?? 1);
