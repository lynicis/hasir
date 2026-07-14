# Hasir Platform Monorepo Release Strategy

## 24. Release Strategy

This document defines the release strategy, versioning model, and branching strategy for the Hasir platform monorepo. The goal is to enable independent releases of services while maintaining a stable and predictable codebase.

### Versioning Model

We use independent per-application versioning. Each application or service in the monorepo is versioned independently using semantic versioning. Git tags are prefixed with the application name, such as `api/v1.4.0` or `dashboard/v2.1.0`.

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

The release process is automated using GitHub Actions. It is triggered by pushing a versioned git tag.

#### Step 1: Version Bump
A developer runs the release script from the repository root:

```bash
# Bump the api service version by a minor version
bun run scripts/release.sh api minor
```

This script performs the following actions:
1. Finds the latest tag for the specified application.
2. Calculates the next version based on the bump type (major, minor, or patch).
3. Creates a new git tag (such as `api/v1.5.0`).
4. Pushes the tag to the remote repository.

#### Step 2: Build and Publish
The push of the tag triggers the `release.yml` workflow in GitHub Actions.

1. The workflow checks out the code at the tag.
2. It builds the Docker image for the specified application.
3. The image is tagged with the version (such as `ghcr.io/protohasir/api:1.5.0`) and `latest`.
4. The image is published to GHCR.

#### Step 3: Helm Chart Update
The release workflow updates the Helm chart values file with the new image tag.

1. The workflow modifies `deploy/helm/hasir/values.yaml` to update the image tag for the released service.
2. It packages the Helm chart and publishes it to the chart registry.
3. The updated chart is deployed to the target Kubernetes cluster.

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
