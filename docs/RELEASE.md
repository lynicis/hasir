# Hasir Platform Monorepo Release Strategy

## 24. Release Strategy

This document defines the release strategy, versioning model, and branching strategy for the Hasir platform monorepo. The goal is to enable independent releases of services while maintaining a stable and predictable codebase.

### Versioning Model

We use independent per-application versioning managed by [Changesets](https://github.com/changesets/changesets). Each application or service in the monorepo is versioned independently using semantic versioning. Git tags are natively prefixed by changesets using the package name, such as `hasir-api@1.4.0` or `hasir-dashboard@2.1.0`.

#### Why Independent Versioning?
* **Decoupled Release Cycles**: Services can be deployed at different times without requiring a full platform release.
* **Reduced Risk**: A bug in one service does not block the release of other services.
* **Clear Traceability**: Git tags map directly to specific service versions in the container registry.

---

### Branching Model

We follow trunk-based development. This model is optimized for continuous integration and delivery.

#### Key Rules
* **Short-Lived Branches**: Developers work on feature branches that last no more than a few days.
* **Direct Merges to Main**: All feature branches are merged directly into the `main` branch via pull requests.
* **CI Verification**: Pull requests must pass all CI checks (linting, testing, building) before they can be merged.
* **No Long-Lived Release Branches**: We do not maintain separate staging or production branches. The `main` branch is always deployable.

---

### Release Process

The release process is fully automated using GitHub Actions and Changesets.

#### Step 1: Create a Changeset
When making changes, developers run the changeset CLI locally:

```bash
bunx changeset
```

This prompts you to select the packages that changed and the type of version bump (major, minor, patch). It generates a markdown file in the `.changeset` directory which must be committed alongside your code changes.

#### Step 2: Version Packages PR
When you merge your feature branch to `main`, the `.github/workflows/changeset.yml` workflow is triggered.
It consumes the `.changeset` files and automatically opens or updates a **"Version Packages"** Pull Request. This PR contains the calculated version bumps and aggregated `CHANGELOG.md` updates.

#### Step 3: Publish and Deploy
When a repository maintainer is ready to release, they merge the **"Version Packages"** PR into `main`.

1. The `changeset.yml` workflow triggers again, but this time it detects the version bumps.
2. It automatically creates Git tags (e.g., `hasir-api@1.5.0`) and publishes GitHub Releases.
3. The newly created tags trigger downstream workflows:
   - `.github/workflows/docker.yml`: Builds Docker images, tags them (e.g., `ghcr.io/lynicis/api:1.5.0`), and publishes them to GHCR.
   - `.github/workflows/helm-release.yml`: Packages the Helm chart and publishes it to the chart registry.

---

### Verification and Rollback

#### Verification
After a release, verify the deployment:

1. Check the application logs to ensure the service started successfully.
2. Verify that the health endpoint returns a successful status.
3. Run integration tests against the environment.

#### Rollback
If a release fails, roll back immediately:

1. Run `helm rollback <release> <revision>` to revert the Kubernetes resources to the previous stable state.
2. If necessary, redeploy the previous stable Docker image tag by updating the Helm values file.
