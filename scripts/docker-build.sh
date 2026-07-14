#!/usr/bin/env bash
set -euo pipefail

# Get the directory of the script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MONOREPO_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "==> Running docker buildx bake..."
cd "${MONOREPO_DIR}"
docker buildx bake -f docker/shared/docker-bake.hcl "$@"
