# Hasir Helm Chart

A Helm chart for deploying the Hasir stack (API, Dashboard, and PostgreSQL) on Kubernetes.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- PV provisioner support in the underlying infrastructure
- (Optional) Ingress controller (e.g., `ingress-nginx`) and `cert-manager`

## Installation

### 1. Add Bitnami Repository (for PostgreSQL dependency)

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

### 2. Install the Chart

Install the chart using the following command:

```bash
helm install hasir ./helm-charts --namespace hasir --create-namespace
```

## Configuration

The following table lists the configurable parameters of the Hasir chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `global.domain` | Main domain name of the application | `hasir.local` |
| `api.replicaCount` | Number of api replicas | `1` |
| `api.image.repository` | API container image repository | `ghcr.io/protohasir/api` |
| `api.image.tag` | API container image tag | Chart's `appVersion` |
| `api.service.port` | API HTTP port | `8080` |
| `api.service.ssh.type` | Service type for SSH access | `LoadBalancer` |
| `api.service.ssh.port` | Port exposed for SSH git pushes | `2222` |
| `api.persistence.enabled` | Enable persistence storage for API | `true` |
| `dashboard.replicaCount` | Number of dashboard replicas | `1` |
| `dashboard.image.repository` | Dashboard image repository | `ghcr.io/protohasir/dashboard` |
| `ingress.enabled` | Enable Ingress resource | `true` |
| `ingress.className` | Ingress class name | `nginx` |
| `postgresql.enabled` | Enable PostgreSQL database subchart | `true` |
| `postgresql.auth.database` | Database name | `hasir` |
| `postgresql.auth.username` | Database username | `hasir` |
| `postgresql.auth.password` | Database password | `hasir-db-password` |

Override configuration values using a custom values file:

```bash
helm install hasir ./helm-charts -f my-values.yaml
```

## Ingress and Domain Configuration

By default, the ingress routes:
- `/api` to the `hasir-api` HTTP service.
- `/` to the `hasir-dashboard` service.

Set `global.domain` and `ingress.hosts[0].host` to your deployment's domain to configure routing correctly.
