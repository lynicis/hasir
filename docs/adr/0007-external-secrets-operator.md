# 0007. External Secrets Operator for Helm Secrets

## Status
Accepted

## Date
2026-07-14

## Decision
We will use External Secrets Operator (ESO) integrated with AWS Secrets Manager (AWS SM) to manage secrets in our Helm charts. We will reference ExternalSecret resources in our Helm templates instead of committing plaintext secrets in values.yaml.

## Problem being solved
Committing plaintext secrets to a git repository is a major security risk. We need a secure, automated way to inject secrets into our Kubernetes deployments without committing them to git. The solution must support independent deployments and scale to multiple environments.

## Alternatives considered
* **Plaintext Secrets in values.yaml**: Committing secrets directly to git. This was rejected because it violates basic security practices.
* **Sealed Secrets**: Encrypting secrets with a cluster-side controller. This was rejected because it requires managing encryption keys locally and does not integrate directly with cloud-native secret managers like AWS Secrets Manager.
* **Helm Secrets (sops)**: Using sops to encrypt values files. This was rejected because it requires developers to have sops and decryption keys installed locally, which complicates the developer experience and local development.

## Trade-offs
Using ESO requires running the operator in the Kubernetes cluster and configuring IAM roles for service accounts (IRSA) to allow access to AWS Secrets Manager. However, it provides a secure and automated way to manage secrets.

## Benefits
* Secrets are stored securely in a dedicated secret manager (AWS Secrets Manager).
* No secrets are committed to the git repository.
* Automatic rotation of secrets is supported.
* Integrates seamlessly with Kubernetes.
* Simplifies multi-environment deployments by mapping external secrets to Kubernetes secrets dynamically.

## Drawbacks
* Requires running and maintaining the External Secrets Operator in the cluster.
* Requires cloud-specific configuration (IAM roles, policies).
* Local development requires a mock or alternative secret injection mechanism.

## Why chosen
ESO is the industry standard for integrating Kubernetes with cloud-native secret managers. It provides the best security posture and developer experience by separating secret management from application deployment. It allows us to manage secrets centrally in AWS Secrets Manager while deploying applications via Helm.
