// path: packages/tooling/bin/proto-gen.mjs
// Thin wrapper around `buf generate` so consumers don't need to remember
// the buffer path. The real entry-point is the `proto` task in the root
// `turbo.json` (one of ADR-0001's tenets: Turbo is the single orchestrator).
// This file exists so it can be invoked directly from a script in case turbo
// isn't available in the developer's shell.

import { spawnSync } from "node:child_process";
import { existsSync } from "node:fs";

const protoRoot = new URL("../../../proto/", import.meta.url);

if (!existsSync(new URL("buf.gen.yaml", protoRoot))) {
  console.error("proto/buf.gen.yaml not found — run from the monorepo root");
  process.exit(1);
}

const result = spawnSync("buf", ["generate", "--include-imports"], {
  cwd: protoRoot,
  stdio: "inherit",
});

process.exit(result.status ?? 1);
