# Hasir Stack Helm Chart

<p align="center">
  <a href="https://github.com/protohasir/helm-charts">
    <img src="https://img.shields.io/badge/Helm-v3-blue.svg?logo=helm" alt="Helm v3">
  </a>
  <a href="https://kubernetes.io">
    <img src="https://img.shields.io/badge/Kubernetes-1.19+-blue.svg?logo=kubernetes" alt="Kubernetes 1.19+">
  </a>
  <a href="https://github.com/protohasir/helm-charts/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT">
  </a>
</p>

A Helm chart for deploying the complete **Hasir stack** (API, Dashboard, and PostgreSQL) on Kubernetes.

---

## Table of Contents

- [Background](#background)
- [Architecture Overview](#architecture-overview)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
  - [1. Add Bitnami Repository](#1-add-bitnami-repository)
  - [2. Install the Chart](#2-install-the-chart)
  - [3. Customizing Configuration](#3-customizing-configuration)
- [Configuration Reference](#configuration-reference)
- [Ingress and Domain Configuration](#ingress-and-domain-configuration)
- [Uninstalling](#uninstalling)
- [Contributing](#contributing)
- [License](#license)

---

## Background

Hasir is a developer platform designed to run lightweight Git/SSH-based workspaces, dashboards, and APIs. This repository provides a unified Helm chart to quickly deploy all required components on a Kubernetes cluster.

## Architecture Overview

The chart orchestrates three primary layers:
1. **API**: The main logic container exposing HTTP ports for API calls and SSH ports (`2222`) for workspace Git operations. Supports persistent volumes for repositories, SSH keys, and SDK code.
2. **Dashboard**: A Next.js-based web dashboard that reads API endpoints dynamically and displays dashboard UI to users.
3. **PostgreSQL**: The relational database subchart provided by Bitnami for storing user metadata, workspaces, and system configurations.

---

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- PV provisioner support in the underlying infrastructure (when persistence is enabled)
- (Optional) Ingress controller (e.g., `ingress-nginx`) and `cert-manager` for SSL/TLS certificates

---

## Installation

### 1. Add Bitnami Repository

This chart depends on the Bitnami PostgreSQL subchart. Add the repository to your Helm client:

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

### 2. Install the Chart

Make sure you update the chart dependencies first:

```bash
helm dependency update
```

Then, install the chart using:

```bash
helm install hasir . --namespace hasir --create-namespace
```

### 3. Customizing Configuration

You can override values during installation using a custom values file:

```bash
helm install hasir . -f my-values.yaml --namespace hasir
```

---

## Configuration Reference

The following table lists the configurable parameters of the Hasir chart and their default values.

### Global Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `global.domain` | Main domain name of the application | `hasir.local` |

### API Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `api.replicaCount` | Number of API service replicas | `1` |
| `api.image.repository` | API container image repository | `ghcr.io/protohasir/api` |
| `api.image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `api.image.tag` | API container image tag (defaults to chart's `appVersion`) | `""` |
| `api.service.type` | Service type for HTTP access | `ClusterIP` |
| `api.service.port` | API HTTP port | `8080` |
| `api.service.ssh.type` | Service type for SSH access | `LoadBalancer` |
| `api.service.ssh.port` | Port exposed for SSH git pushes | `2222` |
| `api.persistence.enabled` | Enable persistent volume storage for API data | `true` |
| `api.persistence.storageClass` | Storage class for PVs | `""` (use default) |
| `api.persistence.accessMode` | Volume access mode | `ReadWriteOnce` |
| `api.persistence.size.sshKeys` | Volume size for SSH keys | `1Gi` |
| `api.persistence.size.repos` | Volume size for repository stores | `5Gi` |
| `api.persistence.size.sdk` | Volume size for the SDK store | `5Gi` |
| `api.resources` | API Pod CPU/Memory resource limits/requests | `{}` |

### Dashboard Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `dashboard.replicaCount` | Number of dashboard service replicas | `1` |
| `dashboard.image.repository` | Dashboard image repository | `ghcr.io/protohasir/dashboard` |
| `dashboard.image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `dashboard.image.tag` | Dashboard container image tag (defaults to chart's `appVersion`) | `""` |
| `dashboard.service.type` | Service type for Dashboard HTTP access | `ClusterIP` |
| `dashboard.service.port` | Dashboard HTTP port | `3000` |
| `dashboard.resources` | Dashboard Pod CPU/Memory resource limits/requests | `{}` |

### Ingress Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable Ingress resource generation | `true` |
| `ingress.className` | Ingress class name | `"nginx"` |
| `ingress.annotations` | Annotations for configuring Ingress | *(See values.yaml)* |
| `ingress.hosts` | List of host rules for Ingress routing | *(See values.yaml)* |
| `ingress.tls` | TLS secrets/hosts setup | *(See values.yaml)* |

### Config & Secrets Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.smtp.host` | SMTP Host for outgoing mail | `"smtp.mailgun.org"` |
| `config.smtp.port` | SMTP Port | `587` |
| `config.smtp.from` | Outgoing sender email address | `"noreply@hasir.local"` |
| `config.smtp.useTls` | Use TLS for mail connection | `"true"` |
| `secrets.jwtSecret` | JWT secret token (Change this in production) | `"change-me-to-a-secure-key"` |
| `secrets.smtp.username` | SMTP username | `"postmaster@hasir.local"` |
| `secrets.smtp.password` | SMTP password | `"smtp-password"` |

### PostgreSQL Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `postgresql.enabled` | Enable PostgreSQL database subchart | `true` |
| `postgresql.auth.database` | Database name | `hasir` |
| `postgresql.auth.username` | Database username | `hasir` |
| `postgresql.auth.password` | Database password | `"hasir-db-password"` |
| `postgresql.persistence.enabled` | Enable database persistence | `true` |
| `postgresql.persistence.size` | Database volume size | `8Gi` |

---

## Ingress and Domain Configuration

By default, the Ingress resource routes traffic:
* `/api` to the `hasir-api` HTTP service.
* `/` to the `hasir-dashboard` service.

Set `global.domain` and `ingress.hosts[0].host` (plus the host inside `ingress.tls`) to your deployment's domain to configure routing correctly.

---

## Uninstalling

To uninstall/delete the `hasir` deployment:

```bash
helm uninstall hasir --namespace hasir
```

This removes all Kubernetes components associated with the chart, except for any persistent volume claims (PVCs) depending on your reclaim policy.

---

## Contributing

Contributions are welcome! Please open an issue or submit a pull request with any improvements.

## License

This project is licensed under the [MIT License](LICENSE).
