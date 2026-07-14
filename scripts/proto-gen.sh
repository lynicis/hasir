#!/usr/bin/env bash
set -euo pipefail

# Get the directory of the script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MONOREPO_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "==> Generating protobuf files..."
cd "${MONOREPO_DIR}"
bunx turbo run proto
