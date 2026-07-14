# path: docker/shared/docker-bake.hcl
variable "REGISTRY" {
  default = "ghcr.io/protohasir"
}
variable "TAG" {
  default = "latest"
}

group "default" {
  targets = ["api", "dashboard"]
}

target "api" {
  context = "."
  dockerfile = "docker/api/Dockerfile"
  tags = ["${REGISTRY}/api:${TAG}"]
  platforms = ["linux/amd64", "linux/arm64"]
}

target "dashboard" {
  context = "."
  dockerfile = "docker/dashboard/Dockerfile"
  tags = ["${REGISTRY}/dashboard:${TAG}"]
  platforms = ["linux/amd64", "linux/arm64"]
}
