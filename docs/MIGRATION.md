# Hasir Platform Monorepo Migration Guide

## 5. Migration Strategy

This section evaluates different approaches for migrating the five standalone repositories into the single monorepo. The goal is to consolidate the codebases while preserving history and minimizing disruption.

### Evaluation of Approaches

#### Git Filter-Repo (Recommended)
This is the modern standard for rewriting git history. It is fast, safe, and allows us to restructure the directory layout of each repository before merging. We can move all files into their target subdirectories (such as `apps/api/` or `apps/dashboard/`) while keeping the commit history intact.

#### Git Subtree
This tool can import repositories as subdirectories. It is simpler than rewriting history but often creates complex merge commits. It also makes it harder to clean up historical files that we do not want in the monorepo.

#### Git Merge with Unrelated Histories
This method merges histories directly. It does not restructure files first, which causes immediate conflicts if multiple repositories have files at the same paths (such as `README.md` or `.gitignore`).

#### Git Filter-Branch
This tool is deprecated and slow. It can corrupt history and should not be used for production migrations.

#### Git Submodules
This approach keeps the repositories separate. It defeats the primary objectives of a monorepo, such as unified build orchestration and simplified dependency management.

### Recommendation

We recommend using `git filter-repo` to restructure each source repository's history, followed by merging them into the monorepo using `git merge --allow-unrelated-histories`. This approach ensures a clean root directory from day one while preserving all historical commits and tags.

---

## 6. Git History Preservation Strategy

To preserve history and tags, we follow a multi-step process for each source repository.

1. **Create a Fresh Clone**: We clone the source repository into a temporary directory.
2. **Rewrite History**: We run `git filter-repo` to move all files into the designated subdirectory. For example, in the `api` repository, we move everything to `apps/api/`.
3. **Rename Tags**: We rename tags to prevent conflicts. For example, tag `v1.0.0` in the `api` repository becomes `api/v1.0.0`.
4. **Merge into Monorepo**: We add the rewritten local repository as a remote in the monorepo and merge it.

---

## 7. Exact Git Commands

Follow these steps to execute the migration. Ensure you have `git-filter-repo` installed on your system before starting.

### Step 1: Initialize the Monorepo

If you have not already initialized the monorepo, run these commands:

```bash
mkdir -p hasir-monorepo
cd hasir-monorepo
git init
touch .gitignore
git add .gitignore
git commit -m "Initial commit"
```

### Step 2: Migrate the API Repository

Run these commands to migrate the `api` repository:

```bash
# Clone the source repository to a temporary directory
git clone git@github.com:lynicis/api.git /tmp/hasir-migration/api
cd /tmp/hasir-migration/api

# Move all files into the target subdirectory apps/api/
git filter-repo --to-subdirectory-filter apps/api

# Rename all tags to include the application prefix
git tag | xargs -I {} git tag api/{} {}
git tag | grep -v "^api/" | xargs git tag -d

# Go back to the monorepo and merge the history
cd -
git remote add temp-api /tmp/hasir-migration/api
git fetch temp-api --tags
git merge temp-api/main --allow-unrelated-histories -m "Merge api repository history"
git remote remove temp-api
```

### Step 3: Migrate the Dashboard Repository

Run these commands to migrate the `dashboard` repository:

```bash
# Clone the source repository to a temporary directory
git clone git@github.com:lynicis/dashboard.git /tmp/hasir-migration/dashboard
cd /tmp/hasir-migration/dashboard

# Move all files into the target subdirectory apps/dashboard/
git filter-repo --to-subdirectory-filter apps/dashboard

# Rename all tags to include the application prefix
git tag | xargs -I {} git tag dashboard/{} {}
git tag | grep -v "^dashboard/" | xargs git tag -d

# Go back to the monorepo and merge the history
cd -
git remote add temp-dashboard /tmp/hasir-migration/dashboard
git fetch temp-dashboard --tags
git merge temp-dashboard/main --allow-unrelated-histories -m "Merge dashboard repository history"
git remote remove temp-dashboard
```

### Step 4: Migrate the Protobuf Repository

Run these commands to migrate the `hasir-proto` repository:

```bash
# Clone the source repository to a temporary directory
git clone git@github.com:lynicis/hasir-proto.git /tmp/hasir-migration/hasir-proto
cd /tmp/hasir-migration/hasir-proto

# Move all files into the target subdirectory proto/
git filter-repo --to-subdirectory-filter proto

# Rename all tags to include the prefix
git tag | xargs -I {} git tag proto/{} {}
git tag | grep -v "^proto/" | xargs git tag -d

# Go back to the monorepo and merge the history
cd -
git remote add temp-proto /tmp/hasir-migration/hasir-proto
git fetch temp-proto --tags
git merge temp-proto/main --allow-unrelated-histories -m "Merge proto repository history"
git remote remove temp-proto
```

### Step 5: Migrate the Helm Charts Repository

Run these commands to migrate the `helm-charts` repository:

```bash
# Clone the source repository to a temporary directory
git clone git@github.com:lynicis/helm-charts.git /tmp/hasir-migration/helm-charts
cd /tmp/hasir-migration/helm-charts

# Move all files into the target subdirectory deploy/helm/
git filter-repo --to-subdirectory-filter deploy/helm

# Rename all tags to include the prefix
git tag | xargs -I {} git tag helm/{} {}
git tag | grep -v "^helm/" | xargs git tag -d

# Go back to the monorepo and merge the history
cd -
git remote add temp-helm /tmp/hasir-migration/helm-charts
git fetch temp-helm --tags
git merge temp-helm/main --allow-unrelated-histories -m "Merge helm-charts repository history"
git remote remove temp-helm
```

### Step 6: Migrate the Docker Images Repository

Run these commands to migrate the `docker-images` repository:

```bash
# Clone the source repository to a temporary directory
git clone git@github.com:lynicis/docker-images.git /tmp/hasir-migration/docker-images
cd /tmp/hasir-migration/docker-images

# Move all files into the target subdirectory docker/
git filter-repo --to-subdirectory-filter docker

# Rename all tags to include the prefix
git tag | xargs -I {} git tag docker/{} {}
git tag | grep -v "^docker/" | xargs git tag -d

# Go back to the monorepo and merge the history
cd -
git remote add temp-docker /tmp/hasir-migration/docker-images
git fetch temp-docker --tags
git merge temp-docker/main --allow-unrelated-histories -m "Merge docker-images repository history"
git remote remove temp-docker
```

### Verification Steps

After merging all repositories, verify the migration:

1. Run `git log --oneline` to confirm that the commit history from all source repositories is present.
2. Run `git tag` to verify that all tags are present with their new prefixes.
3. Check the directory structure to ensure files are in their correct subdirectories.

### Rollback Procedure

If the migration fails at any point, you can reset the monorepo to its initial state:

```bash
git reset --hard origin/main
git clean -fd
```
This command removes all uncommitted changes and untracked files, restoring the repository to the last committed state.
