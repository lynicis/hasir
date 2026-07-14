#!/usr/bin/env bash
set -euo pipefail

# Get the directory of the script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MONOREPO_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "==> Setting up hasir-monorepo dependencies..."

# 1. Bun install
echo "==> Running bun install..."
cd "${MONOREPO_DIR}"
bun install

# 2. Helm dependency update
echo "==> Updating Helm dependencies..."
if [ -d "${MONOREPO_DIR}/deploy/helm/hasir" ]; then
  helm dependency update "${MONOREPO_DIR}/deploy/helm/hasir"
else
  echo "Warning: Helm chart directory not found at deploy/helm/hasir"
fi

# 3. Go mod download
echo "==> Downloading Go modules..."
if [ -d "${MONOREPO_DIR}/apps/api" ]; then
  cd "${MONOREPO_DIR}/apps/api"
  go mod download
else
  echo "Warning: Go API directory not found at apps/api"
fi

echo "==> Setup complete!"
