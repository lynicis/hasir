variable "REGISTRY" {
  default = "ghcr.io/lynicis"
}
variable "TAG" {
  default = "latest"
}

group "default" {
  targets = ["api", "dashboard"]
}

target "api" {
  context = "."
  dockerfile = "apps/api/Dockerfile"
  tags = ["${REGISTRY}/hasir-api:${TAG}"]
  platforms = ["linux/amd64", "linux/arm64", "linux/arm/v7", "linux/386", "linux/ppc64le", "linux/s390x"]
}

target "dashboard" {
  context = "."
  dockerfile = "apps/dashboard/Dockerfile"
  tags = ["${REGISTRY}/hasir-dashboard:${TAG}"]
  # node:26-alpine and oven/bun:1.4-alpine only publish amd64 and arm64.
  platforms = ["linux/amd64", "linux/arm64"]
}
