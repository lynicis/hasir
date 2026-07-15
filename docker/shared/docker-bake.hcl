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
  tags = ["${REGISTRY}/api:${TAG}"]
  platforms = ["linux/amd64", "linux/arm64"]
}

target "dashboard" {
  context = "."
  dockerfile = "apps/dashboard/Dockerfile"
  tags = ["${REGISTRY}/dashboard:${TAG}"]
  platforms = ["linux/amd64", "linux/arm64"]
}
