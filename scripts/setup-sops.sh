#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
KEY_FILE="${ROOT_DIR}/keys.key"

echo "Setting up SOPS age keys..."

if [ ! -f "${KEY_FILE}" ]; then
    echo "Generating new age key at ${KEY_FILE}..."
    age-keygen -o "${KEY_FILE}"
else
    echo "Using existing age key at ${KEY_FILE}."
fi

# Extract public key
PUBLIC_KEY=$(grep -oE "age1[a-z0-9]+" "${KEY_FILE}")
echo "Public Key: ${PUBLIC_KEY}"

echo ""
echo "To configure your shell for local development, run:"
echo "export SOPS_AGE_KEY_FILE=\"${KEY_FILE}\""
echo ""
echo "Or if you want to use the key directly as a variable:"
echo "export SOPS_AGE_KEY=\"\$(grep -v '^#' \"${KEY_FILE}\" | head -n 1)\""
