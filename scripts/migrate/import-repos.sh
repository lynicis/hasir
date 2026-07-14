#!/usr/bin/env bash
set -euo pipefail

# Get the directory of the script and monorepo root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MONOREPO_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

echo "==> Starting monorepo migration import script..."
echo "NOTE: This script is for documentation and manual execution. DO NOT run it automatically."

# Define the repositories to import
# Format: name|git_url|target_subdirectory|tag_prefix
REPOS=(
  "api|git@github.com:protohasir/api.git|apps/api|api/"
  "dashboard|git@github.com:protohasir/dashboard.git|apps/dashboard|dashboard/"
  "proto|git@github.com:protohasir/proto.git|proto|proto/"
  "helm-charts|git@github.com:protohasir/helm-charts.git|deploy/helm|helm/"
  "docker-compose|git@github.com:protohasir/docker-compose.git|docker|docker/"
)

# Create tmp directory for mirrors
MIRRORS_DIR="${MONOREPO_DIR}/tmp/mirrors"
mkdir -p "${MIRRORS_DIR}"

# Ensure we are at the monorepo root
cd "${MONOREPO_DIR}"

# Initialize git if not already initialized
if [ ! -d ".git" ]; then
  echo "==> Initializing git repository..."
  git init
  git checkout -b main
  git commit --allow-empty -m "chore: initial commit"
fi

for repo_info in "${REPOS[@]}"; do
  IFS="|" read -r NAME URL SUBDIR TAG_PREFIX <<< "${repo_info}"

  echo "========================================="
  echo "Processing repository: ${NAME}"
  echo "========================================="

  REPO_PATH="${MIRRORS_DIR}/${NAME}"

  # 1. Clone the repository
  if [ -d "${REPO_PATH}" ]; then
    echo "==> Mirror directory ${REPO_PATH} already exists. Skipping clone."
  else
    echo "==> Cloning ${URL} to ${REPO_PATH}..."
    git clone "${URL}" "${REPO_PATH}"
  fi

  # Navigate to the cloned repo
  cd "${REPO_PATH}"

  # 2. Drop the 'hasir' ssh remote from proto repo if it exists
  if [ "${NAME}" = "proto" ]; then
    if git remote | grep -q "^hasir$"; then
      echo "==> Dropping 'hasir' remote from proto repository..."
      git remote remove hasir
    fi
  fi

  # 3. Rename default branch from master to main if it exists
  if git show-ref --verify --quiet refs/heads/master; then
    echo "==> Renaming branch master to main..."
    git branch -m master main
  fi

  # 4. Run git filter-repo to rewrite history and tags
  echo "==> Rewriting history with git filter-repo..."
  # We use --force to allow running on a fresh clone
  git filter-repo \
    --to-subdirectory-filter "${SUBDIR}" \
    --tag-rename "":"${TAG_PREFIX}" \
    --force

  # Go back to monorepo root
  cd "${MONOREPO_DIR}"

  # 5. Add the rewritten repo as a remote
  echo "==> Adding temporary remote for ${NAME}..."
  git remote add "temp-${NAME}" "${REPO_PATH}"

  # 6. Fetch the rewritten history and tags
  echo "==> Fetching rewritten history..."
  git fetch "temp-${NAME}" --tags

  # 7. Merge the history into the monorepo
  echo "==> Merging ${NAME} history into monorepo..."
  git merge "temp-${NAME}/main" \
    --allow-unrelated-histories \
    -m "chore: merge ${NAME} history into monorepo"

  # 8. Remove the temporary remote
  echo "==> Removing temporary remote..."
  git remote remove "temp-${NAME}"

  echo "==> Successfully imported ${NAME}!"
done

echo "==> All repositories imported successfully!"
