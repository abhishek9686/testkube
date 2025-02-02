// docker-bake.hcl
target "docker-metadata-action" {}

target "build" {
  inherits = ["docker-metadata-action"]
  context = "./"
  dockerfile = "build/api-server/Dockerfile"
  platforms = [
    "linux/amd64",
  ]
}